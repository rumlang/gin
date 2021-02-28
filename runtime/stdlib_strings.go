package runtime

import (
	"strings"

	"github.com/rumlang/rum/parser"
)

// StringsLib struct
type StringsLib struct{}

// LoadLib function to StringLib struct
func (l *StringsLib) LoadLib(ctx *Context, funcPrefix parser.Identifier) {
	if funcPrefix == "" {
		funcPrefix = "strings"
	}
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".contains"), strings.Contains, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".compare"), strings.Compare, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".count"), strings.Count, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".join"), strings.Join, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".split"), strings.Split, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".title"), strings.Title, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".to-lower"), strings.ToLower, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".to-upper"), strings.ToUpper, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".trim"), strings.Trim, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".new-reader"), strings.NewReader, CheckArity(1))
}
