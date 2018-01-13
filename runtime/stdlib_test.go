package runtime

import (
	"fmt"
	"github.com/rumlang/rum/parser"
	"reflect"
	"strings"
	"testing"
)

func checkSExprs(t *testing.T, firstParser string, valid map[string]interface{}) {
	c := NewContext(nil)
	p, err := parser.Parse(parser.NewSource(firstParser))
	if err != nil {
		t.Fatalf(fmt.Sprintf("unable to parse %q, %s", firstParser, err.Error()))
	}
	_, err = c.TryEval(p)
	if err != nil {
		t.Fatalf(firstParser, err.Error())
	}

	for input, expected := range valid {
		p, err := parser.Parse(parser.NewSource(input))
		if err != nil {
			panic(fmt.Sprintf("Unable to parse %q: %v", input, err))
		}

		val, err := c.TryEval(p)
		if err != nil {
			t.Fatalf(input, err.Error())
		}

		r := val.Value()
		if !reflect.DeepEqual(r, expected) {
			t.Errorf("Input %q -- expected <%T>%#+v, got: <%T>%#+v", input, expected, expected, r, r)
		}
	}
}

func TestStrings(t *testing.T) {
	valid := map[string]interface{}{
		"(strings.Compare \"123\" \"123\")":                    int(0),
		"(strings.Compare \"0123\" \"123\")":                   int(-1),
		"(strings.Compare \"123\" \"0123\")":                   int(1),
		"(strings.Contains \"seafood\" \"foo\")":               true,
		"(strings.Contains \"seafood\" \"bar\")":               false,
		"(strings.Contains \"seafood\" \"\")":                  true,
		"(strings.Contains \"\" \"\")":                         true,
		"(strings.Count \"cheese\" \"e\")":                     int(3),
		"(strings.Count \"five\" \"\")":                        int(5),
		"(strings.Split \"a,b,c\" \",\")":                      []string{"a", "b", "c"},
		"(strings.Title \"her royal highness\")":               "Her Royal Highness",
		"(strings.ToLower \"Gophers\")":                        "gophers",
		"(strings.ToUpper \"Gophers\")":                        "GOPHERS",
		"(strings.Trim \" !!! Achtung! Achtung! !!!\" \" !\")": "Achtung! Achtung",
		"(strings.NewReader \"0123456789\")":                   strings.NewReader("0123456789"),
	}
	checkSExprs(t, `(import "strings")`, valid)
}
