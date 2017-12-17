package storage

import (
	"time"

	"fmt"

	"strings"

	"errors"

	"regexp"

	"strconv"

	"path"

	"github.com/google/uuid"
)

/*

This file defines the storage format of the files stored in the DB.
Below is the only canonical example for version 1

--BEGIN HEADER--
Version: 1
ID: 430bf597-74ac-40ad-9453-edcc353bc026
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
	HasHeader() bool
	CreateHeader() error
	UpdateHeader()
	ID() uuid.UUID
	Name() string
}

type fileImpl struct {
	header          headerImpl
	content         string
	currentLocation string
}

var _ File = &fileImpl{}

type headerImpl struct {
	version    uint32
	id         uuid.UUID
	creationTS time.Time
	modifiedTS time.Time
}

func (f *fileImpl) HasHeader() bool {
	return f.header != headerImpl{}
}

func (f *fileImpl) CreateHeader() error {
	id, err := uuid.NewUUID()
	if err != nil {
		// This would only happen if we're out of randomness...
		return err
	}

	now := time.Now()
	f.header = headerImpl{
		version:    currentVersion,
		id:         id,
		creationTS: now,
		modifiedTS: now,
	}
	return nil
}

func (f *fileImpl) UpdateHeader() {
	f.header.modifiedTS = time.Now()
}

func (f *fileImpl) ID() uuid.UUID {
	return f.header.id
}

func (f *fileImpl) Name() string {
	_, name := path.Split(f.currentLocation)
	return name
}

func (f *fileImpl) generateHeader() string {
	return strings.Join([]string{
		headerStart,
		fmt.Sprintf("Version: %d", f.header.version),
		fmt.Sprintf("ID: %s", f.header.id),
		fmt.Sprintf("CreationTS: %d", f.header.creationTS.Unix()),
		fmt.Sprintf("ModifiedTS: %d", f.header.modifiedTS.Unix()),
		headerEnd,
	}, "\n") + "\n"
}

func parseFile(input string) (*fileImpl, error) {
	if input[:len(headerStart)] != headerStart {
		// Assume there's no header
		return &fileImpl{
			content: input,
		}, nil
	}

	// Try to parse the header
	splitByNewlines := strings.Split(input, "\n")
	if len(splitByNewlines) < 7 {
		return nil, errors.New("malformed header, not enough lines")
	}
	if splitByNewlines[0] != headerStart {
		return nil, errors.New("malformed header, header start is incorrect")
	}

	// Parse the version
	r := regexp.MustCompile("^Version: ([0-9]+)$")
	submatches := r.FindStringSubmatch(splitByNewlines[1])
	if len(submatches) != 2 {
		return nil, errors.New("malformed header, version is incorrect")
	}
	version, err := strconv.ParseInt(submatches[1], 10, 32)
	if err != nil {
		return nil, err
	}

	// Parse the id
	r = regexp.MustCompile("^ID: (\\S+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[2])
	if len(submatches) != 2 {
		return nil, errors.New("malformed header, id is incorrect")
	}
	id, err := uuid.Parse(submatches[1])
	if err != nil {
		return nil, errors.New("malformed header, couldn't parse id")
	}

	// Parse the creation time
	r = regexp.MustCompile("^CreationTS: ([0-9]+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[3])
	if len(submatches) != 2 {
		return nil, errors.New("malformed header, creationTS is incorrect")
	}
	ts, err := strconv.ParseInt(submatches[1], 10, 64)
	if err != nil {
		return nil, errors.New("malformed header, creationTS is incorrect")
	}
	creationTS := time.Unix(ts, 0)

	// Parse the modified time
	r = regexp.MustCompile("^ModifiedTS: ([0-9]+)$")
	submatches = r.FindStringSubmatch(splitByNewlines[4])
	if len(submatches) != 2 {
		return nil, errors.New("malformed header, modifiedTS is incorrect")
	}
	ts, err = strconv.ParseInt(submatches[1], 10, 64)
	if err != nil {
		return nil, errors.New("malformed header, modifiedTS is incorrect")
	}
	modifiedTS := time.Unix(ts, 0)

	// Lastly, check the header ending
	if splitByNewlines[5] != headerEnd {
		return nil, errors.New("malformed header, header end is incorrect")
	}
	actualContent := strings.Join(splitByNewlines[6:], "\n")

	return &fileImpl{
		header: headerImpl{
			version:    uint32(version),
			id:         id,
			creationTS: creationTS,
			modifiedTS: modifiedTS,
		},
		content: actualContent,
	}, nil
}
