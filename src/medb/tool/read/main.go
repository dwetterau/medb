package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"medb/storage"
)

func main() {
	var rootPath string

	flag.StringVar(&rootPath, "root", rootPath, "path to the root of the db instance")
	flag.Parse()

	if rootPath == "" {
		panic("Must specify root path!")
	}

	db := storage.NewDB(rootPath)
	files, err := db.AllFiles()
	if err != nil {
		panic(err)
	}

	filesAsJSON := convertToJSON(files)

	raw, err := json.Marshal(filesAsJSON)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(raw))
}

type jsonFile struct {
	Name     string     `json:"name""`
	State    string     `json:"state"`
	Contents []jsonFile `json:"contents"`
}

func convertToJSON(files []storage.File) []jsonFile {
	// TODO: Actually use contents
	rootFiles := make([]jsonFile, 0)
	for _, file := range files {
		rootFiles = append(rootFiles, jsonFile{
			Name:     file.Name(),
			State:    "collapsed",
			Contents: []jsonFile{},
		})
	}
	return rootFiles
}
