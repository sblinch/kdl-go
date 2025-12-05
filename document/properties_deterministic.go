//go:build kdldeterministic

//
// properties_ordered.go provides a deterministic version of the Properties type to ensure that output is always
// consistent during testing; it is slower and more memory-hungry than properties_unordered.go and thus is disabled by
// default; enable with -tags kdldeterministic

package document

import (
	"github.com/sblinch/kdl-go/internal/tokenizer"
)

type Properties struct {
	order []string
	props map[string]*Value
}

func (p *Properties) Allocated() bool {
	return p.order != nil
}
func (p *Properties) Alloc() {
	p.order = make([]string, 0, 8)
	p.props = make(map[string]*Value, 8)
}

func (p *Properties) Len() int {
	return len(p.order)
}

func (p Properties) Unordered() map[string]*Value {
	return p.props
}

func (p Properties) Get(key string) (*Value, bool) {
	v, ok := p.props[key]
	return v, ok
}

func (p *Properties) Add(name string, val *Value) {
	if _, exists := p.props[name]; !exists {
		p.order = append(p.order, name)
	}
	p.props[name] = val
}

func (p *Properties) Exist() bool {
	return len(p.order) > 0
}

func (p *Properties) String() string {
	b := make([]byte, 0, len(p.order)*(1+8+1+8))
	for _, k := range p.order {
		v := p.props[k]
		b = append(b, ' ')
		if len(k) > 0 && tokenizer.IsBareIdentifier(k, 0) {
			b = append(b, k...)
		} else {
			b = AppendQuotedString(b, k, '"')
		}
		b = append(b, '=')
		b = append(b, v.FormattedString()...)
	}
	return string(b)
}

func (p *Properties) UnformattedString() string {
	b := make([]byte, 0, len(p.order)*(1+8+1+8))
	for _, k := range p.order {
		v := p.props[k]
		b = append(b, ' ')
		if len(k) > 0 && tokenizer.IsBareIdentifier(k, 0) {
			b = append(b, k...)
		} else {
			b = AppendQuotedString(b, k, '"')
		}
		b = append(b, '=')
		b = append(b, v.UnformattedString()...)
	}
	return string(b)
}

func (p Properties) AppendTo(b []byte) []byte {
	required := len(p.order) * (1 + 8 + 1 + 8)
	if cap(b)-len(b) < required {
		r := make([]byte, 0, len(b)+required)
		r = append(r, b...)
		b = r
	}
	for _, k := range p.order {
		v := p.props[k]
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
	return b
}
