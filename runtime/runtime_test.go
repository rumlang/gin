package runtime

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rumlang/rum/parser"
)

func mustParse(s string) parser.Value {
	v, err := parser.Parse(parser.NewSource(s))
	if err != nil {
		panic(fmt.Sprintf("Unable to parse %q: %v", s, err))
	}
	return v
}

func mustEval(s string) parser.Value {
	return NewContext(nil).MustEval(mustParse(s))
}

//RunSExpressions execute in sequence the sexprs in exprs param using the Context c
func RunSExpressions(c *Context, exprs []string, t *testing.T) {
	for _, expr := range exprs {
		_, err := c.TryEval(mustParse(expr))
		if err != nil {
			t.Fatalf(expr, err)
		}
	}
}

func TestArray(t *testing.T) {
	n := mustEval("(array (+ 1 2))").Value().([]parser.Value)
	if len(n) != 3 {
		t.Errorf("Expected 3 children, got: %v", n)
	}
}

func TestValid(t *testing.T) {
	valid := map[string]interface{}{
		// Test operators
		"(+ 1 2)":           int64(3),
		"(+ 1.0 2.0)":       float64(3),
		"(+ 1.0 2)":         float64(3),
		"(+ 1 2.0)":         float64(3),
		"(+ 1.1 2)":         float64(3.1),
		"(+ 0.1 0.2 0.3 1)": float64(1.6),
		"(- 1 2)":           int64(-1),
		"(- 1.0 2.0)":       float64(-1.0),
		"(* 3.0 2.0)":       float64(6),
		"(* 3 2)":           int64(6),
		"(** 2.0 2.0)":      float64(4),
		"(** 4.0 2.0)":      float64(16),
		"(** 2.0 1.9)":      float64(3.732131966147229),
		"(** 2.0 1.9 1.8)":  float64(10.703420438288894),
		"(== 3 2)":          false,
		"(== 3 3)":          true,
		"(== 3 3 1)":        false,
		"(== 3 3 3)":        true,
		"(== 3.0 3.0)":      true,
		"(== 3.0 2.0)":      false,
		`(== "a" "a")`:      true,
		`(== "a" "b")`:      false,
		"(!= 3 2)":          true,
		"(!= 3 3)":          false,
		"(!= 3.0 3.0)":      false,
		"(!= 3.0 2.0)":      true,
		"(< 3 2)":           false,
		"(< 2 3)":           true,
		"(< 3.0 3.0)":       false,
		"(< 3.0 2.0)":       false,
		"(<= 3 2)":          false,
		"(<= 2 3)":          true,
		"(<= 3.0 3.0)":      true,
		"(<= 3.0 2.0)":      false,
		"(> 3 2)":           true,
		"(> 2 3)":           false,
		"(> 3.0 3.0)":       false,
		"(> 3.0 2.0)":       true,
		"(>= 3 2)":          true,
		"(>= 2 3)":          false,
		"(>= 3.0 3.0)":      true,
		"(>= 3.0 2.0)":      true,
		// Test 'package'
		`(package "main" 1 (+ 1 1))`:       int64(2),
		`(package "main")`:                 nil,
		`(package "main" (+ 1 2) (+ 3 4))`: int64(7),
		`(package "main" (print 1 2))`:     nil,
		// Test 'let'
		`(package "main" (let a 5) a)`:           int64(5),
		`(package "main" (let a 5) (let b 4) b)`: int64(4),
		// Test 'if'
		"(if true 7)":    int64(7),
		"(if false 7)":   nil,
		"(if false 7 8)": int64(8),
		// Test 'lambda'
		`(package "main" (let d (lambda (n) (+ n n))) (d 3))`: int64(6),
		// Test 'def'
		`(package "main" (def d(n) (+ n n)) (d 3))`: int64(6),
		// Test that inner scopes are not override outer scope.
		`(package "main" (let n 7) (let d (lambda (n) (+ n n))) (+ n (d 3)))`: int64(13),
		// Test float
		".3": float64(.3),
		// Test length
		"(len (array (1 2 3)))": int64(3),
		// Test string
		`"plop"`:   "plop",
		`"p\"lop"`: `p"lop`,
		// Test eval
		`(package "main" (let a (array (+ 1 2))) (eval a))`: int64(3),
		// Test empty
		`()`: nil,
		// Test for
		`(for print (array (10 20 30)))`: nil,
		// Test sprintf
		`(sprintf "%02d %02d" 1 2)`:                               "01 02",
		`(sprintf "%02X" 255)`:                                    "FF",
		`(sprintf "%0.9g" 1.99999999)`:                            "1.99999999",
		`(sprintf "%v" 1.99999999)`:                               "1.99999999",
		`(sprintf "%v" 1.9999999999999998)`:                       "1.9999999999999998",
		`(sprintf "10 is %q char" 10)`:                            "10 is '\\n' char",
		`(sprintf "%% literal percent sign; not literal %v" "%")`: `% literal percent sign; not literal %`,
		`(sprintf "%t is not %t" true false)`:                     "true is not false",
		`(sprintf "%T" 1.99999999)`:                               "float64",
		`(sprintf "%T" true)`:                                     "bool",
	}

	for input, expected := range valid {
		r := mustEval(input).Value()
		if !reflect.DeepEqual(r, expected) {
			t.Errorf("Input %q -- expected <%T>%#+v, got: <%T>%#+v", input, expected, expected, r, r)
		}
	}
}

func TestValidList(t *testing.T) {
	valid := map[string][]interface{}{
		// Test single array notation
		"(array (1 2))": {int64(1), int64(2)},
	}

	for input, expected := range valid {
		r := mustEval(input).Value()

		if _, ok := r.([]parser.Value); !ok {
			t.Errorf("Expected a []Value; got: %T", r)
		}

		for i, v := range r.([]parser.Value) {
			if v.Value() != expected[i] {
				t.Errorf("Item %d - expected %v, got: %v", i, expected[i], v.Value())
			}
		}
	}
}

func TestPanic(t *testing.T) {
	panics := []string{
		"(6)",
		"(+ 1 (2))",
		"(panic 10)",
	}

	for _, s := range panics {
		var r interface{}
		func() {
			defer func() {
				r = recover()
			}()
			mustEval(s)
		}()

		if r == nil {
			t.Fatalf("%q should have generated a panic.", s)
		}

		// Now try with TryEval
		root := mustParse(s)
		_, err := NewContext(nil).TryEval(root)
		if err == nil {
			t.Fatalf("%q should have generated an error.", s)
		}
	}
}

func TestUnknownVariable(t *testing.T) {
	s := "(a)"

	var r interface{}
	func() {
		defer func() {
			r = recover()
		}()
		mustEval(s)
	}()
	if r == nil {
		t.Fatalf("Expected a panic, got nothing")
	}

	e, ok := r.(*Error)
	if !ok {
		t.Fatalf("Expected a runtime.Error; instead: %v", r)
	}
	if e.Code != ErrUnknownVariable {
		t.Errorf("Expected an UnknownVariable; instead: %v", e)
	}

	// Now try with TryEval
	root := mustParse(s)
	_, err := NewContext(nil).TryEval(root)
	if err == nil {
		t.Fatalf("%q should have generated an error.", s)
	}
}

func TestGoFunction(t *testing.T) {
	c := NewContext(nil)

	c.SetFn("rand", rand.Int63)
	c.SetFn("sin", math.Sin)
	c.SetFn("split", strings.Split)

	exprs := []string{
		"(rand)",
		"(sin 1.0)",
		"(sin 1.0)",
		"(print(split \"1,2,3,4,5\" \",\" ))",
	}
	RunSExpressions(c, exprs, t)
}

func TestAdapterGoFunctions(t *testing.T) {
	c := NewContext(nil)

	c.SetFn("rand/Rand", rand.Int63, CheckArity(0))
	c.SetFn("sin", math.Sin, CheckArity(1), ParamToFloat64(0))
	c.SetFn("randn", rand.Int63n, CheckArity(1), ParamToInt64(0))
	c.SetFn("compare", strings.Compare, CheckArity(2))

	exprs := []string{
		"(rand/Rand)",
		"(sin 1.0)",
		"(sin 2)",
		"(randn 100)",
		"(randn 100.0)",
		"(compare \"test\" \"test\")",
	}

	RunSExpressions(c, exprs, t)
}

func TestInvoke(t *testing.T) {
	c := NewContext(nil)

	c.SetFn("now", time.Now, CheckArity(0))
	c.SetFn("http/Get", http.Get, CheckArity(1))
	c.SetFn("ioutil/ReadAll", ioutil.ReadAll, CheckArity(1))
	c.SetFn("bytes/NewBuffer", bytes.NewBuffer, CheckArity(1))

	exprs := []string{
		"(let resp (http/Get \"http://www.google.com/robots.txt\"))",
		"(. resp status)",
		"(let respbytes (ioutil/ReadAll (. resp body)))",
		"(let buf (bytes/NewBuffer respbytes))",
		"(. buf string)",
		"(print (. buf string))",
	}

	RunSExpressions(c, exprs, t)
}

func TestImmutable(t *testing.T) {
	s := "(package \"main\" (let a 5) (let a 4) a)"
	var r interface{}
	func() {
		defer func() {
			r = recover()
		}()
		mustEval(s)
	}()
	if r == nil {
		t.Fatalf("%q should have generated a panic.", s)
	}

	s = `(package "main" (def d(n) (+ n n)) (def d(n) (* 2 n)) (d 3))`
	func() {
		defer func() {
			r = recover()
		}()
		mustEval(s)
	}()
	if r != nil {
		t.Fatalf("%q not should have generated a panic.", s)
	}
}

// TODO: fixed coerce tests
/*
func TestCoerce(t *testing.T) {
	c := NewContext(nil)

	c.SetFn("now", time.Now, CheckArity(0))
	c.RegisterType((*time.Duration)(nil))
	c.RegisterType((*time.Time)(nil))
	c.RegisterType((*http.Transport)(nil))

	exprs := []string{
		"(let n (now))",
		"(print n)",
		`(print (coerce time.Duration 0)
		        (coerce time.Duration 10000000)
				(coerce time.Duration 1000000)
				(coerce time.Duration 100000))`,
	}

	RunSExpressions(c, exprs, t)
}
*/
