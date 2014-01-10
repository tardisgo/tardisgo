// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"os"
)

// TODO remove all global scope

var hadErrors bool = false

var stopOnError bool = true // TODO make this soft and default true

var warnings []string = make([]string, 0) // Warnings are collected up and added to the end of the output code.

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

// A warning does not stop the compiler from claiming success.
func LogWarning(loc, lang string, err error) {
	warnings = append(warnings, fmt.Sprintf("Warning: %s (%s) %v", loc, lang, err))
}

// An error will stop the compilation process.
func LogError(loc, lang string, err error) {
	logMessage("Error", loc, lang, err)
	hadErrors = true
}

// Utility to provide a string version of token.Pos.
// this string should be used for documentation & debug only.
func CodePosition(pos token.Pos) string {

	p := rootProgram.Fset.Position(pos).String()
	if p == "-" {
		return ""
	}
	return p
}

type PosHash int // A hash of the code position, set -ve if a nearby PosHash is used.

const NoPosHash = 0 // Code position hash constant to represent none, lines number from 1, so 0 is invalid.

// Struct to store the code position information for each file in order to generate PosHash values.
type PosHashFileStruct struct {
	FileName    string // The name of the file.
	LineCount   int    // The number of lines in that file.
	BasePosHash int    // The base posHash value for this file.
}

var PosHashFileList []PosHashFileStruct = make([]PosHashFileStruct, 0) // The list of input go files with their posHash information

// Create the PosHashFileList to enable poshash values to be emitted
func setupPosHash() {
	rootProgram.Fset.Iterate(func(fRef *token.File) bool {
		PosHashFileList = append(PosHashFileList, PosHashFileStruct{FileName: fRef.Name(), LineCount: fRef.LineCount()})
		return true
	})
	for f := range PosHashFileList {
		if f > 0 {
			PosHashFileList[f].BasePosHash = PosHashFileList[f-1].BasePosHash + PosHashFileList[f-1].LineCount
		}
	}
}

// The latest valid PosHash value seen, for use when an invalid one requires a "near" reference.
// TODO like all other global values in the pogo module, this should not be a global state.
var LatestValidPosHash PosHash = NoPosHash

// This function keeps track of references put into the code for later extraction in a runtime debug function.
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
		panic(fmt.Errorf("MakePosHash() Cant find file: %s", fname))
	} else {
		if LatestValidPosHash == NoPosHash {
			return NoPosHash
		} else {
			return -LatestValidPosHash // -ve value => nearby reference
		}
	}
}
