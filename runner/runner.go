package runner

import (
	"github.com/palats/glop/nodes"
	"github.com/palats/glop/parser"
)

type Context struct {
	env map[string]interface{}
}

func (c *Context) Get(s string) interface{} {
	return c.env[s]
}

func (c *Context) Set(s string, v interface{}) interface{} {
	c.env[s] = v
	return v
}

func NewContext() *Context {
	c := &Context{
		env: make(map[string]interface{}),
	}

	c.env["+"] = OpAdd
	c.env["begin"] = Begin
	c.env["quote"] = nodes.Internal(Quote)
	c.env["define"] = nodes.Internal(Define)
	c.env["set!"] = nodes.Internal(Define)
	// TODO: if
	// TODO: set!
	// TODO: define
	// TODO: lambda

	return c
}

func OpAdd(values ...int64) int64 {
	var total int64
	for _, v := range values {
		total += v
	}
	return total
}

func Quote(ctx nodes.Context, args ...nodes.Node) interface{} {
	if len(args) != 1 {
		panic("Invalid number of arguments for quote")
	}
	return args[0]
}

func Begin(values ...interface{}) interface{} {
	if len(values) == 0 {
		return nil
	}
	return values[len(values)-1]
}

func Define(ctx nodes.Context, args ...nodes.Node) interface{} {
	if len(args) != 2 {
		panic("Invalid arguments")
	}
	s := args[0].(*nodes.Identifier).Value()
	return ctx.Set(s, args[1].Eval(ctx))
}

func ParseEval(input string) interface{} {
	return parser.Parse(input).Eval(NewContext())
}
