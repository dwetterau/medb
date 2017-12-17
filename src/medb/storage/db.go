package storage

import (
	"io/ioutil"
	"path"
)

var blacklistedFolderNames = map[string]struct{}{
	".git": {},
}

type DB interface {
	ContentByPath() (map[string]string, error)
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

func (d dbImpl) ContentByPath() (map[string]string, error) {
	filenames, err := d.scanForFilenames()
	if err != nil {
		return nil, err
	}

	type filenameAndData struct {
		filename string
		data     string
		err      error
	}
	numWorkers := 10
	resultChan := make(chan filenameAndData)
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
						resultChan <- filenameAndData{err: fileErr}
						continue
					}
					// TODO: Make real files instead (parsing out headers if they're there)
					resultChan <- filenameAndData{filename: filename, data: string(bytes)}
				}
			}
		}()
	}
	go func() {
		for _, filename := range filenames {
			workChan <- filename
		}
	}()

	contentByPath := make(map[string]string, len(filenames))
	for len(contentByPath) < len(filenames) {
		result := <-resultChan
		if result.err != nil {
			return nil, err
		}
		contentByPath[result.filename] = result.data
	}
	return contentByPath, nil
}
