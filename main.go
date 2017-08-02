package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
)

var ClangHeaderDir string

func showVersion() {
	fmt.Printf("%s version %s\n", myApp, GetVersion())
	fmt.Println(GetClangVersion())
}

func init() {
	initLogger()
	initCommands()
}

func release() {
	if e := recover(); e != nil {
		logInfo("%s: %s\n", e, debug.Stack())
	}
	releaseLogger()
}

func exitError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func main() {
	defer release()

	argc := len(os.Args)
	if len(os.Args) <= 1 {
		printHelp()
		return
	}
	ironyApp := NewIrony()
	i := 1
	var interactive = false
	for i < argc {
		arg := os.Args[i]
		if arg[0] != '-' {
			break
		}

		if arg == "-h" || arg == "-help" {
			printHelp()
			return
		} else if arg == "--version" || arg == "-v" {
			showVersion()
			return
		} else if arg == "--debug" || arg == "-d" {
			setDebug(true)
		} else if arg == "-i" || arg == "--interactive" {
			interactive = true
		} else if arg == "--log-file" && (i+1) < argc {
			i += 1
			setupLogger(os.Args[i])
		} else {
			exitError("Error: invalid option %s\n", arg)
			return
		}
		i += 1
	}
	logInfo("Builtin dir: %s\n", ClangHeaderDir)
	var nextCmdFunc func() []string
	if interactive {
		nextCmdFunc = getCmdFromStdin
	} else {
		nextCmdFunc = getCmdFromCliFunc(os.Args[i:])
	}
	runCommands(ironyApp, nextCmdFunc)
}

func getCmdFromCliFunc(args []string) func() []string {
	f := func() []string {
		if args != nil {
			curArgs := args
			args = nil
			return curArgs
		}
		return nil
	}
	return f
}

func quoteParse(line string) ([]string, error) {
	args := []string{}
	buf := ""
	var escaped, doubleQuoted, singleQuoted bool

	got := false

	addArgs := func() {
		args = append(args, buf)
		got = false
		buf = ""
	}

	isSpace := func(r rune) bool {
		switch r {
		case ' ', '\t', '\r', '\n':
			return true
		}
		return false
	}

	for _, r := range line {
		if escaped {
			buf += string(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if isSpace(r) {
			if singleQuoted || doubleQuoted {
				buf += string(r)
			} else if got {
				addArgs()
			}
			continue
		}

		switch r {
		case '"':
			if !singleQuoted {
				doubleQuoted = !doubleQuoted
				if !doubleQuoted {
					addArgs()
				}
				continue
			}
		case '\'':
			if !doubleQuoted {
				singleQuoted = !singleQuoted
				if !singleQuoted {
					addArgs()
				}
				continue
			}
		}
		got = true
		buf += string(r)
	}

	if got {
		args = append(args, buf)
	}

	if escaped || singleQuoted || doubleQuoted {
		return nil, errors.New("invalid command line string")
	}

	return args, nil
}

func getCmdFromStdin() []string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		text := scanner.Text()
		args, err := quoteParse(text)
		if err == nil {
			logDebug("Get cmd %s\n", args)
			return args
		}
		logInfo("Invalid input %s", scanner.Text())
		return nil
	}
	return nil
}

func runCommands(irony *Irony, nextCmd func() []string) {
	cmdMap := make(map[string]*CommandDef)
	for _, cmd := range Commands {
		cmdMap[cmd.Name] = cmd
	}
	for {
		var cmd *CommandDef
		ok := false
		cmdWords := nextCmd()
		if cmdWords == nil {
			return
		}
		if cmd, ok = cmdMap[cmdWords[0]]; !ok {
			logInfo("Invalid command '%s'\n", cmdWords[0])
			return
		}
		if err := cmd.Run(irony, cmdWords); err != nil {
			logInfo("Run [%s]: %s\n", cmd.Name, err)
			return
		}
		fmt.Printf("\n;;EOT\n")
	}
}
