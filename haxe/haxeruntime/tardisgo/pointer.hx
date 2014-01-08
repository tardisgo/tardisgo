
// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.


package tardisgo;

@:keep
class UnsafePointer extends Pointer {  // Unsafe Pointers are not yet supported, but Go library code requires that they can be created
	public function new(x:Dynamic){
		super(x);
	}
}

@:keep
class Pointer {
	private var heapObj:Dynamic; // the actual object holding the value
	private var offs:Array<Int>; // the offsets into the object, if any

	public function new(from:Dynamic){
		heapObj = from;
		offs = new Array<Int>();
	}
	public inline function load():Dynamic { 
		// this was intended to return a copy of the object, rather than a reference to it, as in:
		// return Deep.copy(this.ref()); 
		// but seems to work without problem, and significantly more quickly, without this safeguard
		return this.ref(); // TODO review
	}
	public function ref():Dynamic { // this returns a reference to the pointed-at object, not for general use!
		var ret:Dynamic = heapObj;
		for(i in 0...offs.length) 
				ret = ret[offs[i]];
		return ret;	
	}
	public  function store(v:Dynamic){
		if(offs==null) offs=[]; // C# seems to need this for HaxeInt64Typedef values
		switch ( offs.length ) {
		case 0: heapObj=v;
		case 1: heapObj[offs[0]]=v;
		default:
			var a:Dynamic = heapObj;
			for(i in 0...offs.length-1) a = a[offs[i]];
			a[offs[offs.length-1]] = v;
		}
	}
	public function addr(off:Int):Pointer {
		if(off<0 || off >= this.ref().length) Scheduler.panicFromHaxe("index out of range for valid pointer address");
		var ret:Pointer = new Pointer(this.heapObj);
		ret.offs = this.offs.copy();
		ret.offs[this.offs.length]=off;
		return ret;
	}
	public function len():Int { // used by array bounds check (which ocurrs twice as belt-and-braces while we are in beta testing, see above) TODO optimise
		return this.ref().length; 
	}
	public static function copy(v:Pointer):Pointer {
		var r:Pointer = new Pointer(v.heapObj); // no copy of data, just the reference
		r.offs = v.offs.copy();
		return r;
	}
}
