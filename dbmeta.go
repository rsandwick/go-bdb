package bdb

const (
	_dbFileIdLen = 20 // Unique file ID length.
	_dbIVBytes   = 16 // Bytes per IV
	_dbMACKey    = 20 // Bytes per MAC checksum
)

type _dbmeta33header struct {
	LSN         dbLSN       // 00-07: LSN.
	Pgno        uint32      // 08-11: Current page number.
	Magic       uint32      // 12-15: Magic number.
	Version     uint32      // 16-19: Version.
	PageSize    uint32      // 20-23: Page size.
	EncryptAlg  uint8       // 24: Encryption algorithm.
	Type        uint8       // 25: Page type.
	MetaFlags   dbmetaFlags // 26: Meta-only flags
	Unused1     uint8       // 27: Unused.
	Free        uint32      // 28-31: Free list page number.
	LastPgno    dbPgno      // 32-35: Page number of last page in db.
	NParts      uint32      // 36-39: Number of partitions.
	KeyCount    uint32      // 40-43: Cached key count.
	RecordCount uint32      // 44-47: Cached record count.
	Flags       uint32      // 48-51: Flags: unique to each AM.

	Uid [_dbFileIdLen]uint8 // 52-71: Unique file ID.
}

type _dbmeta33trailer struct {
	CryptoMagic uint32            // 460-463: Crypto magic number
	Trash       [3]uint32         // 464-475: Trash space - Do not use
	IV          [_dbIVBytes]uint8 // 476-495: Crypto IV
	Chksum      [_dbMACKey]uint8  // 496-511: Page chksum
}

type dbmetaFlags = uint8

/*
// DBMETA_*
const (
	dbmetaChksum dbmetaFlags = 1 << iota
	dbmetaPartRange
	dbmetaPartCallback
)
*/
