package runtime

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/rumlang/rum/parser"
)

const (
	// ErrPanic is an error caused by a panic within the interpreter.
	ErrPanic = iota
	// ErrUnknownVariable is raised when trying to resolve an unknown symbol.
	ErrUnknownVariable
)

// ErrorCode type to parser errors
type ErrorCode int

func (c ErrorCode) String() string {
	switch c {
	case ErrPanic:
		return "Panic"
	case ErrUnknownVariable:
		return "UnknownVariable"
	default:
		return fmt.Sprintf("Unknown[%d]", c)
	}
}

// Error is sent through panic when something went wrong during the execution.
type Error struct {
	Code ErrorCode
	Msg  string
	// Stack is the glop call stack corresponding to where the error was raised.
	// First value is inner-most eval'ed value.
	Stack []parser.Value

	PanicRecovered interface{}
	PanicStack     []byte
}

func (e *Error) String() string {
	out := fmt.Sprintf("runtime error: %s[%d] - %s\n", e.Code, e.Code, e.Msg)

	var (
		triggeredAt = "  # triggered at:\n"
		calledFrom  = "  # called from:\n"
		msg         = &triggeredAt
	)
	for _, v := range e.Stack {
		out += *msg
		out += v.Ref().Context("    ")
		msg = &calledFrom
	}

	if e.Code == ErrPanic {
		out += "  # Interpreter trace:\n"
		for _, line := range strings.Split(string(e.PanicStack), "\n") {
			out += fmt.Sprintf("    %s\n", line)
		}
	}
	return out
}

func (e *Error) Error() string {
	return e.String()
}

// Internal is the type used to recognized internal functions (for which
// arguments are not evaluated automatically) from regular functions.
type Internal func(*Context, ...parser.Value) parser.Value

// Context contains details about the current execution frame.
type Context struct {
	parent       *Context
	env          map[parser.Identifier]parser.Value
	typeRegistry map[string]reflect.Type
}

// Get returns the content of the specified variable. It will automatically
// look up parent context if needed. Generate a panic with an Error object if
// the specified variable does not exists.
func (c *Context) Get(id parser.Identifier) parser.Value {
	v, ok := c.env[id]
	if !ok {
		if c.parent != nil {
			return c.parent.Get(id)
		}
		panic(&Error{
			Code: ErrUnknownVariable,
			Msg:  fmt.Sprintf("%q does not exist", string(id)),
		})
	}
	return v
}

// Set an iten in parser function map
func (c *Context) Set(id parser.Identifier, v parser.Value) parser.Value {
	_, ok := c.env[id]
	if !ok {
		c.env[id] = v
		return v
	}
	panic(fmt.Sprintf("Variable called %s just has a value in that scope", id))
}

// SetFn an function in parser function map
func (c *Context) SetFn(id parser.Identifier, v interface{}, adapters ...Adapter) {
	f := func(values ...interface{}) interface{} {
		args := values
		var err error
		for _, adapter := range adapters {
			args, err = adapter(args...)
			if err != nil {
				panic(fmt.Sprint("Error in adapter", values[0], err))
			}
		}

		vargs := []reflect.Value{}
		for _, arg := range args {
			vargs = append(vargs, reflect.ValueOf(arg))
		}

		result := reflect.ValueOf(v).Call(vargs)
		return result[0].Interface()
	}

	c.env[id] = parser.NewAny(f, nil)
}

//RegisterType register an new type in runtime. A nil or zero typed value must be the parameter
func (c *Context) RegisterType(typedNil interface{}) {
	t := reflect.TypeOf(typedNil).Elem()
	c.typeRegistry[t.PkgPath()+"."+t.Name()] = t
}

// dispatch takes the provided value, evaluates it based on the current content
// of the context and returns the result. All errors are sent through panics.
func (c *Context) dispatch(input parser.Value) (parser.Value, error) {
	switch data := input.Value().(type) {
	case []parser.Value:
		if len(data) <= 0 {
			return parser.NewAny(nil, nil), nil
		}

		fn, err := c.eval(data[0])
		if err != nil {
			return nil, err
		}

		if internal, ok := fn.Value().(Internal); ok {
			return internal(c, data[1:]...), nil
		}

		var args []reflect.Value
		for _, child := range data[1:] {
			v, err := c.eval(child)
			if err != nil {
				return nil, err
			}
			data := v.Value()

			if data == nil {
				// Do a ValueOf of the pointer to get the element afterward - that
				// circumvent the special value with nil which is otherwise translated to
				// an invalid element.
				args = append(args, reflect.ValueOf(&data).Elem())
				continue
			}
			args = append(args, reflect.ValueOf(data))

		}
		result := reflect.ValueOf(fn.Value()).Call(args)
		if len(result) == 0 {
			return parser.NewAny(nil, nil), nil
		}
		if len(result) == 1 {
			return parser.NewAny(result[0].Interface(), nil), nil
		}
		panic("Multiple arguments unsupported")
	case parser.Identifier:
		return c.Get(data), nil
	default:
		// If it is neither an identifier or a list, just return the value.
		return input, nil
	}
}

// eval evaluates the provided value. It makes sure to catch any panic and
// create an error (type *Error) with full stack trace when that happens.
func (c *Context) eval(input parser.Value) (parser.Value, error) {
	var recov interface{}
	var result parser.Value
	var stack []byte
	var err error
	func() {
		defer func() {
			const size = 16384
			stack = make([]byte, size)
			// Unfortunately, that also catch itself, adding noise to the trace.
			stack = stack[:runtime.Stack(stack, false)]
			recov = recover()
		}()
		result, err = c.dispatch(input)
	}()

	if recov != nil {
		err = &Error{
			Code:           ErrPanic,
			Msg:            fmt.Sprintf("%v", recov),
			PanicRecovered: recov,
			PanicStack:     stack,
		}
		if details, ok := recov.(*Error); ok {
			err = details
		}
	}

	if err != nil {
		err.(*Error).Stack = append(err.(*Error).Stack, input)
	}

	return result, err
}

// TryEval evaluates the provided value and catch any panic, return an error
// instead.
func (c *Context) TryEval(input parser.Value) (parser.Value, error) {
	return c.eval(input)
}

// MustEval evaluates the provided value, generatic panics when something bad
// happens. Panics will be *Error instances, containing the call stack.
func (c *Context) MustEval(input parser.Value) parser.Value {
	v, err := c.eval(input)
	if err != nil {
		panic(err)
	}
	return v
}

// NewContext create new runtime context
// instance and load default parser funcrions
func NewContext(parent *Context) *Context {
	c := &Context{
		parent:       parent,
		env:          make(map[parser.Identifier]parser.Value),
		typeRegistry: make(map[string]reflect.Type),
	}

	if parent == nil {
		defaults := map[parser.Identifier]interface{}{
			"package": Package,
			"array":   Internal(Array),
			"let":     Internal(Let),
			"if":      Internal(If),
			"def":     Internal(Def),
			"lambda":  Internal(Lambda),
			"eval":    Internal(Eval),
			"for":     Internal(For),
			"coerce":  Internal(Coerce),
			".":       Internal(Invoke),
			"import":  Internal(Import),
			"panic":   Panic,
			"len":     Length,
			"print":   Print,
			"println": Println,
			"sprintf": Sprintf,
			"type":    Type,
			"true":    true,
			"false":   false,
			"+":       OpAdd,
			"-":       OpSub,
			"*":       OpMul,
			"**":      OpPow,
			"==":      OpEqual,
			"!=":      OpNotEqual,
			"<":       OpLess,
			"<=":      OpLessEqual,
			">":       OpGreater,
			">=":      OpGreaterEqual,
		}

		for name, value := range defaults {
			c.env[name] = parser.NewAny(value, nil)
		}
	}

	return c
}

// Array define single or multiple-dimension arrays using the make-array function
func Array(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) != 1 {
		panic("Invalid number of arguments for array")
	}
	return args[0]
}

// Package implements the package reserved word.
func Package(name string, values ...interface{}) interface{} {
	if len(values) == 0 {
		return nil
	}
	return values[len(values)-1]
}

// Import implements the import package feature.
func Import(ctx *Context, args ...parser.Value) (v parser.Value) {
	if len(args) == 0 {
		panic("Invalid arguments")
	}

	for key := range args {
		var packageID parser.Identifier
		var packageName parser.Value
		var packageNameStr string
		input := args[key]
		switch data := input.Value().(type) {
		case []parser.Value:
			packageID = data[0].Value().(parser.Identifier)
			packageName = data[1]
			packageNameStr = data[1].Value().(string)
		case parser.Identifier:
			packageID = data
			packageName = input
			packageNameStr = data.String()
		default:
			panic(fmt.Sprintf("package %s not found", Type(args[key].Value())))
		}
		loadStdLib(packageNameStr, ctx, packageID)
		v = ctx.Set(packageID, packageName)
	}
	return v
}

// Let implements the let reserved word.
func Let(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) != 2 {
		panic("Invalid arguments")
	}

	id, ok := args[0].Value().(parser.Identifier)
	if !ok {
		panic("TODO")
	}
	return ctx.Set(id, ctx.MustEval(args[1]))
}

// If implements the 'if' builtin function. It has to be an Internal interface
// - otherwise, both true & false expressions would have been already
// evaluated.
func If(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) < 2 || len(args) > 3 {
		panic("Invalid arguments")
	}

	cond := ctx.MustEval(args[0]).Value().(bool)
	if cond {
		return ctx.MustEval(args[1])
	}

	if len(args) <= 2 {
		return parser.NewAny(nil, nil)
	}

	return ctx.MustEval(args[2])
}

// Def is a group of statements that together perform a task.
func Def(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) != 3 {
		panic("Invalid arguments")
	}

	id, ok := args[0].Value().(parser.Identifier)
	if !ok {
		panic("TODO")
	}

	params, ok := args[1].Value().([]parser.Value)
	if !ok {
		panic("TODO")
	}
	names := []parser.Identifier{}
	for _, v := range params {
		nameid, ok := v.Value().(parser.Identifier)
		if !ok {
			panic("TODO")
		}
		names = append(names, nameid)
	}
	implValue := args[2]
	impl := func(implCtx *Context, args ...parser.Value) parser.Value {
		if len(args) != len(names) {
			panic("TODO")
		}
		nested := NewContext(implCtx)
		for i, name := range names {
			nested.Set(name, implCtx.MustEval(args[i]))
		}
		return nested.MustEval(implValue)
	}
	ctx.env[id] = parser.NewAny(Internal(impl), nil)
	return ctx.env[id]
}

// Lambda anonymous functions that are evaluated only when they are encountered in the program
func Lambda(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) != 2 {
		panic("Invalid arguments")
	}

	params, ok := args[0].Value().([]parser.Value)
	if !ok {
		panic("TODO")
	}
	names := []parser.Identifier{}
	for _, v := range params {
		id, ok := v.Value().(parser.Identifier)
		if !ok {
			panic("TODO")
		}
		names = append(names, id)
	}
	implValue := args[1]
	impl := func(implCtx *Context, args ...parser.Value) parser.Value {
		if len(args) != len(names) {
			panic("TODO")
		}
		nested := NewContext(implCtx)
		for i, name := range names {
			nested.Set(name, implCtx.MustEval(args[i]))
		}
		return nested.MustEval(implValue)
	}

	return parser.NewAny(Internal(impl), nil)
}

// Type implements the type function.
func Type(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

// Length implements the len function.
func Length(elt []parser.Value) int64 {
	return int64(len(elt))
}

// Panic implements the panic function.
func Panic(v interface{}) {
	panic(v)
}

// Print implements the print function.
func Print(args ...interface{}) {
	for i, v := range args {
		if i != 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%v", v)
	}
}

// Println implements the print function.
func Println(args ...interface{}) {
	Print(args)
	fmt.Printf("\n")
}

// Sprintf implements the sprintf function.
func Sprintf(args ...interface{}) string {
	format := args[0].(string)
	return fmt.Sprintf(format, args[1:]...)
}

// Eval implements the eval function.
func Eval(ctx *Context, raw ...parser.Value) parser.Value {

	// Do the normal evaluation of each parameter first - as we're in an internal
	// function, this is not done automatically.
	var args []parser.Value
	for _, arg := range raw {
		args = append(args, ctx.MustEval(arg))
	}

	// And now, actually do the eval work.
	var result parser.Value
	for _, arg := range args {
		result = ctx.MustEval(arg)
	}
	return result
}

// For implements for loop
func For(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) < 2 {
		panic("Invalid arguments")
	}
	_, ok := args[0].Value().(parser.Identifier)
	if !ok {
		panic("TODO")
	}
	params, ok := args[1].Value().([]parser.Value)
	if !ok {
		panic("TODO")
	}
	for _, s := range params[1:] {
		vs, ok := s.Value().([]parser.Value)
		if !ok {
			panic("TODO")
		}
		for _, v := range vs {
			input := parser.NewAny([]parser.Value{args[0], v}, nil)
			_, err := ctx.dispatch(input)
			if err != nil {
				err.(*Error).Stack = append(err.(*Error).Stack, input)
				return parser.NewAny(nil, nil)
			}
		}
	}
	return parser.NewAny(nil, nil)
}

//Dump the context content
func (c *Context) Dump() {
	for id, val := range c.env {
		fmt.Println(id, val)
	}
}

//Invoke call a method from native value
func Invoke(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) < 2 {
		panic("Invalid arguments")
	}

	obj := ctx.MustEval(args[0]).Value()
	descriptor := args[1].String()

	method := reflect.ValueOf(obj).MethodByName(descriptor)
	if method.IsValid() {
		vargs := []reflect.Value{}
		for _, arg := range args[2:] {
			vargs = append(vargs, reflect.ValueOf(ctx.MustEval(arg).Value()))
		}

		result := method.Call(vargs)
		return parser.NewAny(result[0].Interface(), nil)
	}

	if reflect.ValueOf(obj).Type().Kind() == reflect.Ptr {
		field := reflect.Indirect(reflect.ValueOf(obj)).FieldByName(descriptor)
		if field.IsValid() {
			return parser.NewAny(field.Interface(), nil)
		}
	} else {
		field := reflect.ValueOf(obj).FieldByName(descriptor)
		if field.IsValid() {
			return parser.NewAny(field.Interface(), nil)
		}
	}

	panic("Method or field not found: '" + descriptor +
		"' in type: " + reflect.ValueOf(obj).Type().String())
}

//Coerce returns the value v converted to type t
func Coerce(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) < 2 {
		panic("Invalid arguments")
	}

	name := args[0].String()
	name = name[5 : len(name)-1]
	obj := reflect.ValueOf(ctx.MustEval(args[1]).Value())

	ret := obj.Convert(ctx.typeRegistry[name])

	return parser.NewAny(ret, nil)
}
