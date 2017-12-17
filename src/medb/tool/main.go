package main

import (
	"errors"
	"flag"
	"fmt"
	"medb/storage"

	"os"
	"os/exec"

	"time"

	"github.com/google/uuid"
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
	err = commitToGIT(rootPath, fmt.Sprintf("MeDB Sync - %v", time.Now().Unix()))
	if err != nil {
		panic(err)
	}

	// Step 5: Rebase on new changes?
}

func commitToGIT(rootPath string, message string) error {
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(curDir)

	err = os.Chdir(rootPath)
	if err != nil {
		fmt.Println("1")
		return err
	}
	cmd := exec.Command("git", "add", "-A")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	// See if there's anything to commit
	cmd = exec.Command("git", "diff-index", "--quiet", "HEAD")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		// Nothing to commit, return!
		fmt.Println("Nothing to commit.")
		return nil
	}

	// Now commit everything
	cmd = exec.Command("git", "commit", "-am", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()

}
