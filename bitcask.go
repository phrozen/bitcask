// Package bitcask ...
package bitcask

const (
	TOMBSTONE = "bitcask_tombstone"
)

// Opts
type Opts struct {
	ReadOnly    bool
	SyncOnPut   bool
	MaxFileSize int
}

// Default ...
var Default = Opts{}

// Bitcask ...
type Bitcask struct {
}

// Open ...
func Open(dirname string, opts Opts) (*Bitcask, error) {
	return nil, nil
}

// Get ...
func (bc *Bitcask) Get(key []byte) ([]byte, error) {
	return nil, nil
}

// Put ...
func (bc *Bitcask) Put(key, value []byte) error {
	return nil
}

// Delete ...
func (bc *Bitcask) Delete(key []byte) error {
	return nil
}

// List ...
func (bc *Bitcask) List() ([][]byte, error) {
	return nil, nil
}

// Fold ...
func (bc *Bitcask) Fold(func([]byte, []byte) error) error {
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
	return nil
}
