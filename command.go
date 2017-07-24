package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	myApp = "irony-server"
)

type cliCommandReader struct {
	args []string
}

type stdinCommandReader struct {
}

type CommandDef struct {
	Name string
	Desc string
	Run  func(*Irony, []string) error
}

type commandError struct {
	err string
}

func (c *commandError) Error() string {
	return c.err
}

var Commands []*CommandDef

func initCommands() {
	Commands = []*CommandDef{
		&CommandDef{
			"help",
			"show this message",
			cmdHelp,
		},
		&CommandDef{
			"candidates",
			"print completion candidates (require previous complete)",
			cmdCandidates,
		},
		&CommandDef{
			"complete",
			"FILE LINE COL [-- [COMPILE_OPTIONS...]] - perform code completion at a give location",
			cmdComplete,
		},
		&CommandDef{
			"completion-diagnostics",
			"print the diagnostics generated  during complete",
			nil,
		},
		&CommandDef{
			"diagnostics",
			"print the diagnostics of the last parse",
			cmdDiagnostics,
		},
		&CommandDef{
			"exit",
			"exit interactive mode, print nothing",
			cmdExit,
		},
		&CommandDef{
			"get-compile-options",
			"BUILD_DIR FILE - get compile options for FILE from JSON database in PROJECT_ROOT",
			cmdGetCompileOptions,
		},
		&CommandDef{
			"get-type",
			"LINE COL -get type of symbol at a given location",
			cmdGetType,
		},
		&CommandDef{
			"parse",
			"FILE [-- [COMPILE_OPTIONS...]] - parse the given file",
			cmdParse,
		},
		&CommandDef{
			"reset-unsaved",
			"FILE - reset FILE, its content is up to date",
			cmdResetUnsaved,
		},
		&CommandDef{
			"set-debug",
			"[on/off] - enable or disable verbose logging",
			cmdSetDebug,
		},
		&CommandDef{
			"set-unsaved",
			"FILE UNSAVE - tell irony-server that UNSAVED contains the effective content of FILE",
			cmdSetUnsaved,
		},
	}
}

func printHelp() {
	usageMsg := fmt.Sprintf(
		`usage: %s [OPTIONS...] [COMMAND] [ARGS...]

Options:
  -v, --version
  -h, --help
  -i, --interactive
  -d, --debug
  --log-file PATH
  Commands:`, myApp)
	fmt.Println(usageMsg)
	for _, cmd := range Commands {
		fmt.Printf("%-25s %s\n", cmd.Name, cmd.Desc)
	}

}

func parseUint(arg string) (uint32, error) {
	v, err := strconv.ParseUint(arg, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

func cmdHelp(*Irony, []string) error {
	printHelp()
	return nil
}

func fixupFileName(filename string) string {
	if filename == "-" {
		filename = getTempFilePath()
		logDebug("Convert - to %s\n", filename)
	}
	return filename
}

func readCompileOptions(args []string) []string {
	var i int
	for i := 0; i < len(args); i += 1 {
		if args[i] == "--" {
			break
		}
	}
	if i >= len(args) {
		return nil
	}
	return args[i+1:]
}

func dumpFlags(info string, file string, flags []string) {
	var s string
	for _, flag := range flags {
		s += fmt.Sprintf("`%s` ", flag)
	}
	logDebug("%s: file: %s, flags: %s\n", info, file, s)
}

func cmdExit(*Irony, []string) error {
	os.Exit(0)
	return nil
}

func cmdGetCompileOptions(ir *Irony, args []string) error {
	if len(args) != 3 {
		return &commandError{"Invalid arguments number"}
	}
	buildDir, file := args[1], fixupFileName(args[2])
	ir.GetCompileOptions(buildDir, file)
	return nil
}

func cmdParse(ir *Irony, args []string) error {
	if len(args) < 2 {
		return &commandError{"Invalid argument number"}
	}
	file := fixupFileName(args[1])
	flags := readCompileOptions(args[2:])
	dumpFlags("parse", file, flags)
	ir.Parse(file, flags)
	return nil
}

func cmdResetUnsaved(ir *Irony, args []string) error {
	if len(args) != 2 {
		return &commandError{"Invalid argument number"}
	}
	file := fixupFileName(args[1])
	ir.ResetUnsaved(file)
	return nil
}

func cmdSetUnsaved(ir *Irony, args []string) error {
	if len(args) != 3 {
		return &commandError{"Invalid argument number"}
	}
	file, unsaved := fixupFileName(args[1]), args[2]
	ir.SetUnsaved(file, unsaved)
	return nil
}

func cmdDiagnostics(ir *Irony, args []string) error {
	ir.Diagnostics()
	return nil
}

func cmdComplete(ir *Irony, args []string) error {
	if len(args) < 4 {
		return &commandError{"Invalid argument number"}
	}
	file := fixupFileName(args[1])
	line, err := strconv.ParseUint(args[2], 0, 32)
	if err != nil {
		return &commandError{"Line isn't a integer"}
	}
	column, err := strconv.ParseUint(args[3], 0, 32)
	if err != nil {
		return &commandError{"Column isn't a integer"}
	}
	flags := readCompileOptions(args[4:])
	dumpFlags("complete", file, flags)
	ir.Complete(file, uint32(line), uint32(column), flags)
	return nil
}

func cmdCandidates(ir *Irony, args []string) error {
	ir.Candidates()
	return nil
}

func cmdGetType(ir *Irony, args []string) error {
	if len(args) < 3 {
		return &commandError{"Invalid argument number"}
	}
	line, err := strconv.ParseUint(args[1], 0, 32)
	if err != nil {
		return &commandError{"Invalid line number"}
	}
	col, err := strconv.ParseUint(args[2], 0, 32)
	if err != nil {
		return &commandError{"Invalid col number"}
	}
	ir.GetType(uint32(line), uint32(col))

	return nil
}

func cmdSetDebug(ir *Irony, args []string) error {
	if len(args) < 2 {
		return &commandError{"Invalid argument number"}
	}
	value := strings.ToLower(args[1])
	isOn := false
	if value == "on" {
		isOn = true
	}
	setDebug(isOn)
	return nil
}
