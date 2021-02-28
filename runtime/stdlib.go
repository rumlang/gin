package runtime

import (
	"fmt"

	"github.com/rumlang/rum/parser"
)

// StdLib ...
type StdLib interface {
	LoadLib(ctx *Context, funcPrefix parser.Identifier)
}

func loadStdLib(name string, ctx *Context, funcPrefix parser.Identifier) {
	var stdLib StdLib
	switch name {
	case "strings":
		stdLib = &StringsLib{}
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
