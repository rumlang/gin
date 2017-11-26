package parser

import (
	"reflect"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := map[string][]tokenInfo{
		"foo": {
			{text: []rune{'f', 'o', 'o'}, id: tokIdentifier, value: "foo", ref: &SourceRef{Line: 0, Column: 0}},
		},
		"(foo)": {
			{text: []rune{'('}, id: tokOpen, ref: &SourceRef{Line: 0, Column: 0}},
			{text: []rune{'f', 'o', 'o'}, id: tokIdentifier, value: "foo", ref: &SourceRef{Line: 0, Column: 1}},
			{text: []rune{')'}, id: tokClose, ref: &SourceRef{Line: 0, Column: 4}},
		},
		" (  foo ) ": {
			{text: []rune{'('}, id: tokOpen, ref: &SourceRef{Line: 0, Column: 1}},
			{text: []rune{'f', 'o', 'o'}, id: tokIdentifier, value: "foo", ref: &SourceRef{Line: 0, Column: 4}},
			{text: []rune{')'}, id: tokClose, ref: &SourceRef{Line: 0, Column: 8}},
		},
		" \nfoo": {
			{text: []rune{'f', 'o', 'o'}, id: tokIdentifier, value: "foo", ref: &SourceRef{Line: 1, Column: 0}},
		},
		"1.2": {
			{text: []rune{'1', '.', '2'}, id: tokFloat, value: 1.2, ref: &SourceRef{Line: 0, Column: 0}},
		},
	}

	for input, expected := range tests {
		l := newLexer(NewSource(input))

		tokens := []tokenInfo{}
		for t := range l.tokens {
			tokens = append(tokens, t)
		}
		if len(tokens) != len(expected) {
			t.Fatalf("Expression %q - invalid number of tokens found; expected %#+v, found %#+v", input, expected, tokens)
		}
		for i, token := range tokens {
			// Remove source info for comparaison.
			token.ref.Source = nil
			if !reflect.DeepEqual(token, expected[i]) {
				t.Errorf("Expression %q - expected %#+v, got %#+v", input, expected[i], token)
			}
		}
	}
}

func TestLexerErrors(t *testing.T) {
	tests := map[string]int{
		// Invalid sequence at the beginning - skips the first byte and so gets a
		// ')'.
		"\xc3\x28":   1,
		"a\xc3\x28b": 3,
	}

	for input, count := range tests {
		l := newLexer(NewSource(input))
		for range l.tokens {
			count--
		}
		if count > 0 {
			t.Errorf("Input %# x ; Not enough token found - %d more expected", input, count)
		}
		if count < 0 {
			t.Errorf("Input %# x ; Too many token found - %d more than expected", input, -count)
		}
	}
}

func TestParsingExpression(t *testing.T) {
	tests := map[string]int{
		"":                  -1,
		")":                 -1,
		"(":                 -1,
		"a)b":               -1,
		")b":                -1,
		"a(b":               -1,
		"()":                0,
		"(foo)":             1,
		"(a b)":             2,
		"(a (b c))":         2,
		"(a (b c) d (e f))": 4,
		"(a\nb)":            2,
		"(1.2 .3)":          2,

		// Test strings
		`(" `:        -1,
		`("a b")`:    1,
		`("a \" b")`: 1,

		// Test comments
		"( ; )":        -1,
		"(a ; b \n c)": 2,

		// Test array
		"(a array()":             -1,
		"(array (a array(b c)))": 2,
		"(array(a b) c)":         3,
		"(a (array (b c)))":      2,
	}

	for input, count := range tests {
		r, err := Parse(NewSource(input))

		if count < 0 {
			if err == nil {
				t.Errorf("Input %q parsed instead of generating error", input)
			}

			// Check that the error does not have issue generating context and is of
			// the right type.
			m := err.(MultiError)
			m.Error()
			continue
		}

		if err != nil {
			t.Errorf("Input %q - parsing errors: %v", input, err)
			continue
		}

		if count != len(r.Value().([]Value)) {
			t.Errorf("Input %q - expected %d children, got %d: %v", input, count, len(r.Value().([]Value)), r)
		}
	}
}

func TestParsingAtoms(t *testing.T) {
	r, err := Parse(NewSource("foo"))
	if err != nil {
		t.Fatalf("Expected parsable code, got: %v", err)
	}

	if string(r.Value().(Identifier)) != "foo" {
		t.Errorf("Expected 'foo', got: %v", r.Value())
	}
}

func TestParsingErrors(t *testing.T) {
	type foo struct {
		code ErrorCode
		ref  SourceRef
	}

	tests := map[string]foo{
		"(+": {
			code: ErrMissingClosingParenthesis,
			ref: SourceRef{
				Line:   0,
				Column: 2,
			},
		},
		"(": {
			code: ErrMissingClosingParenthesis,
			ref: SourceRef{
				Line:   0,
				Column: 1,
			},
		},
	}

	for input, expected := range tests {
		_, err := Parse(NewSource(input))
		errs := err.(MultiError)
		if len(errs.Errors) != 1 {
			t.Errorf("Input %q should have 1 error; instead: %v", input, errs)
		} else {
			err := errs.Errors[0].(Error)
			if err.Code != expected.code {
				t.Errorf("Input %q should have returned error code %d; instead: %d", input, expected.code, err.Code)
			}
			// Remove source info for comparaison.
			err.Ref.Source = nil
			if !reflect.DeepEqual(*err.Ref, expected.ref) {
				t.Errorf("Input %q - expected %#+v, got %#+v", input, expected.ref, err.Ref)
			}
		}
	}
}

func TestSource(t *testing.T) {
	s := NewSource("a\n ")
	if l, _ := s.Line(0); string(l) != "a\n" {
		t.Errorf("expected %q, got %q", "a\n", l)
	}
	if l, _ := s.Line(1); string(l) != " " {
		t.Errorf("expected %q, got %q", " ", l)
	}
}

func TestSourceRef(t *testing.T) {
	// Check that SourceRef context generator is resilient - it should not
	// failed, even if the reference is broken.
	src := NewSource("foo\nbar")

	refs := []SourceRef{
		{},
		{Source: src},
		{Source: src, Line: 0},
		{Source: src, Line: -1},
		{Source: src, Line: 10},
		{Source: src, Line: 1, Column: -1},
		{Source: src, Line: 1, Column: 0},
		{Source: src, Line: 1, Column: 100},
		{Source: src, Column: 1},
	}

	for _, ref := range refs {
		(&ref).Context("  ")
	}
}
