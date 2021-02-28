package runtime

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rumlang/rum/parser"
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
		"(strings.compare \"123\" \"123\")":                    int(0),
		"(strings.compare \"0123\" \"123\")":                   int(-1),
		"(strings.compare \"123\" \"0123\")":                   int(1),
		"(strings.contains \"seafood\" \"foo\")":               true,
		"(strings.contains \"seafood\" \"bar\")":               false,
		"(strings.contains \"seafood\" \"\")":                  true,
		"(strings.contains \"\" \"\")":                         true,
		"(strings.count \"cheese\" \"e\")":                     int(3),
		"(strings.count \"five\" \"\")":                        int(5),
		"(strings.split \"a,b,c\" \",\")":                      []string{"a", "b", "c"},
		"(strings.title \"her royal highness\")":               "Her Royal Highness",
		"(strings.to-lower \"Gophers\")":                       "gophers",
		"(strings.to-upper \"Gophers\")":                       "GOPHERS",
		"(strings.trim \" !!! Achtung! Achtung! !!!\" \" !\")": "Achtung! Achtung",
		"(strings.new-reader \"0123456789\")":                  strings.NewReader("0123456789"),
	}
	checkSExprs(t, `(import strings)`, valid)
}

func TestCSV(t *testing.T) {
	valid := map[string]interface{}{
		`(. (csv.new-reader (strings.new-reader "1,2,3,4")) read-all)`: [][]string{{"1", "2", "3", "4"}},
	}
	checkSExprs(t, `(import strings
		                    csv)`, valid)
}
