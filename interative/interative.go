package interative

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
	"github.com/rumlang/rum/parser"
	rumRuntime "github.com/rumlang/rum/runtime"
)

const (
	defaultPrompt = ">>> "
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
		Prompt:              defaultPrompt,
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
	ctx.Set("exit", parser.NewAny(func() {
		os.Exit(0)
	}, nil))

	// log.Println(l.Stderr())
	var src string
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt || err == io.EOF {
			if (line == "" && src == "") || err == io.EOF {
				break
			}
			src = ""
			l.SetPrompt(defaultPrompt)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		src += line + "\n"
		nListEnd := strings.Count(src, "(") - strings.Count(src, ")")
		if src != "" {
			l.SetPrompt(fmt.Sprintf("--> %v", strings.Repeat("  ", nListEnd)))
		}
		if nListEnd != 0 {
			continue
		}
		l.SetPrompt(defaultPrompt)

		srcAux := src
		src = ""
		// Parsing
		var out parser.Value
		root, err := parser.Parse(parser.NewSource(srcAux))
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
