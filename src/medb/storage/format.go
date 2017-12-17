package storage

import (
	"crypto/md5"
	"time"

	"fmt"

	"strings"

	"errors"

	"regexp"

	"strconv"

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
--END HEADER--
content\n
EOF
*/

const (
	currentVersion = 1
	headerStart    = "--BEGIN HEADER--"
	headerEnd      = "--END HEADER--"
)

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
}

func (f fileImpl) HeaderAndContent() string {
	return f.generateHeader() + f.content
}

func (f fileImpl) generateHeader() string {
	md5String := fmt.Sprintf("MD5: %x", md5.Sum([]byte(f.content)))
	return strings.Join([]string{
		headerStart,
		fmt.Sprintf("Version: %d", f.header.version),
		fmt.Sprintf("ID: %s", f.header.id),
		md5String,
		fmt.Sprintf("CreationTS: %d", f.header.creationTS.Unix()),
		fmt.Sprintf("ModifiedTS: %d", f.header.modifiedTS.Unix()),
		headerEnd,
	}, "\n") + "\n"
}

func parseFile(input string) (fileImpl, error) {
	if input[:len(headerStart)] != headerStart {
		// Assume there's no header
		return fileImpl{
			content: input,
		}, nil
	}

	// Try to parse the header
	splitByNewlines := strings.Split(input, "\n")
	if len(splitByNewlines) < 8 {
		return fileImpl{}, errors.New("malformed header, not enough lines")
	}
	if splitByNewlines[0] != headerStart {
		return fileImpl{}, errors.New("malformed header, header start is incorrect")
	}

	// Parse the version
	r := regexp.MustCompile("^Version: ([0-9]+)$")
	submatches := r.FindStringSubmatch(splitByNewlines[1])
	if len(submatches) != 2 {
		return fileImpl{}, errors.New("malformed header, version is incorrect")
	}
	version, err := strconv.ParseInt(submatches[1], 10, 32)
	if err != nil {
		return fileImpl{}, err
	}

	// Parse the id
	r = regexp.MustCompile("^ID: (\\S+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[2])
	if len(submatches) != 2 {
		return fileImpl{}, errors.New("malformed header, id is incorrect")
	}
	id, err := uuid.Parse(submatches[1])
	if err != nil {
		return fileImpl{}, errors.New("malformed header, couldn't parse id")
	}

	// Parse the hash
	actualContent := strings.Join(splitByNewlines[7:], "\n")
	actualMD5 := fmt.Sprintf("%x", md5.Sum([]byte(actualContent)))
	r = regexp.MustCompile("^MD5: (\\S+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[3])
	if len(submatches) != 2 {
		return fileImpl{}, errors.New("malformed header, md5 is incorrect")
	}
	if submatches[1] != actualMD5 {
		return fileImpl{}, errors.New("malformed header, hash doesn't match content's hash")
	}

	// Parse the creation time
	r = regexp.MustCompile("^CreationTS: ([0-9]+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[4])
	if len(submatches) != 2 {
		return fileImpl{}, errors.New("malformed header, creationTS is incorrect")
	}
	ts, err := strconv.ParseInt(submatches[1], 10, 64)
	if err != nil {
		return fileImpl{}, errors.New("malformed header, creationTS is incorrect")
	}
	creationTS := time.Unix(ts, 0)

	// Parse the modified time
	r = regexp.MustCompile("^ModifiedTS: ([0-9]+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[5])
	if len(submatches) != 2 {
		return fileImpl{}, errors.New("malformed header, modifiedTS is incorrect")
	}
	ts, err = strconv.ParseInt(submatches[1], 10, 64)
	if err != nil {
		return fileImpl{}, errors.New("malformed header, modifiedTS is incorrect")
	}
	modifiedTS := time.Unix(ts, 0)

	// Lastly, check the header ending
	if splitByNewlines[6] != headerEnd {
		return fileImpl{}, errors.New("malformed header, header end is incorrect")
	}

	return fileImpl{
		header: headerImpl{
			version:    uint32(version),
			id:         id,
			creationTS: creationTS,
			modifiedTS: modifiedTS,
		},
		content: actualContent,
	}, nil
}
