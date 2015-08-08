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
func (comp *Compilation) WriteAsClass(name, code string) {
	l := comp.TargetLang
	if LanguageList[l].files == nil {
		LanguageList[l].files = make([]FileOutput, 0, 100)
	}
	_, err := LanguageList[l].buffer.WriteString(code)
	if err != nil {
		panic(err)
	}
	var data = make([]byte, LanguageList[l].buffer.Len())
	copy(data, LanguageList[l].buffer.Bytes())
	LanguageList[l].files = append(LanguageList[l].files, FileOutput{name, data})
	LanguageList[l].buffer.Reset()
	comp.emitFileStart()
}

func (comp *Compilation) targetDir() error {
	if err := os.Mkdir(LanguageList[comp.TargetLang].TgtDir, os.ModePerm); err != nil {
		if !os.IsExist(err) { // no problem if it already exists
			comp.LogError("Unable to create tardis output directory", "pogo", err)
			return err
		}
	}
	return nil
}

// Write out the target language file
func (comp *Compilation) writeFiles() {
	l := comp.TargetLang
	if LanguageList[l].buffer.Len() > 0 {
		comp.WriteAsClass("Remnants", "")
	}
	err := comp.targetDir()
	if err == nil {
		for _, fo := range LanguageList[l].files {
			err = writeIfChanged(
				LanguageList[comp.TargetLang].TgtDir+
					string(os.PathSeparator)+fo.filename+
					LanguageList[l].FileTypeSuffix(),
				fo.data)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		comp.LogError("Unable to write output file", "pogo", err)
	}
}
