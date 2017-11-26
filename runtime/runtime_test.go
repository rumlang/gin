package runtime

import (
	"fmt"
	"reflect"
	"testing"

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

func TestArray(t *testing.T) {
	n := mustEval("(array (+ 1 2))").Value().([]parser.Value)
	if len(n) != 3 {
		t.Errorf("Expected 3 children, got: %v", n)
	}
}

func TestValid(t *testing.T) {
	valid := map[string]interface{}{
		// Test operators
		"(+ 1 2)":            int64(3),
		"(+int64 1 2)":       int64(3),
		"(+ 1.0 2.0)":        float64(3),
		"(+float64 1.0 2.0)": float64(3),
		"(- 1 2)":            int64(-1),
		"(- 1.0 2.0)":        float64(-1.0),
		"(* 3.0 2.0)":        float64(6),
		"(* 3 2)":            int64(6),
		"(== 3 2)":           false,
		"(== 3 3)":           true,
		"(== 3 3 1)":         false,
		"(== 3 3 3)":         true,
		"(== 3.0 3.0)":       true,
		"(== 3.0 2.0)":       false,
		"(!= 3 2)":           true,
		"(!= 3 3)":           false,
		"(!= 3.0 3.0)":       false,
		"(!= 3.0 2.0)":       true,
		"(< 3 2)":            false,
		"(< 2 3)":            true,
		"(< 3.0 3.0)":        false,
		"(< 3.0 2.0)":        false,
		"(<= 3 2)":           false,
		"(<= 2 3)":           true,
		"(<= 3.0 3.0)":       true,
		"(<= 3.0 2.0)":       false,
		"(> 3 2)":            true,
		"(> 2 3)":            false,
		"(> 3.0 3.0)":        false,
		"(> 3.0 2.0)":        true,
		"(>= 3 2)":           true,
		"(>= 2 3)":           false,
		"(>= 3.0 3.0)":       true,
		"(>= 3.0 2.0)":       true,
		// Test 'package'
		`(package "main" 1 (+ 1 1))`:       int64(2),
		`(package "main")`:                 nil,
		`(package "main" (+ 1 2) (+ 3 4))`: int64(7),
		`(package "main" (print 1 2))`:     nil,
		// Test 'let'
		`(package "main" (let a 5) a)`:           int64(5),
		`(package "main" (let a 5) (let a 4) a)`: int64(4),
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
		`(package "main" (let a '(+ 1 2)) (eval a))`: int64(3),
		// Test empty
		`()`: nil,
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
		"'(1 2)": {int64(1), int64(2)},
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
		"(+ 1.0 2)",
		"(+int64 1.0 2.0)",
		"(+float64 1 2)",
		"(*int64 1.0 2.0)",
		"(*float64 1 2)",
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
