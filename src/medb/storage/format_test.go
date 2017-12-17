package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testHeader = `--BEGIN HEADER--
Version: 1
ID: 430bf597-74ac-40ad-9453-edcc353bc026
MD5: f75b8179e4bbe7e2b4a074dcef62de95
CreationTS: 1513066695
ModifiedTS: 1513066711
--END HEADER--`

func TestGenerateHeader(t *testing.T) {
	testUUID, err := uuid.Parse("430bf597-74ac-40ad-9453-edcc353bc026")
	if err != nil {
		t.Fatal(err)
	}
	f := fileImpl{
		header: headerImpl{
			version:    currentVersion,
			id:         testUUID,
			creationTS: time.Unix(1513066695, 0),
			modifiedTS: time.Unix(1513066711, 0),
		},
		content: "content\n",
	}
	if f.generateHeader() != testHeader {
		t.Fail()
	}
}

func TestParseFile(t *testing.T) {
	testUUID, err := uuid.Parse("430bf597-74ac-40ad-9453-edcc353bc026")
	if err != nil {
		t.Fatal(err)
	}

	file, err := parseFile(testHeader + "\ncontent\n")
	if err != nil {
		t.Fatal(err)
	}
	expectedFile := fileImpl{
		header: headerImpl{
			version:    currentVersion,
			id:         testUUID,
			creationTS: time.Unix(1513066695, 0),
			modifiedTS: time.Unix(1513066711, 0),
		},
		content: "content\n",
	}
	if file != expectedFile {
		t.Fatal(expectedFile, file)
	}

	// Check a hash mismatch case
	_, err = parseFile(testHeader + "\nnot_the_real_content\n")
	if err.Error() != "malformed header, hash doesn't match content's hash" {
		t.Fail()
	}
}
