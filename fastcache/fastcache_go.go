//go:build !js
// +build !js

package fastcache

import "github.com/VictoriaMetrics/fastcache"

type Cache = fastcache.Cache

func New(maxBytes int) *Cache {
	return fastcache.New(maxBytes)
}

func LoadFromFileOrNew(path string, maxBytes int) *Cache {
	return fastcache.LoadFromFileOrNew(path, maxBytes)
}
