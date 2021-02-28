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
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Contains"), strings.Contains, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Compare"), strings.Compare, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Count"), strings.Count, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Join"), strings.Join, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Split"), strings.Split, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Title"), strings.Title, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".ToLower"), strings.ToLower, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".ToUpper"), strings.ToUpper, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".Trim"), strings.Trim, CheckArity(2))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".NewReader"), strings.NewReader, CheckArity(1))
}
