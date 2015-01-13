package flag

func Bool(name string, value bool, usage string) *bool {
	return &value
}
