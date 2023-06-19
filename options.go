package openscad

import (
	"io/fs"

	"github.com/lestrrat-go/option"
)

type optLookupNameKey struct{}
type optFSKey struct{}

type RegisterFileOption interface {
	registerFileOption()
	option.Interface
}

type registerFileOption struct {
	option.Interface
}

func (registerFileOption) registerFileOption() {}

func WithLookupName(name string) RegisterFileOption {
	return &registerFileOption{option.New(optLookupNameKey{}, name)}
}

func WithFS(src fs.FS) RegisterFileOption {
	return &registerFileOption{option.New(optFSKey{}, src)}
}
