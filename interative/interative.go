package interative

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
	"github.com/rumlang/rum/libs"
	"github.com/rumlang/rum/parser"
	rumRuntime "github.com/rumlang/rum/runtime"
)

// ExpandFilename replace '~/' with the user home directory.
func ExpandFilename(s string, u *user.User) (string, error) {
	if strings.HasPrefix(s, "~/") {
		if u == nil {
			var err error
			u, err = user.Current()
			if err != nil {
				return "", err
			}
		}
		s = u.HomeDir + s[1:]
	}
	return s, nil
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// REPL starts a full interpreter, accepting glop code on its prompt.
func REPL() (err error) {

	usr, err := user.Current()
	if err != nil {
		return
	}
	l, err := readline.NewEx(&readline.Config{
		Prompt:              ">>> ",
		HistoryFile:         usr.HomeDir + "/.rum_history",
		AutoComplete:        readline.NewPrefixCompleter(),
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return
	}
	defer func(l *readline.Instance) {
		errReadlineClose := l.Close()
		if errReadlineClose != nil {
			log.Println(errReadlineClose)
			if err == nil {
				err = errReadlineClose
			}
		}
	}(l)

	// Prepare runtime environment
	ctx := rumRuntime.NewContext(nil)
	libs.LoadStdLib(ctx)
	ctx.Set("exit", parser.NewAny(func() {
		os.Exit(0)
	}, nil))

	// log.Println(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt || err == io.EOF {
			if len(line) == 0 || err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parsing
		var out parser.Value
		root, err := parser.Parse(parser.NewSource(line))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		// Executing
		out, err = ctx.TryEval(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		v := out.Value()
		fmt.Println(v)
	}
	return
}
