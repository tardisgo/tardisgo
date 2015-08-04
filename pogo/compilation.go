package pogo

import (
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types/typeutil"
)

// Pogo contains global variables for an individual pogo run
type Compilation struct {
	rootProgram *ssa.Program // pointer to the root datastructure
	mainPackage *ssa.Package // pointer to the "main" package
	DebugFlag   bool         // DebugFlag is used to signal if we are emitting debug information
	TraceFlag   bool         // TraceFlag is used to signal if we are emitting trace information (big)
	TargetLang  int          // TargetLang holds the language currently being targeted, offset into LanguageList.

	hxPkgName, headerText string
	LibListNoDCE          []string

	hadErrors, stopOnError bool                // TODO make stopOnError soft and default true
	warnings               []string            // Warnings are collected up and added to the end of the output code.
	messagesGiven          map[string]bool     // This map de-dups error messages
	PosHashFileList        []PosHashFileStruct // PosHashFileList holds the list of input go files with their posHash information
	LatestValidPosHash     PosHash             // LatestValidPosHash holds the latest valid PosHash value seen, for use when an invalid one requires a "near" reference.

	fnMap, grMap map[*ssa.Function]bool // which functions are used and if the functions use goroutines/channels

	inlineMap map[string]string
	keysSeen  map[string]int

	previousErrorInfo string // used to give some indication of the error's location, even if it is not given

	TypesEncountered         typeutil.Map // TypesEncountered keeps track of the types we encounter using the excellent go.tools/go/types/typesmap package.
	NextTypeID               int          // NextTypeID is used to give each type we come across its own ID - entry zero is invalid
	catchReferencedTypesSeen map[string]bool
}
