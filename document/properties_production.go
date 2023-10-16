//go:build !kdldeterministic

//
// properties_unordered.go could otherwise be implemented as a simple map, but provides additional methods to make it a
// drop-in replacement for properties_ordered.go for use during testing.

package document

import (
	"github.com/sblinch/kdl-go/internal/tokenizer"
)

// Properties represents a list of properties for a Node
type Properties map[string]*Value

// Allocated indicates whether the property list has been allocated
func (p Properties) Allocated() bool {
	return p != nil
}

// Alloc allocates the property list
func (p *Properties) Alloc() {
	*p = make(map[string]*Value)
}

// Get returns Properties[key]
func (p Properties) Get(key string) (*Value, bool) {
	v, ok := p[key]
	return v, ok
}

// Len returns the number of properties
func (p Properties) Len() int {
	return len(p)
}

// Unordered returns the unordered property map; this simply passes through p in this implementation but is provided
// as it is necessary in the deterministic version
func (p Properties) Unordered() map[string]*Value {
	return p
}

// Add adds a property to the list
func (p Properties) Add(name string, val *Value) {
	p[name] = val
}

// Exist indicates whether any properties exist
func (p Properties) Exist() bool {
	return len(p) > 0
}

// String returns the KDL representation of the property list, formatting numbers per their flags
func (p Properties) String() string {
	b := make([]byte, 0, len(p)*(1+8+1+8))
	for k, v := range p {
		b = append(b, ' ')
		if len(k) > 0 && tokenizer.IsBareIdentifier(k, 0) {
			b = append(b, k...)
		} else {
			b = AppendQuotedString(b, k, '"')
		}
		b = append(b, '=')
		// property values must always be quoted
		b = append(b, v.FormattedString()...)
	}
	return string(b)
}

// UnformattedString returns the KDL representation of the property list, formatting numbers in decimal
func (p Properties) UnformattedString() string {
	b := make([]byte, 0, len(p)*(1+8+1+8))
	for k, v := range p {
		b = append(b, ' ')
		if len(k) > 0 && tokenizer.IsBareIdentifier(k, 0) {
			b = append(b, k...)
		} else {
			b = AppendQuotedString(b, k, '"')
		}
		b = append(b, '=')
		// property values must always be quoted
		b = append(b, v.UnformattedString()...)
	}
	return string(b)
}
