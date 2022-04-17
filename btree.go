package bdb

import (
	"errors"
	"fmt"
	"io"
)

const btreeMagic = 0x053162

type btree struct {
	btreeMetadata
	r     io.ReadSeeker
	pages []*page
}

// btreeMetadata describes a database which uses the BTree access method.
type btreeMetadata struct {
	_dbmeta33header
	_btmeta33
	_dbmeta33trailer
}

type _btmeta33 struct {
	Unused1 uint32     // 72-75: Unused space.
	MinKey  uint32     // 76-79: Btree: Minkey.
	ReLen   uint32     // 80-83: Recno: fixed-length record length.
	RePad   uint32     // 84-87: Recno: fixed length record pad.
	Root    uint32     // 88-91: Root page.
	Unused2 [92]uint32 // 92-459: Unused space.
}

// newBTreeReader opens a Berkeley DB with the BTree access method.
func newBTreeReader(r io.ReadSeeker) (DB, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	db := btree{r: r}
	if err := leRead(r, &db.btreeMetadata); err != nil {
		return nil, err
	}
	if db.Magic != btreeMagic {
		return nil, fmt.Errorf(
			"wanted magic 0x%06x, got 0x%06x", btreeMagic, db.Magic,
		)
	}
	return &db, nil
}

// Get implements the DB interface.
func (t *btree) Get(k string) ([]byte, error) {
	n, i, err := t.search(k)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		return nil, NotFound
	}
	p, err := t.getPage(n)
	if err != nil {
		return nil, err
	}
	e, err := t.readBKeyData(n, p.entryOffsets[i+1])
	return e.Data, err
}

// HasKey implements the DB interface.
func (t *btree) HasKey(k string) (bool, error) {
	_, _, err := t.search(k)
	if errors.Is(err, NotFound) {
		return false, nil
	}
	return (err == nil), err
}

// Keys implements the DB interface.
func (t *btree) Keys() ([][]byte, error) {
	return nil, NotFound
}

// Search traverses the btree for key k and reports its page+entry.
// If an error occurs, the faulty page and entry number are reported,
// when possible. An entry of -1 means no entry number applies; in the
// non-error case, this means no entry was found to match the key.
func (t *btree) search(k string) (dbPgno, int, error) {
	p, err := t.getPage(t.Root)
	if err != nil {
		return t.Root, 0, err
	}
	for {
		switch p.Type {
		case pageTypeIBTree:
			e, err := t.readBInternal(p.Pgno, p.entryOffsets[0])
			if err != nil {
				return p.Pgno, 0, err
			}
			for i := 1; i < int(p.Entries); i++ {
				next, err := t.readBInternal(p.Pgno, p.entryOffsets[i])
				if err != nil {
					return p.Pgno, i, err
				}
				if k < string(next.Data) {
					break
				}
				e = next
			}
			p, err = t.getPage(e.Pgno)
			if err != nil {
				return p.Pgno, -1, err
			}
			// TODO(raynard): add visited check to avoid infinite loop
		case pageTypeLBTree:
			for i := 0; i < int(p.Entries); i += 2 {
				e, err := t.readBKeyData(p.Pgno, p.entryOffsets[i])
				if err != nil {
					return p.Pgno, i, err
				}
				if k == string(e.Data) {
					return p.Pgno, i, nil
				}
			}
			return 0, -1, NotFound
		default:
			return p.Pgno, -1, fmt.Errorf("unexpected page type %s", p.Type)
		}
	}
}

func (t *btree) getPage(n dbPgno) (*page, error) {
	if n < 0 || n > t.LastPgno {
		return nil, fmt.Errorf("page index out of range")
	}
	if len(t.pages) == 0 {
		t.pages = make([]*page, t.LastPgno+1)
	}
	if t.pages[n] == nil {
		sz := t.PageSize
		offs := int64(n * sz)
		if _, err := t.r.Seek(offs, io.SeekStart); err != nil {
			return nil, err
		}
		page := page{}
		err := leRead(t.r, &page.pageHeader)
		if err != nil {
			return nil, err
		}
		// TODO(raynard): check page type actually should have entryOffsets?
		page.entryOffsets = make([]uint16, page.Entries)
		if err := leRead(t.r, &page.entryOffsets); err != nil {
			return nil, err
		}
		t.pages[n] = &page
	}
	return t.pages[n], nil
}

func (t *btree) readBInternal(n dbPgno, entryOffset uint16) (*bInternal, error) {
	offs := int64(n*t.PageSize) + int64(entryOffset)
	if _, err := t.r.Seek(offs, io.SeekStart); err != nil {
		return nil, err
	}
	var e bInternal
	if err := leRead(t.r, &e.binternalHeader); err != nil {
		return nil, err
	}
	e.Data = make([]byte, e.Len)
	if _, err := t.r.Read(e.Data); err != nil {
		return nil, err
	}
	return &e, nil
}

func (t *btree) readBKeyData(n dbPgno, entryOffset uint16) (*bKeyData, error) {
	offs := int64(n*t.PageSize) + int64(entryOffset)
	if _, err := t.r.Seek(offs, io.SeekStart); err != nil {
		return nil, err
	}
	var e bKeyData
	if err := leRead(t.r, &e.btreeEntry); err != nil {
		return nil, err
	}
	e.Data = make([]byte, e.Len)
	if _, err := t.r.Read(e.Data); err != nil {
		return nil, err
	}
	return &e, nil
}
