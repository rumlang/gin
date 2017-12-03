package parser

import (
	"fmt"
)

const (
	_ = iota
	tokEOF
	tokOpen
	tokClose
	tokIdentifier
	tokInteger
	tokFloat
	tokString
	tokArray
)

type tokenID int

func (t tokenID) String() string {
	switch t {
	case tokEOF:
		return "EOF"
	case tokOpen:
		return "Open"
	case tokClose:
		return "Close"
	case tokIdentifier:
		return "Identifier"
	case tokInteger:
		return "Integer"
	case tokFloat:
		return "Float"
	case tokString:
		return "String"
	case tokArray:
		return "Array"
	default:
		return fmt.Sprintf("Unknown[%d", t)
	}
}

var tokenPriorities = map[tokenID]int{
	tokEOF:  0,
	tokOpen: 30,
	// Do not use 0 for closing parenthesis - this way we can trigger an error
	// when we don't reach the end of the stream (e.g., "a)b") instead of
	// considering that the parsing is done.
	tokClose:      5,
	tokIdentifier: 20,
	tokInteger:    20,
	tokFloat:      20,
	tokString:     20,
	tokArray:      20,
}

// tokenInfo give details about a token the lexer extracted - including
// information about where it comes from.
type tokenInfo struct {
	// raw is the representation of the token, identical to the input. It also
	// includes invalid bytes (i.e., not matching a valid unicode point).
	raw string
	// text is the list of valid runes of this token.
	text []rune
	// ref contains information about where the token comes from.
	ref *SourceRef
	// id is the lexer token ID, using tok* symbols.
	id tokenID
	// value is the parsed value of the token - can be a string, int, nil, ...
	value interface{}
}

// Nud implements the Token interface for the top down parser.
// It always returns a []Value. In case of errors, it will add an error through
// the context and return an empty list.
func (t tokenInfo) Nud(ctx Context) interface{} {
	switch t.id {
	case tokOpen:
		var sublist = ftokOpen(ctx)
		return []Value{NewAny(sublist, t.ref)}
	// case tokClose: // Shoud never happen
	case tokArray:
		sublist := ctx.Expression(tokenPriorities[tokOpen]).([]Value)
		array := NewAny(Identifier("array"), t.ref)
		r := NewAny(append([]Value{array}, sublist...), t.ref)
		return []Value{r}
	case tokIdentifier:
		return []Value{NewAny(Identifier(t.value.(string)), t.ref)}
	case tokInteger, tokFloat, tokString:
		return []Value{NewAny(t.value, t.ref)}
	case tokEOF:
		// Needed for when an open parenthesis (or similar) is just before the end
		// of the input.
		return []Value{}
	}

	ctx.Error(Error{
		Msg:  fmt.Sprintf("unexpected %q (token type %s) at the beginning of an expression", string(t.text), t.id),
		Code: ErrInvalidNudToken,
		Ref:  t.ref,
	})
	return []Value{}
}

func ftokOpen(ctx Context) (sublist []Value) {
	if ctx.Peek().(tokenInfo).id != tokClose {
		sublist = ctx.Expression(tokenPriorities[tokClose]).([]Value)
	}
	t := ctx.Peek().(tokenInfo)
	if t.id != tokClose {
		ctx.Error(Error{
			Msg:  fmt.Sprintf("invalid token - expected ')', got: %q", string(t.text)),
			Code: ErrMissingClosingParenthesis,
			Ref:  t.ref,
		})
		return
	}
	ctx.Advance()
	return
}

// Led implements the Token interface for the top down parser.
// It always returns a []Value. In case of errors, it will add an error through
// the context and ignore the token.
func (t tokenInfo) Led(ctx Context, left interface{}) interface{} {
	switch t.id {
	case tokOpen:
		var sublist = ftokOpen(ctx)
		return append(left.([]Value), NewAny(sublist, t.ref))
	// case tokClose: // Should never happen.
	case tokArray:
		sublist := ctx.Expression(tokenPriorities[tokOpen]).([]Value)
		array := NewAny(Identifier("array"), t.ref)
		r := NewAny(append([]Value{array}, sublist...), t.ref)
		return append(left.([]Value), r)
	case tokIdentifier:
		return append(left.([]Value), NewAny(Identifier(t.value.(string)), t.ref))
	case tokInteger, tokFloat, tokString:
		return append(left.([]Value), NewAny(t.value, t.ref))
	}

	ctx.Error(Error{
		Msg:  fmt.Sprintf("unexpected %q (token type %s) in an expression", string(t.text), t.id),
		Code: ErrInvalidLedToken,
		Ref:  t.ref,
	})
	return left
}

func (t tokenInfo) Lbp() int {
	return tokenPriorities[t.id]
}
