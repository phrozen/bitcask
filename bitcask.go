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
	LOCKFILE      = ".lockfile"
	TOMBSTONE     = "bitcask_tombstone"
	MAX_FILE_SIZE = 2 * 1024 * 1024 * 1024
)

// Opts
type Configuration struct {
	ReadOnly     bool
	SyncOnPut    bool
	MaxKeySize   int
	MaxFileSize  int
	MaxValueSize int
}

// Default options
var Default = Configuration{}

// Bitcask ...
type Bitcask struct {
	sync.RWMutex
	config  Configuration
	current *os.File
	cursor  int
	dirname string
	keydir  map[string]lookup
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
func Open(dirname string, config Configuration) (*Bitcask, error) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(path.Join(dirname, LOCKFILE))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf(
			"found: %s\n%s\ndatabase is locked, or did you forget to close it?",
			path.Join(dirname, LOCKFILE), string(data),
		)
	}

	bc := &Bitcask{
		config:  config,
		dirname: dirname,
		keydir:  make(map[string]lookup),
	}

	err = bc.new()
	if err != nil {
		return nil, err
	}

	return bc, nil
}

// Get retrieves the value for a given key fromt he store.
// If a key does not exist, both value and error will be nil.
func (bc *Bitcask) Get(key []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("<nil> values not allowed")
	}

	value, exist := bc.keydir[string(key)]
	if !exist {
		return nil, nil
	}

	if value.file == bc.current {
		bc.RLock()
		defer bc.RUnlock()
	}

	data := make([]byte, value.size)
	_, err := value.file.ReadAt(data, value.position)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Put saves the new key/value pair and syncs the store
func (bc *Bitcask) Put(key, value []byte) error {
	if key == nil || value == nil {
		return fmt.Errorf("<nil> values not allowed")
	}

	now := time.Now()
	data := bc.block(key, value, now)

	if bc.cursor+len(data) > MAX_FILE_SIZE {
		if err := bc.new(); err != nil {
			return err
		}
	}

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
	if key == nil {
		return fmt.Errorf("<nil> values not allowed")
	}

	bc.Lock()
	defer bc.Unlock()

	data := bc.block(key, []byte(TOMBSTONE), time.Now())
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

func (bc *Bitcask) new() error {
	bc.Lock()
	defer bc.Unlock()
	// Create a new active file
	filename := fmt.Sprintf("%d.cask", time.Now().UnixMilli())
	file, err := os.OpenFile(path.Join(bc.dirname, filename), os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	// TODO: Find a way to change ex-active file to readOnly?
	// Assign new active file
	bc.current = file
	bc.cursor = 0
	fi, err := bc.current.Stat()
	if err != nil {
		return err
	}
	// Write lockfile
	return os.WriteFile(
		path.Join(bc.dirname, LOCKFILE),
		[]byte(fmt.Sprintf("pid: %d - file: %s [active]", os.Getpid(), fi.Name())),
		os.ModePerm,
	)
}

func (bc *Bitcask) load(filename string) error {
	return nil
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
