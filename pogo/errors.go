// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"os"
	"sort"
)

func (comp *Compilation) initErrors() {
	comp.hadErrors = false
	comp.stopOnError = true                    // TODO make this soft and default true
	comp.warnings = make([]string, 0)          // Warnings are collected up and added to the end of the output code.
	comp.messagesGiven = make(map[string]bool) // This map de-dups error messages

	// PosHashFileList holds the list of input go files with their posHash information
	comp.PosHashFileList = make([]PosHashFileStruct, 0)
	// LatestValidPosHash holds the latest valid PosHash value seen, for use when an invalid one requires a "near" reference.
	comp.LatestValidPosHash = NoPosHash
}

// Utility message handler for errors
func (comp *Compilation) logMessage(level, loc, lang string, err error) {
	msg := fmt.Sprintf("%s : %s (%s) %v \n", level, loc, lang, err)
	// don't emit duplicate messages
	_, hadIt := comp.messagesGiven[msg]
	if !hadIt {
		fmt.Fprintf(os.Stderr, "%s", msg)
		comp.messagesGiven[msg] = true
	}
}

// LogWarning but a warning does not stop the compiler from claiming success.
func (comp *Compilation) LogWarning(loc, lang string, err error) {
	comp.warnings = append(comp.warnings, fmt.Sprintf("Warning: %s (%s) %v", loc, lang, err))
}

// LogError and potentially stop the compilation process.
func (comp *Compilation) LogError(loc, lang string, err error) {
	comp.logMessage("Error", loc, lang, err)
	comp.hadErrors = true
}

// CodePosition is a utility to provide a string version of token.Pos.
// this string should be used for documentation & debug only.
func (comp *Compilation) CodePosition(pos token.Pos) string {

	p := comp.rootProgram.Fset.Position(pos).String()
	if p == "-" {
		return ""
	}
	return p
}

// A PosHash is a hash of the code position, set -ve if a nearby PosHash is used.
type PosHash int

// NoPosHash is a code position hash constant to represent none, lines number from 1, so 0 is invalid.
const NoPosHash = PosHash(0)

// PosHashFileStruct stores the code position information for each file in order to generate PosHash values.
type PosHashFileStruct struct {
	FileName    string // The name of the file.
	LineCount   int    // The number of lines in that file.
	BasePosHash int    // The base posHash value for this file.
}

type posHashFileSorter []PosHashFileStruct

func (a posHashFileSorter) Len() int           { return len(a) }
func (a posHashFileSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a posHashFileSorter) Less(i, j int) bool { return a[i].FileName < a[j].FileName }

// Create the PosHashFileList to enable poshash values to be emitted
func (comp *Compilation) setupPosHash() {
	comp.rootProgram.Fset.Iterate(func(fRef *token.File) bool {
		comp.PosHashFileList = append(comp.PosHashFileList,
			PosHashFileStruct{FileName: fRef.Name(), LineCount: fRef.LineCount()})
		return true
	})
	sort.Sort(posHashFileSorter(comp.PosHashFileList))
	for f := range comp.PosHashFileList {
		if f > 0 {
			comp.PosHashFileList[f].BasePosHash =
				comp.PosHashFileList[f-1].BasePosHash + comp.PosHashFileList[f-1].LineCount
		}
	}
}

// MakePosHash keeps track of references put into the code for later extraction in a runtime debug function.
// It returns the PosHash integer to be used for exception handling that was passed in.
func (comp *Compilation) MakePosHash(pos token.Pos) PosHash {
	if pos.IsValid() {
		fname := comp.rootProgram.Fset.Position(pos).Filename
		for f := range comp.PosHashFileList {
			if comp.PosHashFileList[f].FileName == fname {
				comp.LatestValidPosHash = PosHash(comp.PosHashFileList[f].BasePosHash +
					comp.rootProgram.Fset.Position(pos).Line)
				return comp.LatestValidPosHash
			}
		}
		panic(fmt.Errorf("pogo.MakePosHash() Cant find file: %s", fname))
	} else {
		if comp.LatestValidPosHash == NoPosHash {
			return NoPosHash
		}
		return -comp.LatestValidPosHash // -ve value => nearby reference
	}
}
