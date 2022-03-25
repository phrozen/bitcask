// Package bitcask ...
package bitcask

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path"
	"sync"
	"time"
)

const (
	LOCKFILE  = ".lock"
	TOMBSTONE = "bitcask_tombstone"
)

// Opts
type Opts struct {
	ReadOnly     bool
	SyncOnPut    bool
	MaxKeySize   int
	MaxFileSize  int
	MaxValueSize int
}

// Default options
var Default = Opts{}

// Bitcask ...
type Bitcask struct {
	sync.RWMutex
	cursor  int
	current *os.File
	keydir  map[string]lookup
	dirname string
}

type lookup struct {
	file      *os.File
	size      int
	position  int64
	timestamp time.Time
}

// Open reads the files at directory and parses them or creates
// a new database. Directory has to exist and be writeable by
// the current process.
func Open(dirname string, opts Opts) (*Bitcask, error) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		return nil, err
	}

	filename := time.Now().Format("20060102130405.cask")
	f, err := os.OpenFile(path.Join(dirname, filename), os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	bc := &Bitcask{}
	bc.current = f
	bc.keydir = make(map[string]lookup)
	bc.dirname = dirname

	err = bc.lockfile()
	if err != nil {
		return nil, err
	}

	return bc, nil
}

// Get retrieves the value for a given key fromt he store.
// If a key does not exist, both value and error will be nil.
func (bc *Bitcask) Get(key []byte) ([]byte, error) {
	value, exist := bc.keydir[string(key)]
	if !exist {
		return nil, nil
	}

	bc.RLock()
	defer bc.RUnlock()

	data := make([]byte, value.size)
	_, err := value.file.ReadAt(data, value.position)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Put saves the new key/value pair and syncs the store
func (bc *Bitcask) Put(key, value []byte) error {
	now := time.Now()
	data := bc.block(key, value, now)

	bc.Lock()
	defer bc.Unlock()

	n, err := bc.current.Write(data)
	if err != nil {
		return err
	}

	bc.keydir[string(key)] = lookup{
		file:      bc.current,
		size:      len(value),
		position:  int64(bc.cursor + 16 + len(key)),
		timestamp: now,
	}

	bc.cursor += n
	return nil
}

// Delete removes the key / value from the store
func (bc *Bitcask) Delete(key []byte) error {
	data := bc.block(key, []byte(TOMBSTONE), time.Now())

	bc.Lock()
	defer bc.Unlock()

	n, err := bc.current.Write(data)
	if err != nil {
		return err
	}

	delete(bc.keydir, string(key))
	bc.cursor += n
	return nil
}

// List returns all the keys in the store as a slice.
func (bc *Bitcask) List() []string {
	bc.RLock()
	defer bc.RUnlock()

	keys := make([]string, 0, len(bc.keydir))
	for k := range bc.keydir {
		keys = append(keys, k)
	}
	return keys
}

// Fold ...
func (bc *Bitcask) Fold(fn func([]byte, []byte) error) error {
	bc.RLock()
	defer bc.RUnlock()

	for key, value := range bc.keydir {
		data := make([]byte, value.size)
		_, err := value.file.ReadAt(data, value.position)
		if err != nil {
			return err
		}

		err = fn([]byte(key), data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Merge ...
func (bc *Bitcask) Merge(dirname string) error {
	return nil
}

// Sync ...
func (bc *Bitcask) Sync() error {
	return nil

}

// Close ...
func (bc *Bitcask) Close() error {
	return os.Remove(path.Join(bc.dirname, LOCKFILE))
}

func (bc *Bitcask) block(key, val []byte, timestamp time.Time) []byte {
	tst := uint32(timestamp.Unix())
	ksz := uint32(len(key))
	vsz := uint32(len(val))
	block := make([]byte, 16, 16+ksz+vsz)
	binary.BigEndian.PutUint32(block[4:], tst)
	binary.BigEndian.PutUint32(block[8:], ksz)
	binary.BigEndian.PutUint32(block[12:], vsz)
	block = append(block, key...)
	block = append(block, val...)
	crc := crc32.ChecksumIEEE(block[4:])
	binary.BigEndian.PutUint32(block[:], crc)
	return block
}

func (bc *Bitcask) lockfile() error {
	fi, err := bc.current.Stat()
	if err != nil {
		return err
	}
	return os.WriteFile(
		path.Join(bc.dirname, LOCKFILE),
		[]byte(fmt.Sprintf("%d:%s", os.Getpid(), fi.Name())),
		os.ModePerm,
	)
}
