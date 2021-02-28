package runtime

import (
	"encoding/csv"

	"github.com/rumlang/rum/parser"
)

// CSVLib struct
type CSVLib struct{}

// LoadLib function to StringLib struct
func (l *CSVLib) LoadLib(ctx *Context, funcPrefix parser.Identifier) {
	if funcPrefix == "" {
		funcPrefix = "csv"
	}
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".new-reader"), csv.NewReader, CheckArity(1))
	ctx.SetFn(ConcatIdentifier(funcPrefix, ".new-writer"), csv.NewWriter, CheckArity(1))
}
