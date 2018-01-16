package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"time"

	"bytes"

	"strconv"

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
	Search(query string, options SearchOptions) ([]File, error)
	SaveFile(File) error
	LoadFile(fileID uuid.UUID) (File, error)
	NewFile(path string, content string) error

	// TODO: Move to a git interface?
	CommitToGIT(message string) error
	Push() error
	Pull() error
	Fetch() error
	LastCommitTS() (time.Time, error)
	LastPullTS() (time.Time, error)
	AheadBehindOriginMaster() (AheadBehindStruct, error)
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

type SearchOptions struct {
	Limit int
}

type AheadBehindStruct struct {
	OriginAheadBy int64
	LocalAheadBy  int64
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
				cur.Contents[foundIndex].State = "file"
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

func (d dbImpl) Search(query string, options SearchOptions) ([]File, error) {
	files, err := d.AllFiles()
	if err != nil {
		return nil, err
	}

	type rankInfo struct {
		matchType    int
		creationTime time.Time
		index        int
	}

	results := make([]rankInfo, 0)
	for i, file := range files {
		f := file.(*fileImpl)
		lQuery := strings.ToLower(query)
		if strings.Contains(strings.ToLower(f.Name()), lQuery) {
			results = append(results, rankInfo{
				1,
				f.header.creationTS,
				i,
			})
			continue
		}

		// TODO: Don't search binary files!

		if strings.Contains(strings.ToLower(f.Content()), lQuery) {
			results = append(results, rankInfo{
				2,
				f.header.creationTS,
				i,
			})
		}
	}

	// Sort the results
	sort.Slice(results, func(i, j int) bool {
		if results[i].matchType == results[j].matchType {
			// Sort by creation time (newest first)
			return results[i].creationTime.After(results[j].creationTime)
		}
		return results[i].matchType < results[j].matchType
	})

	// Return the top limit results
	filesToReturn := make([]File, 0, options.Limit)
	for i := 0; i < options.Limit && i < len(results); i++ {
		filesToReturn = append(filesToReturn, files[results[i].index])
	}
	return filesToReturn, nil
}

func (d dbImpl) SaveFile(fileToSave File) error {
	f, ok := fileToSave.(*fileImpl)
	if !ok {
		return errors.New("don't know how to save this type of file")
	}

	return ioutil.WriteFile(f.currentLocation, []byte(f.generateHeader()+f.content), 0644)
}

func (d dbImpl) LoadFile(fileID uuid.UUID) (File, error) {
	allFiles, err := d.AllFiles()
	if err != nil {
		return nil, err
	}

	for _, f := range allFiles {
		if f.ID() == fileID {
			return f, nil
		}
	}
	return nil, errors.New("file doesn't exist")
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

// Changes the working directory to the rootPath and returns a defer to move
// it back to the original.
func (d dbImpl) moveCurDir() (func(), error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	toDefer := func() { os.Chdir(curDir) }

	err = os.Chdir(d.rootPath)
	if err != nil {
		return nil, err
	}
	return toDefer, nil
}

// Runs the command and returns a string of the output
func (d dbImpl) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	buf.ReadFrom(stdout)

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Returns true if a new commit was made, false otherwise
func (d dbImpl) CommitToGIT(message string) error {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return err
	}
	defer toDefer()

	// Add everything in the directory to be committed
	_, err = d.runCommand("git", "add", "-A")
	if err != nil {
		return err
	}

	// See if there's anything to commit
	_, err = d.runCommand("git", "diff-index", "--quiet", "HEAD")
	if err == nil {
		// Nothing to commit, return!
		return nil
	}

	// Now commit everything
	_, err = d.runCommand("git", "commit", "-am", message)
	return err
}

func (d dbImpl) Push() error {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return err
	}
	defer toDefer()

	_, err = d.runCommand("git", "push")
	return err
}

func (d dbImpl) Fetch() error {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return err
	}
	defer toDefer()

	_, err = d.runCommand("git", "fetch")
	return err
}

func (d dbImpl) Pull() error {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return err
	}
	defer toDefer()

	_, err = d.runCommand("git", "push")
	return err
}

func (d dbImpl) LastCommitTS() (time.Time, error) {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return time.Time{}, err
	}
	defer toDefer()

	lastCommitTSRaw, err := d.runCommand("git", "log", "-1", "--format=%cd", "--date=unix")
	if err != nil {
		return time.Time{}, err
	}
	lastCommitTS, err := strconv.ParseInt(strings.TrimSpace(lastCommitTSRaw), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(lastCommitTS, 0), nil
}

func (d dbImpl) LastPullTS() (time.Time, error) {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return time.Time{}, err
	}
	defer toDefer()

	lastPullTSRaw, err := d.runCommand("stat", "-c", "%Y", ".git/FETCH_HEAD")
	if err != nil {
		return time.Time{}, err
	}
	lastPullTS, err := strconv.ParseInt(strings.TrimSpace(lastPullTSRaw), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(lastPullTS, 0), nil
}

func (d dbImpl) AheadBehindOriginMaster() (AheadBehindStruct, error) {
	toDefer, err := d.moveCurDir()
	if err != nil {
		return AheadBehindStruct{}, err
	}
	defer toDefer()

	leftAndRightRaw, err := d.runCommand("git", "rev-list", "--left-right", "--count", "origin/master...master")
	if err != nil {
		return AheadBehindStruct{}, err
	}
	leftAndRightStrings := strings.Fields(leftAndRightRaw)
	if len(leftAndRightStrings) != 2 {
		return AheadBehindStruct{}, errors.New("unable to parse --left-right output")
	}
	originAheadBy, err := strconv.ParseInt(leftAndRightStrings[0], 10, 64)
	if err != nil {
		return AheadBehindStruct{}, err
	}
	localAheadBy, err := strconv.ParseInt(leftAndRightStrings[1], 10, 64)
	if err != nil {
		return AheadBehindStruct{}, err
	}
	return AheadBehindStruct{originAheadBy, localAheadBy}, nil
}
