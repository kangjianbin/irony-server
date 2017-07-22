package main

import (
	"fmt"
	"github.com/go-clang/v3.9/clang"
	"io/ioutil"
	"strconv"
)

const (
	myVersion = "1.0.0"
)

type Irony struct {
	Debug        bool
	cache        *TUCache
	activeTd     *TUData
	fileContent  map[string]string
	curFile      string
	unsavedFiles []clang.UnsavedFile
	actCmplRes   *clang.CodeCompleteResults
}

func GetVersion() string {
	return myVersion
}

func NewIrony() *Irony {
	var app = Irony{}
	app.cache = NewTuCache()
	app.fileContent = make(map[string]string)
	return &app
}

func quote(s string) string {
	return strconv.Quote(s)
}

func echoError(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	s = "(error . (" + s + "))"
	logInfo(s)
	fmt.Println(s)
}

func echoSuccess() {
	s := "(success . t)\n"
	logDebug(s)
	fmt.Printf(s)
}

func echoInfo(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	logDebug(s)
	fmt.Printf(s)
}

func (irony *Irony) Dispose() {
	irony.cache.Dispose()
}

func (ir *Irony) GetCompileOptions(buildDir string, file string) {
	err, cd := clang.FromDirectory(buildDir)
	if err == clang.CompilationDatabase_CanNotLoadDatabase {
		echoInfo("nil")
		return
	}
	defer cd.Dispose()
	cc := cd.CompileCommands(file)
	defer cc.Dispose()
	for i := uint32(0); i < cc.Size(); i += 1 {
		cmd := cc.Command(i)
		for j := uint32(0); j < cmd.NumArgs(); j += 1 {
			arg := cmd.Arg(j)
			fmt.Printf("%s ", quote(arg))
		}
		fmt.Println()
		dir := cmd.Directory()
		fmt.Printf("%s\n", quote(dir))
	}
}

func (irony *Irony) resetCache() {
	if irony.activeTd != nil {
		irony.activeTd.Dispose()
		irony.activeTd = nil
	}
	if irony.actCmplRes != nil {
		irony.actCmplRes.Dispose()
		irony.actCmplRes = nil
	}
}

func (irony *Irony) computeUnsaved() {
	irony.unsavedFiles = nil
	for file, contents := range irony.fileContent {
		unsavedFile := clang.NewUnsavedFile(file, contents)
		irony.unsavedFiles = append(irony.unsavedFiles, unsavedFile)
	}
}

func (irony *Irony) SetUnsaved(file string, unsaved string) {
	data, err := ioutil.ReadFile(unsaved)
	if err != nil {
		delete(irony.fileContent, file)
		echoError(`file-read-error "failed to read unsaved buffer" %s %s`, quote(file), quote(unsaved))
	} else {
		irony.fileContent[file] = string(data)
		echoSuccess()
	}
	irony.computeUnsaved()
}

func (irony *Irony) ResetUnsaved(file string) {
	irony.resetCache()
	_, ok := irony.fileContent[file]
	if ok {
		delete(irony.fileContent, file)
		irony.computeUnsaved()
	}
	echoSuccess()
}

func (irony *Irony) Parse(file string, flags []string) {
	irony.resetCache()
	td := irony.cache.Parse(file, flags, irony.unsavedFiles)
	if td == nil {
		echoError(`parse-error "failed to parse file" %s`, quote(file))
		return
	}
	irony.activeTd = td
	logDebug("Parse %s done\n", file)
	echoSuccess()
}

func diagnosticSeverity(diagnostic clang.Diagnostic) string {
	switch diagnostic.Severity() {
	case clang.Diagnostic_Ignored:
		return "ignored"
	case clang.Diagnostic_Note:
		return "note"
	case clang.Diagnostic_Warning:
		return "warning"
	case clang.Diagnostic_Error:
		return "error"
	case clang.Diagnostic_Fatal:
		return "fatal"
	}
	return "unknown"
}

func dumpDiagnostic(diagnostic clang.Diagnostic) {
	var file string
	var line, column, offset uint32
	location := diagnostic.Location()
	if !location.Equal(clang.NewNullLocation()) {
		var cxFile clang.File
		cxFile, line, column, offset = location.ExpansionLocation()
		file = cxFile.Name()
	}
	severity := diagnosticSeverity(diagnostic)
	msg := diagnostic.Spelling()
	echoInfo("(%s %d %d %d %s %s)\n", quote(file), line, column, offset, severity, quote(msg))
}

func (irony *Irony) Diagnostics() {
	var count uint32
	if irony.activeTd == nil {
		logInfo("No active tu\n")
		count = 0
	} else {
		count = irony.activeTd.tu.NumDiagnostics()
	}
	echoInfo("(\n")
	for i := uint32(0); i < count; i += 1 {
		diagnostic := irony.activeTd.tu.Diagnostic(i)
		dumpDiagnostic(diagnostic)
		diagnostic.Dispose()
	}
	echoInfo(")\n")
}

func (irony *Irony) Complete(file string, line, col uint32, flags []string) {
	irony.resetCache()
	td := irony.cache.GenTU(file, flags, irony.unsavedFiles)
	if td != nil {
		opts := clang.DefaultCodeCompleteOptions() & (^uint32(clang.CodeComplete_IncludeCodePatterns))
		irony.actCmplRes = td.tu.CodeCompleteAt(file, line, col, irony.unsavedFiles, opts)
		defer td.Dispose()
	}
	if irony.actCmplRes == nil {
		echoError(`complete-error "failed to perform code completion" %s %d %d"`, quote(file), line, col)
		return
	}
	clang.SortCodeCompletionResults(irony.actCmplRes.Results())
	echoSuccess()
}

func getAvaliString(avail clang.AvailabilityKind) string {
	switch avail {
	case clang.Availability_NotAvailable:
		return ""
	case clang.Availability_Available:
		return "available"
	case clang.Availability_Deprecated:
		return "deprecated"
	case clang.Availability_NotAccessible:
		return "not-accessible"
	}
	return ""
}

func dumpCandidate(res clang.CompletionResult) {
	cmplString := res.CompletionString()
	avail := cmplString.Availability()
	if avail == clang.Availability_NotAvailable {
		return
	}
	priority := cmplString.Priority()
	availString := getAvaliString(avail)
	var typedtext, brief, resultType, prototype, postCompCar string
	var postCompCdr []int
	var annotationStart int
	typedTextSet := false
	for i := uint32(0); i < cmplString.NumChunks(); i += 1 {
		ch := ""
		kind := cmplString.ChunkKind(i)
		chunkText := cmplString.ChunkText(i)
		switch kind {
		case clang.CompletionChunk_ResultType:
			resultType = chunkText
		case clang.CompletionChunk_TypedText, clang.CompletionChunk_Text:
			fallthrough
		case clang.CompletionChunk_Placeholder, clang.CompletionChunk_Informative:
			fallthrough
		case clang.CompletionChunk_CurrentParameter:
			prototype += chunkText
		case clang.CompletionChunk_LeftParen:
			ch = "("
		case clang.CompletionChunk_RightParen:
			ch = ")"
		case clang.CompletionChunk_LeftBracket:
			ch = "["
		case clang.CompletionChunk_RightBracket:
			ch = "]"
		case clang.CompletionChunk_LeftBrace:
			ch = "{"
		case clang.CompletionChunk_RightBrace:
			ch = "}"
		case clang.CompletionChunk_LeftAngle:
			ch = "<"
		case clang.CompletionChunk_RightAngle:
			ch = ">"
		case clang.CompletionChunk_Comma:
			ch = ","
		case clang.CompletionChunk_Colon:
			ch = ":"
		case clang.CompletionChunk_SemiColon:
			ch = ";"
		case clang.CompletionChunk_Equal:
			ch = "="
		case clang.CompletionChunk_HorizontalSpace:
			ch = " "
		case clang.CompletionChunk_VerticalSpace:
			ch = "\n"
		case clang.CompletionChunk_Optional:
			//
		}
		if ch != "" {
			prototype += string(ch)
			if ch == "," {
				prototype += " "
			}
		}
		if typedTextSet {
			if ch != "" {
				postCompCar += ch
				if ch == "," {
					postCompCar += " "
				}
			} else if kind == clang.CompletionChunk_Text || kind == clang.CompletionChunk_TypedText {
				postCompCar += chunkText
			} else if kind == clang.CompletionChunk_Placeholder || kind == clang.CompletionChunk_CurrentParameter {
				postCompCdr = append(postCompCdr, len(postCompCar))
				postCompCar += chunkText
				postCompCdr = append(postCompCdr, len(postCompCar))
			}
		}
		if kind == clang.CompletionChunk_TypedText && !typedTextSet {
			typedtext = chunkText
			typedTextSet = true
			annotationStart = len(prototype)
		}
	}
	s := fmt.Sprintf(`  (%s %d %s %s %s %d (%s`,
		quote(typedtext), priority, quote(resultType), quote(brief),
		quote(prototype), annotationStart, quote(postCompCar))
	for _, v := range postCompCdr {
		s += fmt.Sprintf(" %d", v)
	}
	s += fmt.Sprintf(") %s)\n", availString)
	echoInfo(s)

}

func (irony *Irony) Candidates() {
	if irony.actCmplRes == nil {
		fmt.Printf("nil\n")
		return
	}
	cmpl := irony.actCmplRes
	echoInfo("(")
	for _, res := range cmpl.Results() {
		dumpCandidate(res)
	}
	echoInfo(")")
}

func (irony *Irony) GetType(line, col uint32) {
	if irony.activeTd == nil {
		logInfo("W: get-type -parse wasn't called\n")
		echoInfo("nil")
		return
	}
	tu := irony.activeTd.tu
	cxFile := tu.File(irony.activeTd.file)
	srcLoc := tu.Location(cxFile, line, col)
	cursor := tu.Cursor(srcLoc)
	if cursor.IsNull() {
		echoInfo("nil")
		return
	}

	s := "("
	var cxTypes [2]clang.Type
	cxTypes[0] = cursor.Type()
	cxTypes[1] = cxTypes[0].CanonicalType()
	for _, t := range cxTypes {
		typeDesc := t.Spelling()
		if typeDesc == "" {
			break
		}
		s += quote(typeDesc) + " "
	}
	s += ")"
	echoInfo(s)
}
