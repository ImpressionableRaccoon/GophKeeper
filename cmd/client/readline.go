package main

import (
	"io"

	"github.com/chzyer/readline"
)

const defaultPrompt = "\033[31m»\033[0m "

const versionTemplate = `Build info:

Version: %s
Date: %s
Commit: %s
`

var (
	l    *readline.Instance
	line string
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("help"),
	readline.PcItem("get"),
	readline.PcItem("add"),
	readline.PcItem("all"),
	readline.PcItem("delete"),
	readline.PcItem("update"),
	readline.PcItem("version"),
)

var readlineCfg = &readline.Config{
	Prompt:          defaultPrompt,
	AutoComplete:    completer,
	InterruptPrompt: "^C",
	EOFPrompt:       "exit",

	HistorySearchFold:   true,
	FuncFilterInputRune: filterInput,
}

func usage(w io.Writer) {
	_, _ = io.WriteString(w, "commands:\n")
	_, _ = io.WriteString(w, completer.Tree("    "))
}

func filterInput(r rune) (rune, bool) {
	if r == readline.CharCtrlZ {
		return r, false
	}
	return r, true
}
