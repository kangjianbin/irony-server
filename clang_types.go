package main

// #cgo LDFLAGS: -lclang
// #include <clang-c/Index.h>
// #include <clang-c/CXCompilationDatabase.h>
// #include <stdlib.h>
import "C"

type Index struct {
	c C.CXIndex
}

type UnsavedFile struct {
	c C.struct_CXUnsavedFile
}

type TranslationUnit struct {
	c C.CXTranslationUnit
}

type Diagnostic struct {
	c C.CXDiagnostic
}

type SourceLocation struct {
	c C.CXSourceLocation
}

type File struct {
	c C.CXFile
}

type cxstring struct {
	c C.CXString
}

type CodeCompleteResults struct {
	c *C.CXCodeCompleteResults
}

type CompletionResult struct {
	c C.CXCompletionResult
}

type CompletionString struct {
	c C.CXCompletionString
}

type Cursor struct {
	c C.CXCursor
}

type Type struct {
	c C.CXType
}

type ErrorCode uint32

const (
	// No error.
	Error_Success ErrorCode = C.CXError_Success
	/*
		A generic error code, no further details are available.

		Errors of this kind can get their own specific error codes in future
		libclang versions.
	*/
	Error_Failure = C.CXError_Failure
	// libclang crashed while performing the requested operation.
	Error_Crashed = C.CXError_Crashed
	// The function detected that the arguments violate the function contract.
	Error_InvalidArguments = C.CXError_InvalidArguments
	// An AST deserialization error has occurred.
	Error_ASTReadError = C.CXError_ASTReadError
)

type TranslationUnit_Flags uint32

const (
	// Used to indicate that no special translation-unit options are needed.
	TranslationUnit_None TranslationUnit_Flags = C.CXTranslationUnit_None
	/*
		Used to indicate that the parser should construct a "detailed"
		preprocessing record, including all macro definitions and instantiations.

		Constructing a detailed preprocessing record requires more memory
		and time to parse, since the information contained in the record
		is usually not retained. However, it can be useful for
		applications that require more detailed information about the
		behavior of the preprocessor.
	*/
	TranslationUnit_DetailedPreprocessingRecord = C.CXTranslationUnit_DetailedPreprocessingRecord
	/*
		Used to indicate that the translation unit is incomplete.

		When a translation unit is considered "incomplete", semantic
		analysis that is typically performed at the end of the
		translation unit will be suppressed. For example, this suppresses
		the completion of tentative declarations in C and of
		instantiation of implicitly-instantiation function templates in
		C++. This option is typically used when parsing a header with the
		intent of producing a precompiled header.
	*/
	TranslationUnit_Incomplete = C.CXTranslationUnit_Incomplete
	/*
		Used to indicate that the translation unit should be built with an
		implicit precompiled header for the preamble.

		An implicit precompiled header is used as an optimization when a
		particular translation unit is likely to be reparsed many times
		when the sources aren't changing that often. In this case, an
		implicit precompiled header will be built containing all of the
		initial includes at the top of the main file (what we refer to as
		the "preamble" of the file). In subsequent parses, if the
		preamble or the files in it have not changed, \c
		clang_reparseTranslationUnit() will re-use the implicit
		precompiled header to improve parsing performance.
	*/
	TranslationUnit_PrecompiledPreamble = C.CXTranslationUnit_PrecompiledPreamble
	/*
		Used to indicate that the translation unit should cache some
		code-completion results with each reparse of the source file.

		Caching of code-completion results is a performance optimization that
		introduces some overhead to reparsing but improves the performance of
		code-completion operations.
	*/
	TranslationUnit_CacheCompletionResults = C.CXTranslationUnit_CacheCompletionResults
	/*
		Used to indicate that the translation unit will be serialized with
		clang_saveTranslationUnit.

		This option is typically used when parsing a header with the intent of
		producing a precompiled header.
	*/
	TranslationUnit_ForSerialization = C.CXTranslationUnit_ForSerialization
	/*
		DEPRECATED: Enabled chained precompiled preambles in C++.

		Note: this is a *temporary* option that is available only while
		we are testing C++ precompiled preamble support. It is deprecated.
	*/
	TranslationUnit_CXXChainedPCH = C.CXTranslationUnit_CXXChainedPCH
	/*
		Used to indicate that function/method bodies should be skipped while
		parsing.

		This option can be used to search for declarations/definitions while
		ignoring the usages.
	*/
	TranslationUnit_SkipFunctionBodies = C.CXTranslationUnit_SkipFunctionBodies
	// Used to indicate that brief documentation comments should be included into the set of code completions returned from this translation unit.
	TranslationUnit_IncludeBriefCommentsInCodeCompletion = C.CXTranslationUnit_IncludeBriefCommentsInCodeCompletion
	// Used to indicate that the precompiled preamble should be created on the first parse. Otherwise it will be created on the first reparse. This trades runtime on the first parse (serializing the preamble takes time) for reduced runtime on the second parse (can now reuse the preamble).
	TranslationUnit_CreatePreambleOnFirstParse = C.CXTranslationUnit_CreatePreambleOnFirstParse
	/*
		Do not stop processing when fatal errors are encountered.

		When fatal errors are encountered while parsing a translation unit,
		semantic analysis is typically stopped early when compiling code. A common
		source for fatal errors are unresolvable include files. For the
		purposes of an IDE, this is undesirable behavior and as much information
		as possible should be reported. Use this flag to enable this behavior.
	*/
	TranslationUnit_KeepGoing = C.CXTranslationUnit_KeepGoing
)

type DiagnosticSeverity uint32

const (
	// A diagnostic that has been suppressed, e.g., by a command-line option.
	Diagnostic_Ignored DiagnosticSeverity = C.CXDiagnostic_Ignored
	// This diagnostic is a note that should be attached to the previous (non-note) diagnostic.
	Diagnostic_Note = C.CXDiagnostic_Note
	// This diagnostic indicates suspicious code that may not be wrong.
	Diagnostic_Warning = C.CXDiagnostic_Warning
	// This diagnostic indicates that the code is ill-formed.
	Diagnostic_Error = C.CXDiagnostic_Error
	// This diagnostic indicates that the code is ill-formed such that future parser recovery is unlikely to produce useful results.
	Diagnostic_Fatal = C.CXDiagnostic_Fatal
)

type AvailabilityKind uint32

const (
	// The entity is available.
	Availability_Available AvailabilityKind = C.CXAvailability_Available
	// The entity is available, but has been deprecated (and its use is not recommended).
	Availability_Deprecated = C.CXAvailability_Deprecated
	// The entity is not available; any use of it will be an error.
	Availability_NotAvailable = C.CXAvailability_NotAvailable
	// The entity is available, but not accessible; any use of it will be an error.
	Availability_NotAccessible = C.CXAvailability_NotAccessible
)

type CompletionChunkKind uint32

const (
	/*
		A code-completion string that describes "optional" text that
		could be a part of the template (but is not required).

		The Optional chunk is the only kind of chunk that has a code-completion
		string for its representation, which is accessible via
		clang_getCompletionChunkCompletionString(). The code-completion string
		describes an additional part of the template that is completely optional.
		For example, optional chunks can be used to describe the placeholders for
		arguments that match up with defaulted function parameters, e.g. given:

		\code
		void f(int x, float y = 3.14, double z = 2.71828);
		\endcode

		The code-completion string for this function would contain:
		- a TypedText chunk for "f".
		- a LeftParen chunk for "(".
		- a Placeholder chunk for "int x"
		- an Optional chunk containing the remaining defaulted arguments, e.g.,
		- a Comma chunk for ","
		- a Placeholder chunk for "float y"
		- an Optional chunk containing the last defaulted argument:
		- a Comma chunk for ","
		- a Placeholder chunk for "double z"
		- a RightParen chunk for ")"

		There are many ways to handle Optional chunks. Two simple approaches are:
		- Completely ignore optional chunks, in which case the template for the
		function "f" would only include the first parameter ("int x").
		- Fully expand all optional chunks, in which case the template for the
		function "f" would have all of the parameters.
	*/
	CompletionChunk_Optional CompletionChunkKind = C.CXCompletionChunk_Optional
	/*
		Text that a user would be expected to type to get this
		code-completion result.

		There will be exactly one "typed text" chunk in a semantic string, which
		will typically provide the spelling of a keyword or the name of a
		declaration that could be used at the current code point. Clients are
		expected to filter the code-completion results based on the text in this
		chunk.
	*/
	CompletionChunk_TypedText = C.CXCompletionChunk_TypedText
	/*
		Text that should be inserted as part of a code-completion result.

		A "text" chunk represents text that is part of the template to be
		inserted into user code should this particular code-completion result
		be selected.
	*/
	CompletionChunk_Text = C.CXCompletionChunk_Text
	/*
		Placeholder text that should be replaced by the user.

		A "placeholder" chunk marks a place where the user should insert text
		into the code-completion template. For example, placeholders might mark
		the function parameters for a function declaration, to indicate that the
		user should provide arguments for each of those parameters. The actual
		text in a placeholder is a suggestion for the text to display before
		the user replaces the placeholder with real code.
	*/
	CompletionChunk_Placeholder = C.CXCompletionChunk_Placeholder
	/*
		Informative text that should be displayed but never inserted as
		part of the template.

		An "informative" chunk contains annotations that can be displayed to
		help the user decide whether a particular code-completion result is the
		right option, but which is not part of the actual template to be inserted
		by code completion.
	*/
	CompletionChunk_Informative = C.CXCompletionChunk_Informative
	/*
		Text that describes the current parameter when code-completion is
		referring to function call, message send, or template specialization.

		A "current parameter" chunk occurs when code-completion is providing
		information about a parameter corresponding to the argument at the
		code-completion point. For example, given a function

		\code
		int add(int x, int y);
		\endcode

		and the source code add(, where the code-completion point is after the
		"(", the code-completion string will contain a "current parameter" chunk
		for "int x", indicating that the current argument will initialize that
		parameter. After typing further, to add(17, (where the code-completion
		point is after the ","), the code-completion string will contain a
		"current paremeter" chunk to "int y".
	*/
	CompletionChunk_CurrentParameter = C.CXCompletionChunk_CurrentParameter
	// A left parenthesis ('('), used to initiate a function call or signal the beginning of a function parameter list.
	CompletionChunk_LeftParen = C.CXCompletionChunk_LeftParen
	// A right parenthesis (')'), used to finish a function call or signal the end of a function parameter list.
	CompletionChunk_RightParen = C.CXCompletionChunk_RightParen
	// A left bracket ('[').
	CompletionChunk_LeftBracket = C.CXCompletionChunk_LeftBracket
	// A right bracket (']').
	CompletionChunk_RightBracket = C.CXCompletionChunk_RightBracket
	// A left brace ('{').
	CompletionChunk_LeftBrace = C.CXCompletionChunk_LeftBrace
	// A right brace ('}').
	CompletionChunk_RightBrace = C.CXCompletionChunk_RightBrace
	// A left angle bracket ('<').
	CompletionChunk_LeftAngle = C.CXCompletionChunk_LeftAngle
	// A right angle bracket ('>').
	CompletionChunk_RightAngle = C.CXCompletionChunk_RightAngle
	// A comma separator (',').
	CompletionChunk_Comma = C.CXCompletionChunk_Comma
	/*
		Text that specifies the result type of a given result.

		This special kind of informative chunk is not meant to be inserted into
		the text buffer. Rather, it is meant to illustrate the type that an
		expression using the given completion string would have.
	*/
	CompletionChunk_ResultType = C.CXCompletionChunk_ResultType
	// A colon (':').
	CompletionChunk_Colon = C.CXCompletionChunk_Colon
	// A semicolon (';').
	CompletionChunk_SemiColon = C.CXCompletionChunk_SemiColon
	// An '=' sign.
	CompletionChunk_Equal = C.CXCompletionChunk_Equal
	// Horizontal space (' ').
	CompletionChunk_HorizontalSpace = C.CXCompletionChunk_HorizontalSpace
	// Vertical space ('\n'), after which it is generally a good idea to perform indentation.
	CompletionChunk_VerticalSpace = C.CXCompletionChunk_VerticalSpace
)

type CompilationDatabase struct {
	c C.CXCompilationDatabase
}

type CompilationDatabase_Error int32

const (
	CompilationDatabase_NoError            CompilationDatabase_Error = C.CXCompilationDatabase_NoError
	CompilationDatabase_CanNotLoadDatabase                           = C.CXCompilationDatabase_CanNotLoadDatabase
)

type CompileCommands struct {
	c C.CXCompileCommands
}

type CompileCommand struct {
	c C.CXCompileCommand
}
