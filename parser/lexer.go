package parser

import (
	"fmt"
	"strconv"
	"unicode"
)

const (
	// ErrMissingClosingParenthesis a closing parenthesis was expected, found something else instead.
	ErrMissingClosingParenthesis = iota
	// ErrInvalidNudToken an unknown/invalid token was found at the beginning of an expression.
	ErrInvalidNudToken
	// ErrInvalidLedToken an unknown/invalid token was found in an expression.
	ErrInvalidLedToken
)

// ErrorCode type to parser errors
type ErrorCode int

func (c ErrorCode) String() string {
	switch c {
	case ErrMissingClosingParenthesis:
		return "MissingClosingParenthesis"
	case ErrInvalidNudToken:
		return "InvalidNudToken"
	case ErrInvalidLedToken:
		return "InvalidLedToken"
	default:
		return fmt.Sprintf("Unknown[%d]", c)
	}
}

// Error provides more details about a given parser error with references to
// the source. It satisfies the error interface.
type Error struct {
	Msg  string
	Code ErrorCode
	Ref  *SourceRef
}

func (e Error) Error() string {
	return fmt.Sprintf("%s at line %d, col %d: %s", e.Code, e.Ref.Line+1, e.Ref.Column+1, e.Msg)
}

// stateFn is the prototype for function of the lexer state machine. They don't
// take any parameter - data is shared through the object.
type stateFn func() stateFn

// lexer extract the tokens seen in the input.
type lexer struct {
	// source is the data that the lexer is working on.
	source *Source
	// scan is obtained from the source object and provides all the valid rune in
	// the input.
	scan <-chan rune
	// next contains the next rune that would be added to current token with
	// avance().
	next rune
	// nextCol is the rune index on the current line. Ignores invalid byte
	// sequences; 0-indexed.
	nextCol int
	// current line number of the rune in 'next'. 0 indexed.
	line int
	// token is the token being built - it will be sent to the channel once
	// accept() is called.
	token   *tokenInfo
	tokens  chan tokenInfo
	errors  []string
	program Value
}

// peek looks one rune ahead in the input but does not advance the current
// pointer. It should never return RuneError.
func (l *lexer) peek() rune {
	return l.next
}

// advance moves the current position by one rune. Returns the rune encountered
// or 0 if there is nothing remaining.
func (l *lexer) advance() rune {
	r := l.next
	l.token.text = append(l.token.text, r)
	l.nextCol++
	if r == '\n' {
		l.line++
		l.nextCol = 0
	}
	l.next = <-l.scan
	return r
}

func (l *lexer) accept() tokenInfo {
	t := l.token
	l.token = &tokenInfo{
		ref: &SourceRef{
			Source: l.source,
			Line:   l.line,
			Column: l.nextCol,
		},
	}
	return *t
}

// stateIdentifier parses arbitrary strings & numbers.
func (l *lexer) stateIdentifier() (next stateFn) {
	for next == nil {
		r := l.peek()
		switch {
		case r == '(':
			next = l.stateOpen
		case r == ')':
			next = l.stateClose
		case r == ';':
			next = l.stateComment
		case r == '"':
			next = l.stateString
		case unicode.IsSpace(r):
			next = l.stateSpace
		case r == 0: // rune is 0 when scan is finished.
			next = l.stateEnd
		default:
			l.advance()
		}
	}

	token := l.accept()

	// Ignore empty transition - they're just a parsing artifact.
	if len(token.text) == 0 {
		return
	}

	token.id = tokIdentifier
	// Use []rune for identifiers?
	token.value = string(token.text)

	// Check the first rune to determine whether it is just an arbitrary
	// identifier or a number. Anything starting with [+-.]?[0-9] is considered a
	// number.
	if (len(token.text) > 1 && (token.text[0] == '+' || token.text[0] == '-' || token.text[0] == '.') && unicode.IsDigit(token.text[1])) || unicode.IsDigit(token.text[0]) {
		// Try first to parse it as an integer and if it does not work, try as a
		// float. This is ugly and number management should probably be rewritten.
		token.id = tokInteger
		i, err := strconv.ParseInt(string(token.text), 10, 64)
		if err != nil {
			f, err := strconv.ParseFloat(string(token.text), 64)
			if err != nil {
				panic(err) // TODO
			}
			token.id = tokFloat
			token.value = f
			l.tokens <- token
			return
		}
		token.value = i
	}
	l.tokens <- token
	return
}

func (l *lexer) stateOpen() stateFn {
	l.advance()
	token := l.accept()
	// TODO: check that it is the right character and fail otherwise.
	token.id = tokOpen
	l.tokens <- token
	return l.stateIdentifier
}

func (l *lexer) stateClose() stateFn {
	l.advance()
	token := l.accept()
	// TODO: check that it is the right character and fail otherwise.
	token.id = tokClose
	l.tokens <- token
	return l.stateIdentifier
}

func (l *lexer) stateArray() stateFn {
	l.advance()
	token := l.accept()
	// TODO: check that it is the right character and fail otherwise.
	token.id = tokArray
	l.tokens <- token
	return l.stateIdentifier
}

func (l *lexer) stateSpace() stateFn {
	for unicode.IsSpace(l.peek()) {
		l.advance()
	}
	// We don't emit anything for spaces.
	l.accept()
	return l.stateIdentifier
}

func (l *lexer) stateComment() stateFn {
	for l.peek() != '\n' && l.peek() != 0 {
		l.advance()
	}
	// We don't emit anything for comments.
	l.accept()
	return l.stateIdentifier
}

func (l *lexer) stateString() stateFn {
	// Get the opening array.
	l.advance()
	s := ""
	for l.peek() != '"' {
		r := l.advance()

		if r == '\\' {
			// Just get the character after the backslash - good enough to catch
			// escape arrays.
			r = l.advance()
		}

		if r == 0 {
			// TODO - generate an error
			break
		}
		s += string(r)
	}

	// Get the last array
	l.advance()
	token := l.accept()
	token.id = tokString
	token.value = s
	l.tokens <- token
	return l.stateIdentifier
}

func (l *lexer) stateEnd() stateFn {
	return nil
}

func (l *lexer) run() {
	for s := l.stateIdentifier(); s != nil; {
		s = s()
	}
	close(l.tokens)
}

func (l *lexer) Next() Token {
	token, ok := <-l.tokens
	if !ok {
		token = tokenInfo{
			id: tokEOF,
			ref: &SourceRef{
				Source: l.source,
				Line:   l.line,
				Column: l.nextCol,
			},
		}
	}
	return token
}

func newLexer(src *Source) *lexer {
	l := &lexer{
		source: src,
		scan:   src.Scan(),
		token:  &tokenInfo{},
		// As we are starting with a fake advance, we must make sure that indexing
		// stays correct.
		nextCol: -1,
		tokens:  make(chan tokenInfo),
	}
	// Do an initial advance/accept to get the first character into 'next' and
	// make sure than the current token is properly initialized.
	l.advance()
	l.accept()
	go l.run()
	return l
}
