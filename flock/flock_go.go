//go:build !js
// +build !js

package flock

import (
	"github.com/gofrs/flock"
)

type Flock = flock.Flock

func New(path string) *Flock {
	return flock.New(path)
}
