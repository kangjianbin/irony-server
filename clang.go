package main

// #cgo LDFLAGS: -lclang
// #include <clang-c/Index.h>
// #include <stdlib.h>
import "C"
import "reflect"
import "unsafe"

func DefaultEditingTranslationUnitOptions() uint32 {
	return uint32(C.clang_defaultEditingTranslationUnitOptions())
}

func DefaultCodeCompleteOptions() uint32 {
	return uint32(C.clang_defaultCodeCompleteOptions())
}

func SortCodeCompletionResults(results []CompletionResult) {
	gos_results := (*reflect.SliceHeader)(unsafe.Pointer(&results))
	cp_results := (*C.CXCompletionResult)(unsafe.Pointer(gos_results.Data))

	C.clang_sortCodeCompletionResults(cp_results, C.uint(len(results)))
}

func NewIndex(excludeDeclarationsFromPCH int32, displayDiagnostics int32) Index {
	return Index{C.clang_createIndex(C.int(excludeDeclarationsFromPCH), C.int(displayDiagnostics))}
}

func (c cxstring) String() string {
	cstr := C.clang_getCString(c.c)
	return C.GoString(cstr)
}

func (c cxstring) Dispose() {
	C.clang_disposeString(c.c)
}

func NewUnsavedFile(filename string, contents string) UnsavedFile {
	return UnsavedFile{
		C.struct_CXUnsavedFile{
			Filename: C.CString(filename),
			Contents: C.CString(contents),
			Length:   C.ulong(len(contents)),
		},
	}
}

func NewNullLocation() SourceLocation {
	return SourceLocation{C.clang_getNullLocation()}
}

func (i Index) Dispose() {
	C.clang_disposeIndex(i.c)
}

func (i Index) ParseTranslationUnit2FullArgv(source string, args []string, unsaved []UnsavedFile, options uint32, outTu *TranslationUnit) ErrorCode {
	cArgs := make([]*C.char, len(args))
	var pArgs **C.char
	if len(cArgs) > 0 {
		pArgs = &cArgs[0]
	}
	for i := range args {
		ci_str := C.CString(args[i])
		defer C.free(unsafe.Pointer(ci_str))
		cArgs[i] = ci_str
	}
	gos_unsavedFiles := (*reflect.SliceHeader)(unsafe.Pointer(&unsaved))
	cp_unsaved := (*C.struct_CXUnsavedFile)(unsafe.Pointer(gos_unsavedFiles.Data))
	c_source := C.CString(source)
	defer C.free(unsafe.Pointer(c_source))

	return ErrorCode(C.clang_parseTranslationUnit2FullArgv(i.c, c_source, pArgs, C.int(len(args)), cp_unsaved, C.uint(len(unsaved)), C.uint(options), &outTu.c))
}

func (tu TranslationUnit) ReparseTranslationUnit(unsavedFiles []UnsavedFile, options uint32) ErrorCode {
	gos_unsavedFiles := (*reflect.SliceHeader)(unsafe.Pointer(&unsavedFiles))
	cp_unsavedFiles := (*C.struct_CXUnsavedFile)(unsafe.Pointer(gos_unsavedFiles.Data))

	ret := C.clang_reparseTranslationUnit(tu.c, C.uint(len(unsavedFiles)), cp_unsavedFiles, C.uint(options))
	return ErrorCode(ret)
}

func (tu TranslationUnit) Dispose() {
	C.clang_disposeTranslationUnit(tu.c)
}

func (tu TranslationUnit) IsValid() bool {
	return tu.c != nil
}

func (tu TranslationUnit) DefaultReparseOptions() uint32 {
	return uint32(C.clang_defaultReparseOptions(tu.c))
}

func (tu TranslationUnit) NumDiagnostics() uint32 {
	return uint32(C.clang_getNumDiagnostics(tu.c))
}

func (tu TranslationUnit) Diagnostic(index uint32) Diagnostic {
	return Diagnostic{C.clang_getDiagnostic(tu.c, C.uint(index))}
}

func (tu TranslationUnit) CodeCompleteAt(completeFilename string, completeLine uint32, completeColumn uint32, unsavedFiles []UnsavedFile, options uint32) *CodeCompleteResults {
	gos_unsavedFiles := (*reflect.SliceHeader)(unsafe.Pointer(&unsavedFiles))
	cp_unsavedFiles := (*C.struct_CXUnsavedFile)(unsafe.Pointer(gos_unsavedFiles.Data))

	c_completeFilename := C.CString(completeFilename)
	defer C.free(unsafe.Pointer(c_completeFilename))

	o := C.clang_codeCompleteAt(tu.c, c_completeFilename, C.uint(completeLine), C.uint(completeColumn), cp_unsavedFiles, C.uint(len(unsavedFiles)), C.uint(options))

	var gop_o *CodeCompleteResults
	if o != nil {
		gop_o = &CodeCompleteResults{o}
	}

	return gop_o
}

func (tu TranslationUnit) File(fileName string) File {
	c_fileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(c_fileName))
	return File{C.clang_getFile(tu.c, c_fileName)}
}

func (tu TranslationUnit) Location(file File, line uint32, column uint32) SourceLocation {
	return SourceLocation{C.clang_getLocation(tu.c, file.c, C.uint(line), C.uint(column))}
}

func (tu TranslationUnit) Cursor(sl SourceLocation) Cursor {
	return Cursor{C.clang_getCursor(tu.c, sl.c)}
}

func (unsaved *UnsavedFile) Dispose() {
	C.free(unsafe.Pointer(unsaved.c.Filename))
	C.free(unsafe.Pointer(unsaved.c.Contents))
}

func (d Diagnostic) Severity() DiagnosticSeverity {
	return DiagnosticSeverity(C.clang_getDiagnosticSeverity(d.c))
}

func (d Diagnostic) Location() SourceLocation {
	return SourceLocation{C.clang_getDiagnosticLocation(d.c)}
}

func (d Diagnostic) Spelling() string {
	o := cxstring{C.clang_getDiagnosticSpelling(d.c)}
	defer o.Dispose()

	return o.String()
}

func (d Diagnostic) Dispose() {
	C.clang_disposeDiagnostic(d.c)
}

func (sl SourceLocation) Equal(sl2 SourceLocation) bool {
	o := C.clang_equalLocations(sl.c, sl2.c)
	return o != C.uint(0)
}

func (sl SourceLocation) ExpansionLocation() (File, uint32, uint32, uint32) {
	var file File
	var line C.uint
	var column C.uint
	var offset C.uint

	C.clang_getExpansionLocation(sl.c, &file.c, &line, &column, &offset)

	return file, uint32(line), uint32(column), uint32(offset)
}

func (f File) Name() string {
	o := cxstring{C.clang_getFileName(f.c)}
	defer o.Dispose()

	return o.String()
}

func (ccr *CodeCompleteResults) Results() []CompletionResult {
	var s []CompletionResult
	gos_s := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	gos_s.Cap = int(ccr.c.NumResults)
	gos_s.Len = int(ccr.c.NumResults)
	gos_s.Data = uintptr(unsafe.Pointer(ccr.c.Results))

	return s
}

func (ccr *CodeCompleteResults) Dispose() {
	C.clang_disposeCodeCompleteResults(ccr.c)
}

// The code-completion string that describes how to insert this code-completion result into the editing buffer.
func (cr CompletionResult) CompletionString() CompletionString {
	return CompletionString{cr.c.CompletionString}
}

func (cs CompletionString) Availability() AvailabilityKind {
	return AvailabilityKind(C.clang_getCompletionAvailability(cs.c))
}

func (cs CompletionString) Priority() uint32 {
	return uint32(C.clang_getCompletionPriority(cs.c))
}

func (cs CompletionString) NumChunks() uint32 {
	return uint32(C.clang_getNumCompletionChunks(cs.c))
}

func (cs CompletionString) ChunkKind(chunkNumber uint32) CompletionChunkKind {
	return CompletionChunkKind(C.clang_getCompletionChunkKind(cs.c, C.uint(chunkNumber)))
}

func (cs CompletionString) ChunkText(chunkNumber uint32) string {
	o := cxstring{C.clang_getCompletionChunkText(cs.c, C.uint(chunkNumber))}
	defer o.Dispose()

	return o.String()
}

func (c Cursor) IsNull() bool {
	o := C.clang_Cursor_isNull(c.c)
	return o != C.int(0)
}

func (c Cursor) Type() Type {
	return Type{C.clang_getCursorType(c.c)}
}

func (t Type) CanonicalType() Type {
	return Type{C.clang_getCanonicalType(t.c)}
}

func (t Type) Spelling() string {
	o := cxstring{C.clang_getTypeSpelling(t.c)}
	defer o.Dispose()

	return o.String()
}
