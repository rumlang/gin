package libs

import (
	"encoding/csv"

	"github.com/rumlang/rum/runtime"
)

//LoadCSV load the library to cvs processing
func LoadCSV(ctx *runtime.Context) {
	ctx.SetFn("encoding/csv.NewReader", csv.NewReader, runtime.CheckArity(1))
	ctx.SetFn("encoding/csv.NewWriter", csv.NewWriter, runtime.CheckArity(1))
}
