package main

import (
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
	contentByPath, err := db.ContentByPath()
	if err != nil {
		panic(err)
	}

	for filename, data := range contentByPath {
		fmt.Printf("%s: %s\n", filename, data)
	}
}
