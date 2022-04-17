package bdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrBadMagic = errors.New("bad magic number for access method")
	NotFound    = errors.New("key not found")
)

/*
// An accessMethod identifies a database access method.
type accessMethod = uint8

const (
	_           accessMethod = iota
	TypeBTree                // DB_BTREE
	TypeHash                 // DB_HASH
	TypeRecNo                // DB_RECNO
	TypeQueue                // DB_QUEUE
	TypeUnknown              // DB_UNKNOWN
	TypeHeap                 // DB_HEAP
)
*/

type dbLSN struct {
	File   uint32 // File ID.
	Offset uint32 // File offset.
}

type dbPgno = uint32 // db_pgno_t
type dbIndx = uint16 // db_indx_t

// DB is the key-value API for a Berkeley DB using any access method.
type DB interface {
	// Get retrieves the value associated with the supplied key. If the key
	// does not exist and no other error occurs, the error will be NotFound.
	Get(key string) ([]byte, error)

	// HasKey reports whether the DB contains the supplied key.
	HasKey(key string) (bool, error)

	// Keys returns a slice of all the keys in the DB.  Berkeley DB does not
	// impose limits on key or value size other than the 4GiB implied by the
	// 32-bit length; be careful.
	Keys() ([][]byte, error)
}

// NewReader opens a Berkeley DB, detecting its access method from its
// header. Only one database per file is currently directly supported,
// but a bytes.Reader or io.SectionReader may help emulate multiple.
func NewReader(r io.ReadSeeker) (DB, error) {
	buf := make([]byte, 16)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}
	magic := binary.LittleEndian.Uint32(buf[12:16])
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	switch magic {
	case btreeMagic:
		return newBTreeReader(r)
	}
	return nil, fmt.Errorf("unknown magic: 0x%06x", magic)
}

// Shorten little-endian reads a bit...
func leRead(r io.Reader, x interface{}) error {
	return binary.Read(r, binary.LittleEndian, x)
}
