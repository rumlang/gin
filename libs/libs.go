package libs

import "github.com/rumlang/rum/runtime"

//LoadStdLib the basic lib into Context
func LoadStdLib(ctx *runtime.Context) {
	LoadStrings(ctx)
	LoadCSV(ctx)
}
