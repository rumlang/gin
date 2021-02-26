package runtime

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/rumlang/rum/parser"
)

type StdLib interface {
	LoadLib(ctx *Context, funcPrefix parser.Identifier)
}

// StringLib struct
type StringLib struct{}

func SIdentifierf(funcPrefix parser.Identifier, funcName parser.Identifier) parser.Identifier {
	return funcPrefix + funcName
}

// LoadLib function to StringLib struct
func (l *StringLib) LoadLib(ctx *Context, funcPrefix parser.Identifier) {
	if funcPrefix == "" {
		funcPrefix = "strings"
	}
	ctx.SetFn(SIdentifierf(funcPrefix, ".Contains"), strings.Contains, CheckArity(2))
	ctx.SetFn(SIdentifierf(funcPrefix, ".Compare"), strings.Compare, CheckArity(2))
	ctx.SetFn(SIdentifierf(funcPrefix, ".Count"), strings.Count, CheckArity(2))
	ctx.SetFn(SIdentifierf(funcPrefix, ".Join"), strings.Join, CheckArity(2))
	ctx.SetFn(SIdentifierf(funcPrefix, ".Split"), strings.Split, CheckArity(2))
	ctx.SetFn(SIdentifierf(funcPrefix, ".Title"), strings.Title, CheckArity(1))
	ctx.SetFn(SIdentifierf(funcPrefix, ".ToLower"), strings.ToLower, CheckArity(1))
	ctx.SetFn(SIdentifierf(funcPrefix, ".ToUpper"), strings.ToUpper, CheckArity(1))
	ctx.SetFn(SIdentifierf(funcPrefix, ".Trim"), strings.Trim, CheckArity(2))
	ctx.SetFn(SIdentifierf(funcPrefix, ".NewReader"), strings.NewReader, CheckArity(1))
}

// CSVLib struct
type CSVLib struct{}

// LoadLib function to StringLib struct
func (l *CSVLib) LoadLib(ctx *Context, funcPrefix parser.Identifier) {
	if funcPrefix == "" {
		funcPrefix = "csv"
	}
	ctx.SetFn(SIdentifierf(funcPrefix, ".NewReader"), csv.NewReader, CheckArity(1))
	ctx.SetFn(SIdentifierf(funcPrefix, ".NewWriter"), csv.NewWriter, CheckArity(1))
}

func loadStdLib(name string, ctx *Context, funcPrefix parser.Identifier) {
	var stdLib StdLib
	switch name {
	case "strings":
		stdLib = &StringLib{}
		stdLib.LoadLib(ctx, funcPrefix)
		return
	case "csv":
		stdLib = &CSVLib{}
		stdLib.LoadLib(ctx, funcPrefix)
		return
	default:
		panic(fmt.Sprintf("package %s not found", name))
	}
}
