package openscad

import "github.com/lestrrat-go/option"

// EmitWriteFileOption is an option that can be passed to Emit() and WriteFile()
type EmitWriteFileOption interface {
	EmitOption
	WriteFileOption
}

// EmitFileWriteFileOption is an option that can be passed to Emit(), EmitFile() and WriteFile()
type EmitFileWriteFileOption interface {
	EmitOption
	EmitFileOption
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

type emitFileWriteFileOption struct {
	option.Interface
}

func (emitFileWriteFileOption) emitOption()      {}
func (emitFileWriteFileOption) emitFileOption()  {}
func (emitFileWriteFileOption) writeFileOption() {}

type optAmalgamationKey struct{}
type optRegistryKey struct{}
type optOutputDirKey struct{}

func WithAmalgamation() EmitFileWriteFileOption {
	return &emitFileWriteFileOption{option.New(optAmalgamationKey{}, true)}
}

// WithRegistry sets the registry to use when emitting the OpenSCAD code
func WithRegistry(r *Registry) EmitFileWriteFileOption {
	return &emitFileWriteFileOption{option.New(optRegistryKey{}, r)}
}

func WithOutputDir(dir string) WriteFileOption {
	return &emitWriteFileOption{option.New(optOutputDirKey{}, dir)}
}
