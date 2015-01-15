// +build haxe

package reflect

// the haxe-specific parts

var haxeIDmap = make(map[int]Type)

func haxeInterfaceUnpack(i interface{}) *emptyInterface {
	panic("reflect.haxeInterfaceUnpack() not yet implemented")
	return nil
}

func haxeInterfacePack(*emptyInterface) interface{} {
	panic("reflect.haxeInterfacePack() not yet implemented")
	return interface{}(nil)
}
