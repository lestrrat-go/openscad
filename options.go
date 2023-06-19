package openscad

import "github.com/lestrrat-go/option"

type optLookupNameKey struct{}

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
