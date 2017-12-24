package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"medb/storage"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func main() {
	var staticDir string
	var rootPath string
	flag.StringVar(&staticDir, "static", staticDir, "path to the static dir for the ui")
	flag.StringVar(&rootPath, "root", rootPath, "path to the root of the db instance")
	flag.Parse()

	if staticDir == "" {
		panic("Must specify static dir!")
	}
	if rootPath == "" {
		panic("Must specify root path!")
	}

	staticServer := http.FileServer(http.Dir(staticDir))

	http.Handle("/static/", staticServer)
	http.HandleFunc("/", rootHandler(staticDir))

	// API v1
	http.HandleFunc("/api/1/list", listHandler(rootPath))
	http.HandleFunc("/api/1/sync", syncHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func rootHandler(staticDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fullPath := staticDir + "/" + r.URL.Path[1:]
		http.ServeFile(w, r, fullPath)
	}
}

func listHandler(rootPath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := storage.NewDB(rootPath)
		files, err := db.AllFiles()
		if err != nil {
			panic(err)
		}

		filesAsJSON := convertToJSON(rootPath, files)

		raw, err := json.Marshal(filesAsJSON)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(w, string(raw))
	}
}

type jsonFile struct {
	Name     string      `json:"name""`
	State    string      `json:"state"`
	Contents []*jsonFile `json:"contents"`
	Id       uuid.UUID   `json:"id"`
}

func convertToJSON(rootPath string, files []storage.File) []*jsonFile {
	// Sort shortest paths first to keep things stable
	sort.Slice(files, func(i, j int) bool {
		p1 := files[i].Path()
		p2 := files[j].Path()
		if len(p1) != len(p2) {
			return len(p1) < len(p2)
		}
		return strings.Compare(p1, p2) < 0
	})
	tree := &jsonFile{}

	for _, file := range files {
		cur := tree
		path := file.Path()[len(rootPath):]
		pathStrings := strings.Split(path, "/")
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
					&jsonFile{Name: component},
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
	toProcess := []*jsonFile{tree}
	for len(toProcess) > 0 {
		cur := toProcess[len(toProcess)-1]
		toProcess = toProcess[:len(toProcess)-1]

		if len(cur.Contents) > 0 {
			cur.State = "expanded"
		} else {
			cur.State = "collapsed"
		}
		for _, c := range cur.Contents {
			toProcess = append(toProcess, c)
		}
	}

	return tree.Contents
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Print the synced status
}
