package openscad

import "github.com/lestrrat-go/option"

// EmitWriteFileOption is an option that can be passed to Emit() and WriteFile()
type EmitWriteFileOption interface {
	EmitOption
	WriteFileOption
}

type EmitFileOption interface {
	EmitOption
	emitFileOption()
}

type WriteFileOption interface {
	writeFileOption()
	option.Interface
}

type EmitOption interface {
	emitOption()
	option.Interface
}

type emitWriteFileOption struct {
	option.Interface
}

func (emitWriteFileOption) emitOption()      {}
func (emitWriteFileOption) writeFileOption() {}

type optAmalgamationKey struct{}
type optRegistryKey struct{}
type optOutputDirKey struct{}

func WithAmalgamation() EmitWriteFileOption {
	return &emitWriteFileOption{option.New(optAmalgamationKey{}, true)}
}

// WithRegistry sets the registry to use when emitting the OpenSCAD code
func WithRegistry(r *Registry) EmitWriteFileOption {
	return &emitWriteFileOption{option.New(optRegistryKey{}, r)}
}

func WithOutputDir(dir string) WriteFileOption {
	return &emitWriteFileOption{option.New(optOutputDirKey{}, dir)}
}
