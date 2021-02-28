package runtime

import (
	"strings"

	"github.com/rumlang/rum/parser"
)

// ConcatIdentifier ...
func ConcatIdentifier(funcPrefix parser.Identifier, funcName parser.Identifier) parser.Identifier {
	return funcPrefix + funcName
}

// MethodNameTransform ...
func MethodNameTransform(name string) (newName string) {
	letters := strings.Split(name, "-")
	for _, letter := range letters {
		newName += strings.Title(strings.ToLower(letter))
	}
	return
}
