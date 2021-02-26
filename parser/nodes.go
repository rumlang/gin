package parser

import (
	"fmt"
	"strings"
)

// Value encapsulate all data that the language manipulate. There is a layer of
// indirection between glop and go in order to allow for extra annotation and
// reduce type conversions needed when writing code with glop.
type Value interface {
	Value() interface{}
	String() string
	Ref() *SourceRef
}

// Identifier represents a parsed identifier - as opposed to a regular string
// constant.
type Identifier string

func (id Identifier) String() string {
	// return fmt.Sprintf("<id>%q", string(id))
	return string(id)
}

// Any implements Value interface, provided an encapsulation for any valid
// Go type.
type Any struct {
	value interface{}
	ref   *SourceRef
}

// Value function return value of current node
func (a Any) Value() interface{} {
	return a.value
}

func (a Any) String() string {
	switch data := a.value.(type) {
	case []Value:
		var elt []string
		for _, v := range data {
			elt = append(elt, v.String())
		}
		return fmt.Sprintf("<[]Value>(%s)", strings.Join(elt, " "))
	case Identifier:
		return data.String()
	default:
		return fmt.Sprintf("<%T>%#+v", data, data)
	}
}

// Ref funcrion return current value by reference
func (a Any) Ref() *SourceRef {
	return a.ref
}

// NewAny create new instance of Any
func NewAny(v interface{}, ref *SourceRef) Value {
	return Any{
		value: v,
		ref:   ref,
	}
}
