package openscad

import (
	"io/fs"

	"github.com/lestrrat-go/option"
)

type optLookupNameKey struct{}
type optFSKey struct{}

type ParseFileOption interface {
	parseFileOption()
	RegisterFileOption
	option.Interface
}

type RegisterFileOption interface {
	registerFileOption()
	option.Interface
}

type registerFileOption struct {
	option.Interface
}

type parseFileOption struct {
	option.Interface
}

func (registerFileOption) registerFileOption() {}

func (parseFileOption) parseFileOption()    {}
func (parseFileOption) registerFileOption() {}

func WithLookupName(name string) RegisterFileOption {
	return &registerFileOption{option.New(optLookupNameKey{}, name)}
}

func WithFS(src fs.FS) ParseFileOption {
	return &parseFileOption{option.New(optFSKey{}, src)}
}
