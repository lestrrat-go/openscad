package ast

import (
	"sync"
)

var globalRegistry = &Registry{
	storage: make(map[string]Stmt),
}

func Register(name string, s Stmt) error {
	return globalRegistry.Register(name, s)
}

func Lookup(name string) (Stmt, bool) {
	return globalRegistry.Lookup(name)
}

// Registry is used to register pieces of OpenSCAD code to a virtual
// filename.
type Registry struct {
	mu      sync.RWMutex
	storage map[string]Stmt
}

func (r *Registry) Register(name string, s Stmt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storage[name] = s
	return nil
}

func (r *Registry) Lookup(name string) (Stmt, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.storage == nil {
		return nil, false
	}
	s, ok := r.storage[name]
	return s, ok
}

/*
// Code generates code for the given name. By default it generates
// regular files, but if you set the XXXX option, it can generate
// single piece of code with all `use` and `include` statements
// expanded appropriately.
func (r *Registry) Code(ctx context.Context, w io.Writer, name string) error {

}
*/
