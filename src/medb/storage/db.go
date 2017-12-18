package storage

import (
	"errors"
	"io/ioutil"
	"path"
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
	SaveFile(File) error
}

func NewDB(rootPath string) DB {
	return dbImpl{rootPath: rootPath}
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

func (d dbImpl) SaveFile(fileToSave File) error {
	f, ok := fileToSave.(*fileImpl)
	if !ok {
		return errors.New("don't know how to save this type of file")
	}

	return ioutil.WriteFile(f.currentLocation, []byte(f.generateHeader()+f.content), 0644)
}
