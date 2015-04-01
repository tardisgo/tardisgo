package haxegoruntime

func defer_close(c chan interface{}) {
	close(c)
}

//func init() {
//	x := make(chan interface{})
//	defer_close(x) // to make sure it is not removed by Dead Code Elimination
//}
