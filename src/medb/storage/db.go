package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
)

var blacklistedFolderNames = map[string]struct{}{
	".git":  {},
	".medb": {},
}
var blacklistedFileNames = map[string]struct{}{
	".gitignore": {},
}

type DB interface {
	AllFiles() ([]File, error)
	AsJSON() ([]*JSONFile, error)
	SaveFile(File) error
	NewFile(path string, content string) error
	CommitToGIT(message string) (bool, error)
	Push() error
	Pull() error
}

func NewDB(rootPath string) DB {
	return dbImpl{rootPath: rootPath}
}

type JSONFile struct {
	Name     string      `json:"name""`
	State    string      `json:"state"`
	Contents []*JSONFile `json:"contents"`
	Id       uuid.UUID   `json:"id"`
}

type dbImpl struct {
	rootPath string
}

func (d dbImpl) scanForFilenames() ([]string, error) {
	seenDirectories := make(map[string]struct{}, 1)
	directories := []string{d.rootPath}
	files := make(map[string]struct{}, 0)
	for len(directories) > 0 {
		d := directories[len(directories)-1]
		directories = directories[:len(directories)-1]

		infos, err := ioutil.ReadDir(d)
		if err != nil {
			return nil, err
		}
		for _, info := range infos {
			p := path.Join(d, info.Name())
			if info.IsDir() {
				if _, ok := blacklistedFolderNames[info.Name()]; ok {
					continue
				}

				if _, ok := seenDirectories[p]; !ok {
					seenDirectories[p] = struct{}{}
					directories = append(directories, p)
				}
			} else {
				if _, ok := blacklistedFileNames[info.Name()]; ok {
					continue
				}
				files[p] = struct{}{}
			}
		}
	}

	fileSlice := make([]string, 0, len(files))
	for file := range files {
		fileSlice = append(fileSlice, file)
	}
	return fileSlice, nil
}

func (d dbImpl) AllFiles() ([]File, error) {
	filenames, err := d.scanForFilenames()
	if err != nil {
		return nil, err
	}

	type fileOrError struct {
		file fileImpl
		err  error
	}
	numWorkers := 10
	resultChan := make(chan fileOrError)
	workChan := make(chan string)
	stopChan := make(chan struct{})
	defer close(stopChan)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				select {
				case <-stopChan:
					return
				case filename := <-workChan:
					bytes, fileErr := ioutil.ReadFile(filename)
					if fileErr != nil {
						resultChan <- fileOrError{err: fileErr}
						continue
					}
					file, err := parseFile(string(bytes))
					if err != nil {
						resultChan <- fileOrError{err: fileErr}
						continue
					}
					file.currentLocation = filename
					resultChan <- fileOrError{
						file: *file,
					}
				}
			}
		}()
	}
	go func() {
		for _, filename := range filenames {
			workChan <- filename
		}
	}()

	allFiles := make([]File, 0, len(filenames))
	for len(allFiles) < len(filenames) {
		result := <-resultChan
		if result.err != nil {
			return nil, err
		}
		allFiles = append(allFiles, &result.file)
	}
	return allFiles, nil
}

func (d dbImpl) AsJSON() ([]*JSONFile, error) {
	files, err := d.AllFiles()
	if err != nil {
		return nil, err
	}

	// Sort shortest paths first to keep things stable
	sort.Slice(files, func(i, j int) bool {
		p1 := files[i].Path()
		p2 := files[j].Path()
		if len(p1) != len(p2) {
			return len(p1) < len(p2)
		}
		return strings.Compare(p1, p2) < 0
	})
	tree := &JSONFile{}

	for _, file := range files {
		cur := tree
		p := file.Path()[len(d.rootPath):]
		pathStrings := strings.Split(p, "/")
		for componentIndex, component := range pathStrings {
			if component == "" {
				continue
			}
			foundIndex := -1
			for i, val := range cur.Contents {
				if val.Name == component {
					foundIndex = i
					break
				}
			}
			if foundIndex == -1 {
				cur.Contents = append(
					cur.Contents,
					&JSONFile{Name: component, State: "collapsed"},
				)
				foundIndex = len(cur.Contents) - 1
			}
			if componentIndex == len(pathStrings)-1 {
				cur.Contents[foundIndex].Id = file.ID()
			}
			// Move down the tree
			cur = cur.Contents[foundIndex]
		}
	}
	toProcess := []*JSONFile{tree}
	for len(toProcess) > 0 {
		cur := toProcess[len(toProcess)-1]
		toProcess = toProcess[:len(toProcess)-1]

		for _, c := range cur.Contents {
			toProcess = append(toProcess, c)
		}
	}

	return tree.Contents, nil
}

func (d dbImpl) SaveFile(fileToSave File) error {
	f, ok := fileToSave.(*fileImpl)
	if !ok {
		return errors.New("don't know how to save this type of file")
	}

	return ioutil.WriteFile(f.currentLocation, []byte(f.generateHeader()+f.content), 0644)
}

func (d dbImpl) NewFile(desiredPath string, content string) error {
	fileToSave := &fileImpl{
		content:         content,
		currentLocation: path.Join(d.rootPath, desiredPath),
	}
	err := fileToSave.CreateHeader()
	if err != nil {
		return err
	}
	return d.SaveFile(fileToSave)
}

// Returns true if a new commit was made, false otherwise
func (d dbImpl) CommitToGIT(message string) (bool, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return false, err
	}
	defer os.Chdir(curDir)

	err = os.Chdir(d.rootPath)
	if err != nil {
		return false, err
	}
	cmd := exec.Command("git", "add", "-A")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return false, err
	}

	// See if there's anything to commit
	cmd = exec.Command("git", "diff-index", "--quiet", "HEAD")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		// Nothing to commit, return!
		return false, nil
	}

	// Now commit everything
	cmd = exec.Command("git", "commit", "-am", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return true, cmd.Run()
}

func (d dbImpl) Push() error {
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(curDir)

	err = os.Chdir(d.rootPath)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d dbImpl) Pull() error {
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(curDir)

	err = os.Chdir(d.rootPath)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
