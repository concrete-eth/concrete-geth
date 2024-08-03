//go:build js
// +build js

package fastcache

type Cache struct{}

func (c *Cache) Del(k []byte) {}

func (c *Cache) Get(dst []byte, k []byte) []byte {
	return nil
}

func (c *Cache) GetBig(dst []byte, k []byte) (r []byte) {
	return nil
}

func (c *Cache) Has(k []byte) bool {
	return false
}

func (c *Cache) HasGet(dst []byte, k []byte) ([]byte, bool) {
	return nil, false
}

func (c *Cache) Reset() {
}

func (c *Cache) SaveToFile(filePath string) error {
	return nil
}

func (c *Cache) SaveToFileConcurrent(filePath string, concurrency int) error {
	return nil
}

func (c *Cache) Set(k []byte, v []byte) {}

func (c *Cache) SetBig(k []byte, v []byte) {}

func New(maxBytes int) *Cache {
	return &Cache{}
}

func LoadFromFileOrNew(path string, maxBytes int) *Cache {
	return &Cache{}
}
