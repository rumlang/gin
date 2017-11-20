package interative

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
	"github.com/gin-lang/gin/parser"
	ginRuntime "github.com/gin-lang/gin/runtime"
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
	l, err := readline.NewEx(&readline.Config{
		Prompt:              ">>> ",
		HistoryFile:         "~/.gin_history",
		AutoComplete:        readline.NewPrefixCompleter(),
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return
	}
	defer l.Close()

	// Prepare runtime environment
	ctx := ginRuntime.NewContext(nil)
	ctx.Set("exit", parser.NewAny(func() {
		os.Exit(0)
	}, nil))

	// log.Println(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parsing
		var out parser.Value
		root, err := parser.Parse(parser.NewSource(line))
		if err == nil {
			// Executing
			out, err = ctx.TryEval(root)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		v := out.Value()
		fmt.Println(v)
	}
	return
}
