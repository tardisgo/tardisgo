package tgoutil

import (
	"fmt"
	"unicode"
)

// MakeID cleans-up Go names to replace characters outside (_,0-9,a-z,A-Z) with a decimal value surrounded by underlines, with special handling of '.', '*' etc..
// It also doubles-up uppercase letters, because file names are made from these names and OSX is case insensitive.
func MakeID(s string) (r string) {
	var b []rune
	b = []rune(s)
	for i := range b {
		if b[i] == '_' || ((b[i] >= 'a') && (b[i] <= 'z')) || ((b[i] >= 'A') && (b[i] <= 'Z')) || ((b[i] >= '0') && (b[i] <= '9')) {
			r += string(b[i])
			if unicode.IsUpper(b[i]) {
				r += string(b[i])
			}
		} else {
			switch b[i] {
			case '.':
				r += "_dt_"
			case '*':
				r += "_str_"
			case '/':
				r += "_slsh_"
			case ':':
				r += "_cln_"
			case '#':
				r += "_hsh_"
			case '$':
				r += "_dlr_"
			case '(':
				r += "_obr_"
			case ')':
				r += "_cbr_"
			default:
				r += fmt.Sprintf("_%d_", b[i])
			}
		}
	}
	return r
}
