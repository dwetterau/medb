package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testHeader = "--BEGIN HEADER--\n" +
	"Version: 1\n" +
	"ID: 430bf597-74ac-40ad-9453-edcc353bc026\n" +
	"CreationTS: 1513066695\n" +
	"--END HEADER--"

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
		},
	}
	generatedHeader := f.generateHeader()
	if generatedHeader != testHeader+"\n" {
		t.Fatal(generatedHeader, testHeader)
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
	expectedFile := &fileImpl{
		header: headerImpl{
			version:    currentVersion,
			id:         testUUID,
			creationTS: time.Unix(1513066695, 0),
		},
		content: "content\n",
	}
	if *file != *expectedFile {
		t.Fatal(expectedFile, file)
	}

	// Check a malformed header case
	_, err = parseFile(testHeader + "trailing on last line" + "content\n")
	if err.Error() != "malformed header, header end is incorrect" {
		t.Fail()
	}
}
