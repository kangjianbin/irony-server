package main

import (
	"bufio"
	"fmt"
	"github.com/go-clang/v3.9/clang"
	"github.com/mattn/go-shellwords"
	"os"
)

var ClangHeaderDir string

func showVersion() {
	fmt.Printf("%s version %s\n", myApp, GetVersion())
	fmt.Println(clang.GetClangVersion())
}

func init() {
	initLogger()
	initCommands()
}

func release() {
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

func getCmdFromStdin() []string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		args, err := shellwords.Parse(scanner.Text())
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
