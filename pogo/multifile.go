package pogo

import (
	"bytes"
	"io/ioutil"
	"os"
)

func writeIfChanged(filename string, data []byte) error {
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		if bytes.Equal(content, data) {
			//println("DEBUG writeIfChanged() unchanged: " + filename)
			return nil
		}
	}
	//println("DEBUG writeIfChanged() changed: " + filename) // TODO add a flag to show this
	return ioutil.WriteFile(filename, data, 0666)
}

// WriteAsClass writes the contents of the buffer as a given class file name.
// For haxe this name must begin with an upper-case letter and match the underlying class name.
func WriteAsClass(name, code string) {
	l := TargetLang
	if LanguageList[l].files == nil {
		LanguageList[l].files = make([]FileOutput, 0, 100)
	}
	LanguageList[l].buffer.WriteString(code)
	var data = make([]byte, LanguageList[l].buffer.Len())
	copy(data, LanguageList[l].buffer.Bytes())
	LanguageList[l].files = append(LanguageList[l].files, FileOutput{name, data})
	LanguageList[l].buffer.Reset()
	emitFileStart()
}

const TgtDir = "tardis" // TODO move to the correct directory based on a command line argument

func targetDir() error {
	if err := os.Mkdir(TgtDir, os.ModePerm); err != nil {
		if !os.IsExist(err) { // no problem if it already exists
			LogError("Unable to create tardis output directory", "pogo", err)
			return err
		}
	}
	return nil
}

// Write out the target language file
// TODO consider writing multiple output files, if this would be better/required for some target languages.
func writeFiles() {
	l := TargetLang
	if LanguageList[l].buffer.Len() > 0 {
		WriteAsClass("Remnants", "")
	}
	err := targetDir()
	if err == nil {
		for _, fo := range LanguageList[l].files {
			err = writeIfChanged(
				TgtDir+string(os.PathSeparator)+fo.filename+LanguageList[l].FileTypeSuffix(), // Ubuntu requires the first letter of the haxe file to be uppercase
				fo.data)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		LogError("Unable to write output file", "pogo", err)
	}
}
