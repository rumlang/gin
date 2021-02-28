package runtime

import "github.com/rumlang/rum/parser"

// ConcatIdentifier ...
func ConcatIdentifier(funcPrefix parser.Identifier, funcName parser.Identifier) parser.Identifier {
	return funcPrefix + funcName
}
