package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/google/uuid"

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

	// TODO: Check that folders have a file
	// Step 1: Make sure all files have a header
	fileIDsToSave := make(map[uuid.UUID]struct{}, 0)
	for _, file := range files {
		if !file.HasHeader() {
			fmt.Printf("INFO: Creating new header for filename: %s\n", file.Name())
			file.CreateHeader()
			fileIDsToSave[file.ID()] = struct{}{}
		}
	}
	// Step 2: Make sure there are no duplicate ids.
	idToFileMap := make(map[uuid.UUID]storage.File, len(files))
	for _, file := range files {
		if _, ok := idToFileMap[file.ID()]; ok {
			panic(errors.New(fmt.Sprintf("id: %s appears twice in the DB", file.ID())))
		}
		idToFileMap[file.ID()] = file
	}
	// Step 3: Save all the files we added headers for to disk
	for fileID := range fileIDsToSave {
		file := idToFileMap[fileID]
		fmt.Printf("INFO: Saving %s with new header.\n", file.ID())
		err = db.SaveFile(file)
		if err != nil {
			panic(err)
		}
	}
	// Step 4: Create a new git commit with all the changes + all new files
	err = db.CommitToGIT(fmt.Sprintf("MeDB Sync - %v", time.Now().Unix()))
	if err != nil {
		panic(err)
	}

	// Print some stuff for fun
	lastCommitTS, err := db.LastCommitTS()
	if err != nil {
		panic(err)
	}
	lastPullTS, err := db.LastPullTS()
	if err != nil {
		panic(err)
	}
	aheadBehind, err := db.AheadBehindOriginMaster()
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"Last committed %v ago, last pulled %v ago. Origin is ahead by %d and we are ahead by %d\n",
		time.Since(lastCommitTS),
		time.Since(lastPullTS),
		aheadBehind.OriginAheadBy,
		aheadBehind.LocalAheadBy,
	)

	// Step 5: Rebase on new changes?
	// Step 6: Push out changes
	if aheadBehind.LocalAheadBy > 0 {
		err = db.Push()
		if err != nil {
			panic(err)
		}
	}
}
