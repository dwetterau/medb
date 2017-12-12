package storage

import (
	"crypto/md5"
	"time"

	"fmt"

	"strings"

	"github.com/google/uuid"
)

/*

This file defines the storage format of the files stored in the DB.
Below is the only canonical example for version 1

--BEGIN HEADER--
Version: 1
ID: 430bf597-74ac-40ad-9453-edcc353bc026
MD5: f75b8179e4bbe7e2b4a074dcef62de95
CreationTS: 1513066695
ModifiedTS: 1513066711
Tags: example,first_tag,oh_oops
--END HEADER--
content\n
EOF
*/

const currentVersion = 1

type File interface {
	HeaderAndContent() string
}

type fileImpl struct {
	header  headerImpl
	content string
}

var _ File = fileImpl{}

type headerImpl struct {
	version    uint32
	id         uuid.UUID
	creationTS time.Time
	modifiedTS time.Time
	tags       []string
}

func (f fileImpl) HeaderAndContent() string {
	return f.generateHeader() + f.content
}

func (f fileImpl) generateHeader() string {
	md5String := fmt.Sprintf("MD5: %x", md5.Sum([]byte(f.content)))
	return strings.Join([]string{
		"--BEGIN HEADER--",
		fmt.Sprintf("Version: %d", f.header.version),
		fmt.Sprintf("ID: %s", f.header.id),
		md5String,
		fmt.Sprintf("CreationTS: %d", f.header.creationTS.Unix()),
		fmt.Sprintf("ModifiedTS: %d", f.header.modifiedTS.Unix()),
		fmt.Sprintf("Tags: %s", strings.Join(f.header.tags, ",")),
		"--END HEADER--",
	}, "\n") + "\n"
}
