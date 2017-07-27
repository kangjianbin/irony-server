package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
)

const (
	myVersion     = "1.0.0"
	MaxCandidates = 20
)

type Irony struct {
	Debug        bool
	cache        *TUCache
	activeTd     *TUData
	fileContent  map[string]string
	curFile      string
	unsavedFiles []UnsavedFile
	actCmplRes   *CodeCompleteResults
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
	echoInfo("nil")
}

func (irony *Irony) resetCache() {
	if irony.actCmplRes != nil {
		irony.actCmplRes.Dispose()
		irony.actCmplRes = nil
	}
	if irony.activeTd != nil {
		irony.activeTd.Dispose()
		irony.activeTd = nil
	}
}

func (irony *Irony) computeUnsaved() {
	irony.unsavedFiles = nil
	for _, unsaved := range irony.unsavedFiles {
		unsaved.Dispose()
	}

	for file, contents := range irony.fileContent {
		unsavedFile := NewUnsavedFile(file, contents)
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

func diagnosticSeverity(diagnostic Diagnostic) string {
	switch diagnostic.Severity() {
	case Diagnostic_Ignored:
		return "ignored"
	case Diagnostic_Note:
		return "note"
	case Diagnostic_Warning:
		return "warning"
	case Diagnostic_Error:
		return "error"
	case Diagnostic_Fatal:
		return "fatal"
	}
	return "unknown"
}

func dumpDiagnostic(diagnostic Diagnostic) {
	var file string
	var line, column, offset uint32
	location := diagnostic.Location()
	if !location.Equal(NewNullLocation()) {
		var cxFile File
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
		opts := DefaultCodeCompleteOptions()
		irony.actCmplRes = td.tu.CodeCompleteAt(file, line, col, irony.unsavedFiles, opts)
		defer td.Dispose()
	}
	if irony.actCmplRes == nil {
		echoError(`complete-error "failed to perform code completion" %s %d %d"`, quote(file), line, col)
		return
	}
	SortCodeCompletionResults(irony.actCmplRes.Results())
	echoSuccess()
}

func getAvaliString(avail AvailabilityKind) string {
	switch avail {
	case Availability_NotAvailable:
		return ""
	case Availability_Available:
		return "available"
	case Availability_Deprecated:
		return "deprecated"
	case Availability_NotAccessible:
		return "not-accessible"
	}
	return ""
}

func dumpCandidate(res CompletionResult) {
	cmplString := res.CompletionString()
	avail := cmplString.Availability()
	if avail == Availability_NotAvailable {
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
		case CompletionChunk_ResultType:
			resultType = chunkText
		case CompletionChunk_TypedText, CompletionChunk_Text:
			fallthrough
		case CompletionChunk_Placeholder, CompletionChunk_Informative:
			fallthrough
		case CompletionChunk_CurrentParameter:
			prototype += chunkText
		case CompletionChunk_LeftParen:
			ch = "("
		case CompletionChunk_RightParen:
			ch = ")"
		case CompletionChunk_LeftBracket:
			ch = "["
		case CompletionChunk_RightBracket:
			ch = "]"
		case CompletionChunk_LeftBrace:
			ch = "{"
		case CompletionChunk_RightBrace:
			ch = "}"
		case CompletionChunk_LeftAngle:
			ch = "<"
		case CompletionChunk_RightAngle:
			ch = ">"
		case CompletionChunk_Comma:
			ch = ", "
		case CompletionChunk_Colon:
			ch = ":"
		case CompletionChunk_SemiColon:
			ch = ";"
		case CompletionChunk_Equal:
			ch = "="
		case CompletionChunk_HorizontalSpace:
			ch = " "
		case CompletionChunk_VerticalSpace:
			ch = "\n"
		case CompletionChunk_Optional:
			//
		}
		if ch != "" {
			prototype += string(ch)
		}
		if typedTextSet {
			if ch != "" {
				postCompCar += ch
			} else if kind == CompletionChunk_Text || kind == CompletionChunk_TypedText {
				postCompCar += chunkText
			} else if kind == CompletionChunk_Placeholder || kind == CompletionChunk_CurrentParameter {
				postCompCdr = append(postCompCdr, len(postCompCar))
				postCompCar += chunkText
				postCompCdr = append(postCompCdr, len(postCompCar))
			}
		}
		if kind == CompletionChunk_TypedText && !typedTextSet {
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
	results := irony.actCmplRes.Results()
	num := len(results)
	// if num > MaxCandidates {
	// 	num = MaxCandidates
	// }

	echoInfo("(")
	for i := 0; i < num; i += 1 {
		dumpCandidate(results[i])
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
	var cxTypes [2]Type
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
