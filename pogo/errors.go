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

// TODO remove all global scope

var hadErrors = false

var stopOnError = true // TODO make this soft and default true

var warnings = make([]string, 0) // Warnings are collected up and added to the end of the output code.

var messagesGiven = make(map[string]bool) // This map de-dups error messages

// Utility message handler for errors
func logMessage(level, loc, lang string, err error) {
	msg := fmt.Sprintf("%s : %s (%s) %v \n", level, loc, lang, err)
	// don't emit duplicate messages
	_, hadIt := messagesGiven[msg]
	if !hadIt {
		fmt.Fprintf(os.Stderr, "%s", msg)
		messagesGiven[msg] = true
	}
}

// LogWarning but a warning does not stop the compiler from claiming success.
func LogWarning(loc, lang string, err error) {
	warnings = append(warnings, fmt.Sprintf("Warning: %s (%s) %v", loc, lang, err))
}

// LogError and potentially stop the compilation process.
func LogError(loc, lang string, err error) {
	logMessage("Error", loc, lang, err)
	hadErrors = true
}

// CodePosition is a utility to provide a string version of token.Pos.
// this string should be used for documentation & debug only.
func CodePosition(pos token.Pos) string {

	p := rootProgram.Fset.Position(pos).String()
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

// PosHashFileList holds the list of input go files with their posHash information
var PosHashFileList = make([]PosHashFileStruct, 0)

// Create the PosHashFileList to enable poshash values to be emitted
func setupPosHash() {
	rootProgram.Fset.Iterate(func(fRef *token.File) bool {
		PosHashFileList = append(PosHashFileList, PosHashFileStruct{FileName: fRef.Name(), LineCount: fRef.LineCount()})
		return true
	})
	sort.Sort(posHashFileSorter(PosHashFileList))
	for f := range PosHashFileList {
		if f > 0 {
			PosHashFileList[f].BasePosHash = PosHashFileList[f-1].BasePosHash + PosHashFileList[f-1].LineCount
		}
	}
}

// LatestValidPosHash holds the latest valid PosHash value seen, for use when an invalid one requires a "near" reference.
// TODO like all other global values in the pogo module, this should not be a global state.
var LatestValidPosHash = NoPosHash

// MakePosHash keeps track of references put into the code for later extraction in a runtime debug function.
// It returns the PosHash integer to be used for exception handling that was passed in.
func MakePosHash(pos token.Pos) PosHash {
	if pos.IsValid() {
		fname := rootProgram.Fset.Position(pos).Filename
		for f := range PosHashFileList {
			if PosHashFileList[f].FileName == fname {
				LatestValidPosHash = PosHash(PosHashFileList[f].BasePosHash + rootProgram.Fset.Position(pos).Line)
				return LatestValidPosHash
			}
		}
		panic(fmt.Errorf("pogo.MakePosHash() Cant find file: %s", fname))
	} else {
		if LatestValidPosHash == NoPosHash {
			return NoPosHash
		}
		return -LatestValidPosHash // -ve value => nearby reference
	}
}
