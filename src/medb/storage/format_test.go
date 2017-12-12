package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateHeader(t *testing.T) {
	testUUID, err := uuid.Parse("430bf597-74ac-40ad-9453-edcc353bc026")
	if err != nil {
		t.Fail()
	}
	f := fileImpl{
		header: headerImpl{
			version:    currentVersion,
			id:         testUUID,
			creationTS: time.Unix(1513066695, 0),
			modifiedTS: time.Unix(1513066711, 0),
			tags:       []string{"example", "first_tag", "oh_oops"},
		},
		content: "content\n",
	}
	expectedHeader := `--BEGIN HEADER--
Version: 1
ID: 430bf597-74ac-40ad-9453-edcc353bc026
MD5: f75b8179e4bbe7e2b4a074dcef62de95
CreationTS: 1513066695
ModifiedTS: 1513066711
Tags: example,first_tag,oh_oops
--END HEADER--
`
	if f.generateHeader() != expectedHeader {
		t.Fail()
	}
}
