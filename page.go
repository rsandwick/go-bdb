package bdb

// A pageType identifies a page's purpose, such as BTree Leaf vs  Hash pages.
type pageType uint8

//go:generate stringer -trimprefix pageType -type pageType
const (
	pageTypeInvalid      pageType = iota // Invalid page type.
	pageTypeDuplicate                    // Duplicate. DEPRECATED in 3.1
	pageTypeHashUnsorted                 // Hash pages created pre 4.6. DEPRECATED
	pageTypeIBTree                       // Btree internal.
	pageTypeIRecNo                       // Recno internal.
	pageTypeLBTree                       // Btree leaf.
	pageTypeLRecNo                       // Recno leaf.
	pageTypeOverflow                     // Overflow.
	pageTypeHashMeta                     // Hash metadata page.
	pageTypeBTreeMeta                    // Btree metadata page.
	pageTypeQAMMeta                      // Queue metadata page.
	pageTypeQAMData                      // Queue data page.
	pageTypeLDup                         // Off-page duplicate leaf.
	pageTypeHash                         // Sorted hash page.
	pageTypeHeapMeta                     // Heap metadata page.
	pageTypeHeap                         // Heap data page.
	pageTypeIHeap                        // Heap internal.
	pageTypeMax
)

// isLeaf reports whether t is a leaf page type.
func (t pageType) isLeaf() bool {
	switch t {
	case pageTypeLBTree, pageTypeLRecNo, pageTypeLDup:
		return true
	}
	return false
}

type pageHeader struct {
	LSN      dbLSN    // 00-07: Log sequence number.
	Pgno     dbPgno   // 08-11: Current page number.
	PrevPgno dbPgno   // 12-15: Previous page number.
	NextPgno dbPgno   // 16-19: Next page number.
	Entries  dbIndx   // 20-21: Number of items on the page.
	HFOffset dbIndx   // 22-23: High free byte page offset.
	Level    uint8    // 24: Btree tree level.
	Type     pageType // 25: Page type.
}

type page struct {
	pageHeader
	entryOffsets []uint16
}

// IsLeaf reports whether p is a leaf page.
func (p page) IsLeaf() bool {
	return p.Type.isLeaf()
}

// EntryOffset returns the byte offset of entry n in this page.
func (p *page) EntryOffset(n int) int64 {
	return int64(p.entryOffsets[n])
}

type btreeEntryType = uint8

/*
const (
	_         btreeEntryType = iota
	KeyData                  // Key/data item.
	Duplicate                // Duplicate key/data item.
	Overflow                 // Overflow key/data item.
)
*/

type btreeEntry struct {
	Len  dbIndx         // 00-01: Key/data item length.
	Type btreeEntryType // 02: Page type AND DELETE FLAG.
}

// bKeyData represents a key/value pair in memory or on disk.
type bKeyData struct {
	btreeEntry
	Data []byte // 03-n: Variable length key/data item.
}

// bInternal represents an internal btree reference page.
type bInternal struct {
	binternalHeader
	Data []byte // 12-n: Variable length key item.
}

type binternalHeader struct {
	btreeEntry
	bInternalData
}

type bInternalData struct {
	Unused uint8  // 03: Padding, unused.
	Pgno   dbPgno // 04-07: Page number of referenced page.
	NRecs  uint32 // 08-11: Subtree record count.
}
