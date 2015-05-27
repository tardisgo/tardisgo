// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import "github.com/tardisgo/tardisgo/pogo"

// Runtime Haxe code for Go, which may eventually become a haxe library when the system settles down.
// TODO All runtime class names are currently carried through if the haxe code uses "import tardis.Go;" and some are too generic,
// others, like Int64, will overload the Haxe standard library version for some platforms, which may cause other problems.
// So the haxe Class names eventually need to be prefaced with "Go" to ensure there are no name-clashes.
// TODO using a library would reduce some of the compilation overhead of always re-compiling this code.
// However, there are references to Go->Haxe generated classes, like "Go", that would need to be managed somehow.
// TODO consider merging and possibly renaming the Deep and Force classes as they both hold general utility code

func haxeruntime() string {

	pogo.WriteAsClass("Console", `

class Console {
	public static inline function naclWrite(v:String){
		#if ( cpp || cs || java || neko || php || python )
			Sys.print(v);
		#else
			haxe.Log.trace(v);
		#end
	}
	public static inline function println(v:Array<Dynamic>) {
		#if ( cpp || cs || java || neko || php || python )
			Sys.println(join(v));
		#else
			haxe.Log.trace(join(v));
		#end
	}
	public static inline function print(v:Array<Dynamic>) {
		#if ( cpp || cs || java || neko || php || python )
			Sys.print(join(v));
		#else
			haxe.Log.trace(join(v));
		#end
	}
	static function join(v:Array<Dynamic>):String {
		var s = "";
		for (i in 0...v.length) {
			if(Std.is(v[i],String)) 
				s+= Force.toHaxeString(v[i]) + " ";
			else
				s += Std.string(v[i]) + " " ;
		}
		return s;
	}
	public static function readln():Null<String> {
		#if (cpp || cs || java || neko || php )
			var s:String="";
			var ch:Int=0;
			while(ch != 13 ){ // carrage return (mac)
				ch = Sys.getChar(true);
				//Sys.println(ch);
				if(ch == 127){ // backspace (mac)
					s = s.substr(0,s.length-1);
					Sys.print("\n"+s);
				}else{
					s += String.fromCharCode(ch);
				}
			}
			s = s.substr(0,s.length-1); // loose final CR
			Sys.print("\n");
			if(s.length==0)
				return null;
			else
				return s;
		#else
			return null;
		#end
	}
}

`)
	pogo.WriteAsClass("Force", `
// TODO: consider putting these go-compatibiliy classes into a separate library for general Haxe use when calling Go

class Force { // TODO maybe this should not be a separate haxe class, as no non-Go code needs access to it
	public static inline function toUint8(v:Int): #if cpp cpp.UInt8 #else Int #end
	{
		#if cpp 
			return v;
		#elseif cs
			return cast(cast(v,cs.StdTypes.UInt8),Int);
		#else
			return v & 0xFF;
		#end
	}	
	public static inline function toUint16(v:Int): #if cpp cpp.UInt16 #else Int #end {
		#if cpp 
			return v; 
		#elseif cs
			return cast(cast(v,cs.StdTypes.UInt16),Int);
		#else
			return v & 0xFFFF;
		#end
	}	
	public static inline function toUint32(v:Int):Int { 
		#if js 
			return v >>> untyped __js__("0"); // using GopherJS method (with workround to stop it being optimized away by Haxe)
		#elseif php
       		return v & untyped __php__("0xffffffff");
		#else
			return v; 
		#end
	}	
	public static inline function toUint64(v:GOint64):GOint64 {
		return v;
	}	
	public static #if (cpp||java||cs) inline #end function toInt8(v:Int): #if cpp cpp.Int8 #else Int #end  {
		#if cpp 
			return v; 
		#elseif java
			return cast(cast(v,java.StdTypes.Int8),Int);
		#elseif cs
			return cast(cast(v,cs.StdTypes.Int8),Int);
		#else
			var r:Int = v & 0xFF;
			if ((r & 0x80) != 0){ // it should be -ve
				return -1 - 0xFF + r;
			}
			return r;
		#end
	}	
	public static #if (cpp||java||cs) inline #end function toInt16(v:Int): #if cpp cpp.Int16 #else Int #end {
		#if cpp 
			return v; 
		#elseif java
			return cast(cast(v,java.StdTypes.Int16),Int);
		#elseif cs
			return cast(cast(v,cs.StdTypes.Int16),Int);
		#else
			var r:Int = v & 0xFFFF;
			if ((r & 0x8000) != 0){ // it should be -ve
				return -1 - 0xFFFF + r;
			}
			return r;
		#end
	}	
	public static inline function toInt32(v:Int):Int {
		#if js 
			return v >> untyped __js__("0"); // using GopherJS method (with workround to stop it being optimized away by Haxe)
		#elseif php
			//see: http://stackoverflow.com/questions/300840/force-php-integer-overflow
     		v = (v & untyped __php__("0xFFFFFFFF"));
 		    if( (v & untyped __php__("0x80000000")) != 0)
		        v = -((~v & untyped __php__("0xFFFFFFFF")) + 1);
		    return v;
		#else
			return v;
		#end
	}	
	public static inline function toInt64(v:GOint64):GOint64 { // this in case special handling is required for some platforms
		return v;
	}	
	public static function toInt(v:Dynamic):Int { // get an Int from a Dynamic variable (uintptr is stored as Dynamic)
		if(v==null) return 0;
		if (Reflect.isObject(v)) 
			if(Std.is(v,Interface)) {
				v=v.val; 						// it is in an interface, so get the value
				return toInt(v); 				// recurse to handle 64-bit or float or uintptr
			} else								// it should be an Int64 if not an interface
				if(Std.is(v,Pointer)) {
					//Scheduler.panicFromHaxe("attempt to do Pointer arithmetic"); 
					return v.hashInt();
				}else
					return GOint64.toInt(v);	// may never get here if GOint64 is an abstract
		else
			if(Std.is(v,Int))
				return v;
			else
				if(Std.is(v,Float)) 			
					return v>=0?Math.floor(v):Math.ceil(v);
				else
					if(haxe.Int64.is(v)) // detect the abstract
						return haxe.Int64.toInt(v);
					else
						return cast(v,Int);	// default cast
	}
	public static inline function toFloat(v:Float):Float {
		// neko target platform requires special handling because it auto-converts whole-number Float into Int without asking
		// see: https://github.com/HaxeFoundation/haxe/issues/1282 which was marked as closed, but was not fixed as at 2013.9.6
		// NOTE: this issue means that the neko/--interp is useless for testing anything math-based
		#if neko
			if(Std.is(v,Int)) {
				return v + 2.2251e-308; // add the smallest value possible for a 64-bit float to ensure neko doesn't still think it is an int
			} else
				return v;
		#else
			return v;
		#end
	}	
	#if (cpp || neko)
		static public var f64byts = haxe.io.Bytes.alloc(8);
	#elseif js  // NOTE this code uses js dataview even when not in fullunsafe mode
		static private var f32dView = new js.html.DataView(new js.html.ArrayBuffer(8),0,8); 
	#end
	public static function toFloat32(v:Float):Float {
		#if (cpp || neko)
			f64byts.setFloat(0,v);
			return f64byts.getFloat(0);
		#elseif js 
			f32dView.setFloat32(0,v); 
			return f32dView.getFloat32(0); 
		#elseif cs
			return untyped __cs__("(double)((float)v)");
		#elseif java
			return untyped __java__("(double)((float)v)");
		#else
			if(Go.haxegoruntime_IInFF32fb.load_bool()) { // in the Float32frombits() function so don't recurse
				return v;
			} else {
				return Go_haxegoruntime_FFloat32frombits.callFromRT(0,Go_haxegoruntime_FFloat32bits.callFromRT(0,v));
			}
		#end
	}
	// TODO implement speed improvements when code below is correct
	//START POTENTIAL SPEED-UP CODE
	/*
	public static function Float32bits(v:Float):Int {
		#if (cpp || neko)
			f64byts.setFloat(0,v);
			return toUint32( f64byts.get(0) | (f64byts.get(1)<<8)  | (f64byts.get(2)<<16)  | (f64byts.get(3)<<24) ); //little-endian
		#elseif js 
			f32dView.setFloat32(0,v); 
			return toUint32(f32dView.getUint32(0)); 
		#else
			Scheduler.panicFromHaxe("Force.Float32bits unreachable code");
			return 0; 
		#end
	}
	public static function Float32frombits(v:Int):Float {
		#if (cpp || neko)
			f64byts.set(0,v&0xff);
			f64byts.set(1,(v>>8)&0xff);
			f64byts.set(2,(v>>16)&0xff);
			f64byts.set(3,(v>>24)&0xff); //little-endian
			return f64byts.getFloat(0);
		#elseif js 
			f32dView.setUint32(0,v);
			return f32dView.getFloat32(0); 
		#else
			Scheduler.panicFromHaxe("Force.Float32frombits unreachable code"); 
			return 0;
		#end
	}
	*/
	public static function Float64bits(v:Float):GOint64 {
		#if cs
			var rv:haxe.Int64 = untyped __cs__("System.BitConverter.DoubleToInt64Bits(v)");
			return rv;
		//#elseif (cpp || neko)
		//	f64byts.setDouble(0,v);
		//	return GOint64.make(
		//		f64byts.get(4) | (f64byts.get(5)<<8)  | (f64byts.get(6)<<16)  | (f64byts.get(7)<<24) ,
		//		f64byts.get(0) | (f64byts.get(1)<<8)  | (f64byts.get(2)<<16)  | (f64byts.get(3)<<24) ); //little-endian
		//#elseif js 
		//	f32dView.setFloat64(0,v); 
		//	return GOint64.make(f32dView.getUint32(4),f32dView.getUint32(0)); 
		#else
			Scheduler.panicFromHaxe("Force.Float64bits unreachable code");
			return GOint64.ofInt(0); 
		#end
	}
	public static function Float64frombits(v:GOint64):Float {
		#if cs
			var hv:haxe.Int64=v;
			return untyped __cs__("System.BitConverter.Int64BitsToDouble(hv)");
		//#elseif (cpp || neko)
		//	var v0 = GOint64.getLow(v);
		//	var v1 = GOint64.getHigh(v);
		//	f64byts.set(0,v0&0xff);
		//	f64byts.set(1,(v0>>8)&0xff);
		//	f64byts.set(2,(v0>>16)&0xff);
		//	f64byts.set(3,(v0>>24)&0xff); //little-endian
		//	f64byts.set(4,v1&0xff);
		//	f64byts.set(5,(v1>>8)&0xff);
		//	f64byts.set(6,(v1>>16)&0xff);
		//	f64byts.set(7,(v1>>24)&0xff); //little-endian
		//	return f64byts.getDouble(0);
		//#elseif js 
		//	f32dView.setUint32(0,GOint64.getLow(v));
		//	f32dView.setUint32(4,GOint64.getHigh(v));
		//	return f32dView.getFloat64(0); 
		#else
			Scheduler.panicFromHaxe("Force.Float64frombits unreachable code"); 
			return 0;
		#end
	}
	//END POTENTIAL SPEED-UP CODE
	//
	public static function uintCompare(x:Int,y:Int):Int { // +ve if uint(x)>unint(y), 0 equal, else -ve 
		#if js x=x>>>untyped __js__("0");y=y>>>untyped __js__("0"); #end
		if(x==y) return 0; // simple case first for speed TODO is it faster with this in or out?
		if(x>=0) {
			if(y>=0){ // both +ve so normal comparison
				return (x-y);
			}else{ // y -ve and so larger than x
				return -1;
			}
		}else { // x -ve
			if(y>=0){ // -ve x larger than +ve y
				return 1;
			}else{ // both are -ve so the normal comparison
				return (x-y); //eg -1::-7 gives -1--7 = +6 meaning -1 > -7
			}
		}
	}
	private static function checkIntDiv(x:Int,y:Int,byts:Int):Int { // implement the special processing required by Go
		var r:Int=y;
		switch(y) {
		case 0:
			Scheduler.panicFromHaxe("attempt to divide integer value by 0"); 
		case -1:
			switch (byts) {
			case 1:
				if(x== -128) r=1; // special case in the Go spec
			case 2:
				if(x== -32768) r=1; // special case in the Go spec
 			case 4:
				if(x== -2147483648) r=1; // special case in the Go spec
			default:
				// noOp - 0 => unsigned
			}
		}
		return r;
	}
	//TODO maybe optimize by not passing the special value and having multiple versions of functions
	public static function intDiv(x:Int,y:Int,sv:Int):Int {
		y = checkIntDiv(x,y,sv);
		if(y==1) return x; // x div 1 is x
		if((sv>0)||((x>0)&&(y>0))){ // signed division will work (even though it may be unsigned)
			var f:Float=  cast(x,Float) / cast(y,Float);
			return f>=0?Math.floor(f):Math.ceil(f);
		} else { // unsigned division 
			return toUint32(GOint64.toInt(GOint64.div(GOint64.make(0x0,x),GOint64.make(0x0,y),false)));
		}
	}
	public static function intMod(x:Int,y:Int,sv:Int):Int {
		y = checkIntDiv(x,y,sv);
		if(y==1) return 0; // x mod 1 is 0
		if((sv>0)||((x>0)&&(y>0))){ // signed mod will work (even though it may be unsigned)
			return x % y;
		} else { // unsigned mod (do it in 64 bits to ensure unsigned)
			return toUint32(GOint64.toInt(GOint64.mod(GOint64.make(0x0,x),GOint64.make(0x0,y),false)));
		}
	}
	public static inline function intMul(x:Int,y:Int,sv:Int):Int { // TODO optimize away sv
		#if (js || php)
			if(sv>0){ // signed mul
				return  x*y; //toInt32(GOint64.toInt(GOint64.mul(GOint64.ofInt(x),GOint64.ofInt(y)))); // TODO review if this required
			} else { // unsigned mul 
				return toUint32(GOint64.toInt(GOint64.mul(GOint64.ofUInt(x),GOint64.ofUInt(y)))); // required for overflowing mul
			}
		#else
			return x * y;
		#end
	}
	public static var minusZero:Float= 1.0 / Math.NEGATIVE_INFINITY ; 
	private static var zero:Float=0.0;
	private static var MinFloat64:Float = -1.797693134862315708145274237317043567981e+308; // 2**1023 * (2**53 - 1) / 2**52
	public static #if !php  inline #end function floatDiv(x:Float,y:Float):Float {
		#if ( php ) 
			// NOTE for php 0 != 0.0 !!!
			if(y==0) // divide by zero gives +/- infinity - so valid ... TODO check back to Go spec
				if(x==0) 
					return Math.NaN; // NaN +/-
				else
					if(x>0) return Math.POSITIVE_INFINITY;
					else return Math.NEGATIVE_INFINITY;
			if(x==0)
				if(y>0) return 0; // x==y==0.0 already handled above
				else return minusZero; // should be -0
		#end
		return x/y;
	}
	public static function floatMod(x:Float,y:Float):Float {
		if(y==0.0)
			Scheduler.panicFromHaxe("attempt to modulo float value by 0"); 
		#if ( php  ) 
			if(x==0)
				if(y>=0) return x; // to allow for -0
				else return return zero * -1; // should be -0
		#end
		return x%y;
	}

	public static inline function toUTF8length(gr:Int,s:String):Int {
		return s.length;
	}
	// return the UTF8 version of a string in a Slice
	public static function toUTF8slice(gr:Int,s:String):Slice { // TODO remove gr param
		var sl=s.length;
		var obj = Object.make(sl);
		for(i in 0...sl) {
			#if (js || php || neko ) // nullable targets
				var t:Null<Int>=s.charCodeAt(i);
				if(t==null) t=0;
				obj.set_uint8(i,t);
			#else
				obj.set_uint8(i,s.charCodeAt(i));
			#end
		}
		var ptr = Pointer.make(obj);
		var ret = new Slice(ptr,0,-1,sl,1);
		ptr=null; obj=null; // for GC
		return ret;
	}
	public static function toRawString(gr:Int,sl:Slice):String { // TODO remove gr param
		if(sl==null) return "";
		var sll=sl.len();
		if(sll==0) return "";
		var ptr = sl.itemAddr(0); // pointer to the start of the slice
		var obj = ptr.obj; // the object containing the slice data
		var off = ptr.off; // the offset to the start of that data
		var end = sll+off;
		#if cpp
			var buf=haxe.io.Bytes.alloc(sll);
			for( i in off...end) {
				buf.set(i-off,obj.get_uint8(i));
			}
			var ret=buf.getString(0,sll);
			buf=null;
			ptr=null;
			obj=null;
			return ret;
		#else
			// very slow for cpp:
			var ret = new StringBuf(); // use StringBuf for speed
			for( i in off...end ) {
				ret.addChar( obj.get_uint8(i) );
			}
			ptr=null;
			obj=null;
			var s=ret.toString();
			ret=null;
			return s;
		#end
	}

	public static function toHaxeParam(v:Dynamic):Dynamic { // TODO optimize if we know it is a function or string
		if(v==null) return null;
		if(Std.is(v,Interface)){
			if(Std.is(v.val,Closure) && v.typ!=-1){ // a closure not made by hx.CallbackFunc
				return v.val.buildCallbackFn();
			}else{
				v = v.val;
				return toHaxeParam(v);
			}
		}
		if(Std.is(v,String)){
			v=toHaxeString(v);
		}
		// TODO !
		//if(GOint64.is(v)) { 
		//	v=cast(v,haxe.Int64);
		//}
		return v;
	}
	
	public static #if (cpp || neko || php) inline #end function toHaxeString(v:String):String {
		#if !( cpp || neko || php ) // need to translate back to UTF16 when passing back to Haxe
			var sli:Slice=new Slice(Pointer.make(Object.make(v.length)),0,-1,v.length,1);
			var ptr:Pointer;
			for(i in 0...v.length){
				//if(v.charCodeAt(i)>0xff) return v; // probably already encoded as UTF-16
				ptr=sli.itemAddr(i);
				ptr.store_uint8(v.charCodeAt(i));
			}
			var slr = Go_haxegoruntime_UUTTFF8toRRunes.callFromRT(0,sli);
			var slo = Go_haxegoruntime_RRunesTToUUTTFF16.callFromRT(0,slr);
			v="";
			for(i in 0...slo.len()) {
				ptr = slo.itemAddr(i);
				v += String.fromCharCode( ptr.load_uint16() );
			}
			ptr=null;
			slr=null;
			slo=null;
		#end
		return v;
	}

	public static #if (cpp || neko || php) inline #end function fromHaxeString(v:String):String {
		#if !( cpp || neko || php ) // need to translate from UTF16 to UTF8 when passing back to Go
			#if (js || php) if(v==null) return ""; #end
			var sli:Slice=new Slice(Pointer.make(Object.make(v.length<<1)),0,-1,v.length,2);
			var ptr:Pointer;
			for(i in 0...v.length){
				ptr=sli.itemAddr(i);
				ptr.store_uint16(v.charCodeAt(i));
			}
			var slr = Go_haxegoruntime_UUTTFF16toRRunes.callFromRT(0,sli);
			var slo = Go_haxegoruntime_RRunesTToUUTTFF8.callFromRT(0,slr);
			v="";
			for(i in 0...slo.len()) {
				ptr=slo.itemAddr(i);
				v += String.fromCharCode( ptr.load_uint8() );
			}
			ptr=null;
			slr=null;
			slo=null;
		#end
		return v;
	}

	public static function checkTuple(t:Dynamic):Dynamic {
		if(t==null) Scheduler.panicFromHaxe("tuple returned from function or range is null");
		return t;
	}

	public static function stringAt(s:String,i:Int):Int{
		var c = s.charCodeAt(i);
		if(c==null) 
			Scheduler.panicFromHaxe("string index out of range");
		return toUint8(c);
	}
	public static function stringAtOK(s:String,i:Int):Dynamic {
		var c = s.charCodeAt(i);
		if(c==null)
			return {r0:0,r1:false};
		else 
			return {r0:toUint8(c),r1:true};
	}
	public static function isEqualDynamic(a:Dynamic,b:Dynamic):Bool{
		if(a==b) 
			return true; // simple equality
		if(Reflect.isObject(a) && Reflect.isObject(b)){
			if(Std.is(a,String) && Std.is(b,String))
				return a==b;
			if(Std.is(a,Pointer) && Std.is(b,Pointer))
				return Pointer.isEqual(a,b); // could still be equal within a Pointer type     
			if(Std.is(a,Complex) && Std.is(b,Complex))
				return Complex.eq(a,b);
			if(Std.is(a,Interface) && Std.is(b,Interface))
				return Interface.isEqual(a,b); // interfaces   
			#if !abstractobjects
			if(Std.is(a,Object) && Std.is(b,Object))
				return a.isEqual(0,b,0);
			#end

			if(Std.is(a,GOmap)||Std.is(a,Closure)||Std.is(a,Slice)||Std.is(a,Channel)) {
				Scheduler.panicFromHaxe("Unsupported isEqualDynamic haxe type:"+a);
				return false; // never equal
			}
	  
			// assume GOint64 - Std.is() does not work for abstract types
			return GOint64.compare(a,b)==0;
		}
		if(Std.is(a,Float)&&Std.is(b,Float)){
			return GOint64.compare(
				Go_haxegoruntime_FFloat64bits.callFromRT(0,a),
				Go_haxegoruntime_FFloat64bits.callFromRT(0,b))==0;
		}
		return false;	
	}
	public static function stringFromRune(rune:Int):String{
		var _ret:String="";
		var _r:Slice=Go_haxegoruntime_RRune2RRaw.callFromRT(0,rune);
		var _ptr:Pointer;
		var rl=_r.len();
		for(_i in 0...rl){
			_ptr=_r.itemAddr(_i);
			_ret+=String.fromCharCode(_ptr.load_int32());
		}
		_ptr=null;
		_r=null;
		return	_ret;
	}
}
`)
	objClass := `

// Object code
// a single type of Go object
@:keep
#if abstractobjects
abstract Object (haxe.ds.Vector<Dynamic>) to haxe.ds.Vector<Dynamic> from haxe.ds.Vector<Dynamic> {
#else
class Object { 
#end
	public static inline function make(size:Int,?byts:haxe.io.Bytes):Object {
		#if abstractobjects
			var ret = new haxe.ds.Vector<Dynamic>(size);
			if(byts!=null) 
				for(i in 0 ... byts.length) 
					ret[i] = byts.get(i);
			return ret;
		#else
			return new Object(size,byts);
		#end
	}

	#if ((js || cpp || neko) && fullunsafe) 
		public static var nativeFloats:Bool=true; 
	#else
		public static var nativeFloats:Bool=false; 	
	#end

#if abstractobjects
	public inline function len():Int {
		return this.length;
	}
	public inline function uniqueRef():Int {
		Console.naclWrite("uniqueRef!");
		return 0; // TODO
	}
#else
	private var dVec4:haxe.ds.Vector<Dynamic>; // on 4-byte boundaries 
	#if (js && fullunsafe) // native memory access (nearly)
		private var arrayBuffer:js.html.ArrayBuffer;
		private var dView:js.html.DataView;
	#elseif !fullunsafe	// Simple! 1 address per byte, non-Int types are always on 4-byte
		private var iVec:haxe.ds.Vector<Int>; 
	#else // fullunsafe position is to allow unsafe pointers, and therefore run slowly...
		private var byts:haxe.io.Bytes;
	#end
	public inline function len():Int {
		return length;
	}
	private var length:Int;
	public inline function uniqueRef():Int {
		return uRef;
	}
	private var uRef:Int; // to give pointers a unique numerical value
#end
	private static var uniqueCount:Int=0;
	#if godebug
		public static var memory = new Map<Int,Object>();
	#end

#if abstractobjects
  	inline public function new(v:haxe.ds.Vector<Dynamic>) {
    	this = v;
  	}
#else
	public function new(byteSize:Int,?bytes:haxe.io.Bytes){ // size is in bytes
		dVec4 = new haxe.ds.Vector<Dynamic>(1+(byteSize>>2)); // +1 to make sure non-zero
		if(bytes!=null) byteSize = bytes.length;
		#if (js && fullunsafe)
			arrayBuffer = new js.html.ArrayBuffer(byteSize);
			if(byteSize>0)
				dView = new js.html.DataView(arrayBuffer,0,byteSize); // complains if size is 0, TODO review
			if(bytes!=null)
				for(i in 0 ... byteSize) 
					set_uint8(i, bytes.get(i));
		#elseif !fullunsafe
			iVec = new haxe.ds.Vector<Int>(byteSize);
			if(bytes!=null)
				for(i in 0 ... byteSize) 
					iVec[i] = bytes.get(i);
		#else
			if(bytes==null)	{
				byts = haxe.io.Bytes.alloc(byteSize);
				//byts.fill(0,byteSize,0); 
			} else byts = bytes;
		#end
		length = byteSize;
		uniqueCount += 1;
		uRef = uniqueCount;
		#if godebug
			memory.set(uniqueRef(),this);
		#end
		#if nonulltests
			#if (js || php || neko ) 
				for(i in 0...length)
					set_uint8(i,0); 
			#end
		#end
	}
#end
	public function getBytes():haxe.io.Bytes {
		#if (js && fullunsafe)
			var byts = haxe.io.Bytes.alloc(length);
			for(i in 0 ... length) 
				byts.set(i,get_uint8(i));
		#elseif abstractobjects
			var byts = haxe.io.Bytes.alloc(this.length);
			for(i in 0 ... this.length) 
				byts.set(i,this[i]);
		#elseif !fullunsafe
			var byts = haxe.io.Bytes.alloc(length);
			for(i in 0 ... length) 
				byts.set(i,iVec[i]);
		#else
			// the byts field already exists
		#end
		return byts;
	}
	public function clear():Object {
		for(i in 0...this.length){
			set_uint8(i,0);
			if(i&3==0) set(i,null);
		}
		return this; // to allow use without a temp var
	}
	public function isEqual(off:Int,target:Object,tgtOff:Int):Bool { // TODO check if correct, used by interface{} value comparison
		if((this.length-off)!=(target.len()-tgtOff)) return false;
		for(i in 0...(this.length-off)) {
			if((i+off)&3==0){
				var a:Dynamic=this.get(i+off);
				var b:Dynamic=target.get(i+tgtOff);
				if(!Force.isEqualDynamic(a,b)) return false;
				/* code like that below all now in isEqualDynamic
				if(a!=b){
					if(Reflect.isObject(a)&&Reflect.isObject(b)) { // also deals with one being null
						if(Std.is(a,Pointer)&&Std.is(b,Pointer)) {
							if(!Pointer.isEqual(a,b) )
								return false;
						} else {
							if(Std.is(a,Interface)&&Std.is(b,Interface)){
								if(!Interface.isEqual(a,b))
									return false;
							} else {
								if(Std.is(a,GOmap)||Std.is(b,GOmap))
									return false; //maps are never equal
								else
									if(GOint64.compare(a,b)!=0) // Assume Goint64
										return false;
							}
						}
					} else
						return false;
				}
				*/
			}
 			#if fullunsafe
				if(this.get_uint8(i+off)!=target.get_uint8(i+tgtOff))
					return false;
 			#elseif abstractobjects
 				var ths=this.get(i+off);
 				var tgt=target.get(i+tgtOff);
				if(ths!=tgt) 
					if(ths==null)
						if(Std.is(tgt,Int))
							return cast(tgt,Int)!=0;
						else
							if(Std.is(tgt,Float))
								return cast(tgt,Float)!=0;
							else
								if(Std.is(tgt,Bool))
									return cast(tgt,Bool)!=false;
								else
									return false;
					else
						if(tgt==null)
							if(Std.is(ths,Int))
								return cast(ths,Int)!=0;
							else
								if(Std.is(ths,Float))
									return cast(ths,Float)!=0;
								else
									if(Std.is(ths,Bool))
										return cast(ths,Bool)!=false;
									else
										return false;
						else
							return false;
			#else
				if(this.get_uint32(i+off)!=target.get_uint32(i+tgtOff))
					return false;
			#end
		}
		return true;
	}
	public static function objBlit(src:Object,srcPos:Int,dest:Object,destPos:Int,size:Int):Void{
		if(size>0&&src!=null) {
`
	if pogo.DebugFlag {
		objClass += `
		if(!Std.is(src,Object)) { 
			Scheduler.panicFromHaxe("Object.objBlt() src parameter is not an Object - Value: "+Std.string(src)+" Type: "+Type.typeof(src));
			return;
		}
		if(!Std.is(dest,Object)) { 
			Scheduler.panicFromHaxe("Object.objBlt() dest parameter is not an Object - Value: "+Std.string(dest)+" Type: "+Type.typeof(dest));
			return;
		}
		if(srcPos<0 || srcPos>=src.length){
			Scheduler.panicFromHaxe("Object.objBlt() srcPos out-of-range - Value= "+Std.string(srcPos)+
				" src.length= "+Std.string(src.length));
			return;			
		}
		if(destPos<0 || destPos>=dest.length){
			Scheduler.panicFromHaxe("Object.objBlt() destPos out-of-range - Value= "+Std.string(destPos)+
				" dest.length= "+Std.string(dest.length));
			return;			
		}
		//if(size>(src.length-srcPos) ) size = src.length-srcPos ; // TODO review why this defensive code is needed
		if(size<0 || size > (dest.length-destPos) || size>(src.length-srcPos) ) {
			Scheduler.panicFromHaxe("Object.objBlt() size out-of-range - Value: "+Std.string(size)+
				" DestSize: "+Std.string(dest.length-destPos)+
				" SrcSize: "+Std.string(src.length-srcPos));
			return;			
		}
`
	} else {
		/*objClass += `
		if(size>(src.length-srcPos) ) size = src.length-srcPos ; // TODO review why this defensive code is needed
		`*/
	}
	objClass += `
		#if fullunsafe //(js && fullunsafe)
			if((size&3==0)&&(srcPos&3==0)&&(destPos&3==0)) {
				var i:Int=0;
				var s:Int=srcPos;
				var d:Int=destPos;
				while(i<size){
					dest.set_uint32(d,src.get_uint32(s)); 
					dest.set(d,src.get(s));
					i+=4;
					s+=4;
					d+=4;
				}
			}
			else{
				var s:Int=srcPos;
				var d:Int=destPos;
				for(i in 0...size) {
					dest.set_uint8(d,src.get_uint8(s));
					if(((size-i)>3)&&(s&3)==0){ 
						dest.set(d,src.get(s));
					}
					s+=1;
					d+=1;
				}
			}
		#elseif abstractobjects
			haxe.ds.Vector.blit(src,srcPos, dest, destPos, size); 
		#else //if !fullunsafe
			if((size>>2)>0)
				haxe.ds.Vector.blit(src.dVec4,srcPos>>2, dest.dVec4, destPos>>2, size>>2); 
			haxe.ds.Vector.blit(src.iVec,srcPos, dest.iVec, destPos, size); 
		#end
		} // end of: if(size>0&&src!=null) {
	}
	public inline function get_object(size:Int,from:Int):Object { // TODO SubObj class that is effectively a pointer?
		var so:Object = make(size);
		objBlit(this,from, so, 0, size); 
		return so;
	}
	public inline function set_object(size:Int, to:Int, from:Object):Void {
		//#if php
		//	if(!Std.is(from,Object)) { 
		//		Scheduler.panicFromHaxe("Object.set_object() from parameter is not an Object - Value: "+Std.string(from)+" Type: "+Type.typeof(from));
		//		return; // treat as null object (seen examples have been integer 0)
		//	}
		//#end
		objBlit(from,0,this,to,size);
	}
	public inline function copy():Object{
		return get_object(len(),0);
	}
	public inline function get(i:Int):Dynamic {
		#if abstractobjects
			return this[i];
		#else
			return dVec4[i>>2];
		#end
	}
	public inline function get_bool(i:Int):Bool { 
		#if (js && fullunsafe)
			return dView.getUint8(i)==0?false:true;
		#elseif abstractobjects
			if(this[i]==null) return false; 
			return this[i];
		#elseif !fullunsafe
			var r:Int=iVec[i]; 
			#if (js || php || neko ) 
				return r==null?false:(r==0?false:true); 
			#else 
				return r==0?false:true; 
			#end
		#else
			return byts.get(i)==0?false:true;
		#end
	}
	public inline function get_int8(i:Int):Int { 
		#if (js && fullunsafe)
			return dView.getInt8(i);
		#elseif abstractobjects
			#if (js || php || neko ) return this[i]==null?0:0|this[i]; #else return this[i]; #end
		#elseif !fullunsafe
			#if ((js || php || neko )&&!nonulltests) return iVec[i]==null?0:0|iVec[i]; #else return iVec[i]; #end
		#else
			return Force.toInt8(byts.get(i));
		#end
	}
	public inline function get_int16(i:Int):Int { 
		#if (js && fullunsafe)
			return dView.getInt16(i,true); // little-endian
		#elseif abstractobjects
			#if (js || php || neko ) return this[i]==null?0:0|this[i]; #else return this[i]; #end
		#elseif !fullunsafe
			#if ((js || php || neko )&&!nonulltests) return iVec[i]==null?0:0|iVec[i]; #else return iVec[i]; #end
		#else
			return Force.toInt16((get_uint8(i+1)<<8)|get_uint8(i)); // little end 1st
		#end
	}
	public inline function get_int32(i:Int):Int {
		#if (js && fullunsafe)
			return dView.getInt32(i,true); // little-endian
		#elseif abstractobjects
			#if (js || php || neko ) return this[i]==null?0:0|this[i]; #else return this[i]; #end
		#elseif !fullunsafe
			#if ((js || php || neko )&&!nonulltests) return iVec[i]==null?0:0|iVec[i]; #else return iVec[i]; #end
		#else
			return Force.toInt32((get_uint16(i+2)<<16)|get_uint16(i)); // little end 1st			
		#end
	}
	public inline function get_int64(i:Int):GOint64 {
		#if !fullunsafe
			if(get(i)==null) return GOint64.ofInt(0);	
			return get(i); 
		#else
			return Force.toInt64(GOint64.make(get_uint32(i+4),get_uint32(i)));
		#end
	} 
	public inline function get_uint8(i:Int):Int { 
		#if (js && fullunsafe)
			return dView.getUint8(i);
		#elseif abstractobjects
			#if (js || php || neko ) return this[i]==null?0:0|this[i]; #else return this[i]; #end
		#elseif !fullunsafe
			#if ((js || php || neko )&&!nonulltests) return iVec[i]==null?0:0|iVec[i]; #else return iVec[i]; #end
		#else 
			return Force.toUint8(byts.get(i));
		#end
	}
	public inline function get_uint16(i:Int):Int {
		#if (js && fullunsafe)
			return dView.getUint16(i,true); // little-endian
		#elseif abstractobjects
			#if (js || php || neko ) return this[i]==null?0:0|this[i]; #else return this[i]; #end
		#elseif !fullunsafe
			#if ((js || php || neko )&&!nonulltests) return iVec[i]==null?0:0|iVec[i]; #else return iVec[i]; #end
		#else
			return Force.toUint16((get_uint8(i+1)<<8)|get_uint8(i)); // little end 1st
		#end
	}
	public inline function get_uint32(i:Int):Int {
		#if (js && fullunsafe)
			return dView.getUint32(i,true); // little-endian
		#elseif abstractobjects
			#if (js || php || neko ) return this[i]==null?0:0|this[i]; #else return this[i]; #end
		#elseif !fullunsafe
			#if ((js || php || neko )&&!nonulltests) return iVec[i]==null?0:0|iVec[i]; #else return iVec[i]; #end
		#else
			return Force.toUint32((get_uint16(i+2)<<16)|get_uint16(i)); // little end 1st
		#end
	}
	public inline function get_uint64(i:Int):GOint64 { 
		#if !fullunsafe
			if(get(i)==null) return GOint64.ofInt(0); 
			return get(i); 
		#else
			return Force.toUint64(GOint64.make(get_uint32(i+4),get_uint32(i)));
		#end
	} 
	public inline function get_uintptr(i:Int):Dynamic { // uintptr holds Haxe objects
		// TODO consider some type of read-from-mem if Dynamic type is Int 
		return get(i); 
	} 
	public inline function get_float32(i:Int):Float { 
		#if (js && fullunsafe)
			return dView.getFloat32(i,true); // little-endian
		#elseif !fullunsafe
			return get(i)==null?0.0:get(i); 
		#else 
			return byts.getFloat(i); // Go_haxegoruntime_FFloat32frombits.callFromRT(0,get_uint32(i)); 
		#end
	}
	public inline function get_float64(i:Int):Float { 
		#if (js && fullunsafe)
			return dView.getFloat64(i,true); // little-endian
		#elseif !fullunsafe
			return get(i)==null?0.0:get(i); 
		#else
			return byts.getDouble(i); // Go_haxegoruntime_FFloat64frombits.callFromRT(0,get_uint64(i)); 		
		#end
	}
	public inline function get_complex64(i:Int):Complex {
		// TODO optimize for dataview & unsafe
		var r:Complex=get(i); 
		return r==null?new Complex(0.0,0.0):r;			
	}
	public inline function get_complex128(i:Int):Complex { 
		// TODO optimize for dataview & unsafe
		var r:Complex=get(i); 
		return r==null?new Complex(0.0,0.0):r;			
	}
	public inline function get_string(i:Int):String { 
		var r=get(i); 
		return r==null?"":Std.string(r);
	}
	public inline function set(i:Int,v:Dynamic):Void { 
		#if abstractobjects
			this[i]=v;
		#else
			dVec4[i>>2]=v;
		#end
	}
	public inline function set_bool(i:Int,v:Bool):Void { 
		#if (js && fullunsafe)
			dView.setUint8(i,v?1:0);
		#elseif abstractobjects
			set(i,v);//this[i]=v?1:null;
		#elseif !fullunsafe
			iVec[i]=v?1:0;
			#if ((js || php || neko ) &&!nonulltests)
				if(iVec[i]==0) iVec[i]=null; 
			#end
		#else
			byts.set(i,v?1:0); 
		#end
	} 
	public inline function set_int8(i:Int,v:Int):Void { 
		#if (js && fullunsafe)
			dView.setInt8(i,v);
		#elseif abstractobjects
			set(i,v);//this[i]=v==0?null:v;
		#elseif !fullunsafe
			iVec[i]=v;
			#if ((js || php || neko ) &&!nonulltests)
				if(iVec[i]==0) iVec[i]=null; 
			#end
		#else
			byts.set(i,v&0xff); 
		#end
	}
	public inline function set_int16(i:Int,v:Int):Void { 
		#if (js && fullunsafe)
			dView.setInt16(i,v,true); // little-endian
		#elseif abstractobjects
			set(i,v);//this[i]=v==0?null:v;
		#elseif !fullunsafe
			iVec[i]=v;
			#if ((js || php || neko ) &&!nonulltests)
				if(iVec[i]==0) iVec[i]=null; 
			#end
		#else
			set_int8(i,v);
			set_int8(i+1,v>>8);
		#end
	}
	public inline function set_int32(i:Int,v:Int):Void { 
		#if (js && fullunsafe)
			dView.setInt32(i,v,true); // little-endian
		#elseif abstractobjects
			set(i,v);//this[i]=v==0?null:v;
		#elseif !fullunsafe
			#if ((js || php || neko ) &&!nonulltests)
				iVec[i]=v==0?null:v; 
			#else
				iVec[i]=v;
			#end
		#else
			set_int16(i,v);
			set_int16(i+2,v>>16); 
		#end
	}
	public inline function set_int64(i:Int,v:GOint64):Void { 
		#if !fullunsafe
			if(GOint64.isZero(v)) 	set(i,null);
			else					set(i,v);  
		#else
			set_uint32(i,GOint64.getLow(v));
			set_uint32(i+4,GOint64.getHigh(v));
		#end
	} 
	public inline function set_uint8(i:Int,v:Int):Void { 
		#if (js && fullunsafe)
			dView.setUint8(i,v);
		#elseif abstractobjects
			set(i,v);//this[i]=v==0?null:v;
		#elseif !fullunsafe
			iVec[i]=v;
			#if ((js || php || neko ) &&!nonulltests)
				if(iVec[i]==0) iVec[i]=null; 
			#end
		#else
			byts.set(i,v&0xff);
		#end
	}
	public inline function set_uint16(i:Int,v:Int):Void { 
		#if (js && fullunsafe)
			dView.setUint16(i,v,true); // little-endian
		#elseif abstractobjects
			set(i,v);//this[i]=v==0?null:v;
		#elseif !fullunsafe
			iVec[i]=v;
			#if ((js || php || neko ) &&!nonulltests)
				if(iVec[i]==0) iVec[i]=null; 
			#end
		#else
			set_uint8(i,v); 
			set_uint8(i+1,v>>8); 
		#end
	}
	public inline function set_uint32(i:Int,v:Int):Void { 
		#if (js && fullunsafe)
			dView.setUint32(i,v,true); // little-endian
		#elseif abstractobjects
			set(i,v);//this[i]=v==0?null:v;
		#elseif !fullunsafe
			iVec[i]=v;
			#if ((js || php || neko ) &&!nonulltests)
				if(iVec[i]==0) iVec[i]=null; 
			#end
		#else
			set_uint16(i,v);
			set_uint16(i+2,v>>16); 
		#end
	}
	public inline function set_uint64(i:Int,v:GOint64):Void { 
		#if !fullunsafe
			if(GOint64.isZero(v)) 	set(i,null);
			else					set(i,v);  
		#else
			set_uint32(i,GOint64.getLow(v));
			set_uint32(i+4,GOint64.getHigh(v));
		#end
	} 
	public inline function set_uintptr(i:Int,v:Dynamic):Void { 
		if(Std.is(v,Int)) {
			set(i,Force.toUint32(v)); // make sure we only store 32 bits if int
			#if !abstractobjects
				set_uint32(i,v); // also write through to ordinary memory if the type is Int
			#end
			return;
		}
		set(i,v);
		#if !abstractobjects
			set_uint32(i,0); // value overwritten
		#end
	}
	public static var MinFloat64:Float = -1.797693134862315708145274237317043567981e+308; // 2**1023 * (2**53 - 1) / 2**52
	public inline function set_float32(i:Int,v:Float):Void {
		#if (js && fullunsafe)
			dView.setFloat32(i,v,true); // little-endian
		#elseif !fullunsafe
			v=Force.toFloat32(v);
			#if (js || php || neko ) 
				if(v==0.0) {
					#if !php
					var t:Float=1/v; // result is +/- infinity
					if(t>MinFloat64) // ie not -0
					#end
						v=null; 
				}
			#end
			set(i,v);
		#else 
			#if (cpp||neko)
				byts.setFloat(i,v);
			#else
				set_uint32(i,Go_haxegoruntime_FFloat32bits.callFromRT(0,v));
			#end 
		#end	
	}
	public inline function set_float64(i:Int,v:Float):Void {
	 	#if (js && fullunsafe)
			dView.setFloat64(i,v,true); // little-endian
		#elseif !fullunsafe
			#if (js || php || neko ) 
				if(v==0.0) {
					#if !php
					var t:Float=1/v; // result is +/- infinity
					if(t>MinFloat64) // ie not -0
					#end
						v=null;
				} 
			#end
			set(i,v);
		#else
			#if (cpp||neko)
				byts.setDouble(i,v);
			#else
				set_uint64(i,Go_haxegoruntime_FFloat64bits.callFromRT(0,v));
			#end 
		#end	
	}
	
	public inline function set_complex64(i:Int,v:Complex):Void { 
		if(v.real==0 && v.imag==0) set(i,null);
		else set(i,v); // TODO review
	} 
	public inline function set_complex128(i:Int,v:Complex):Void { 
		if(v.real==0 && v.imag==0) set(i,null);
		else set(i,v); // TODO review
	} 
	public inline function set_string(i:Int,v:String):Void { 
		if(v=="") set(i,null);
		else set(i,v); 
	}
	private static function str(v:Dynamic):String{
		return v==null?"nil":Std.is(v,Pointer)?v.toUniqueVal():Std.string(v);
	}
	public function toString(addr:Int=0,count:Int=-1):String{
		if(count==-1) count=this.length;
		if(addr<0) addr=0;
		if(count<0 || count>(this.length-addr)) count = this.length-addr;
		var ret:String =  "{";
		for(i in 0...count){
			if(i>0) ret = ret + ",";
			#if abstractobjects
				ret += str(get(addr));
			#else
				if((addr)&3==0) ret += str(get(addr));
				ret = ret+"<"+Std.string(get_uint8(addr))+">";
			#end
			addr = addr+1;
		}
		return ret+"}";
	}
}
`
	pogo.WriteAsClass("Object", objClass)

	ptrClass := `
@:keep
class Pointer { 
	public var obj:Object; // reference to the object holding the value
	public var off:Int; // the offset into the object, if any 
`
	if pogo.DebugFlag {
		ptrClass += `
	public function new(from:Object,offset:Int){
		if(from==null) Scheduler.panicFromHaxe("attempt to make a new Pointer from a nil object");
`
	} else {
		ptrClass += `
	public #if inlinepointers inline #end function new(from:Object,offset:Int){
`
	}
	ptrClass += `		obj = from; 
	off = offset;
	}
	public function len(){
		if(obj==null) return 0;
		return obj.len()-off;
	}
	public function hashInt():Int {
		var ur:Int=obj.uniqueRef();
		var r = ((ur&0xffff)<<16) | (off&0xffff); // hash value for a pointer
		//trace("DEBUG Pointer.hashInt="+Std.string(r)+" this="+this.toUniqueVal());
		return r;
	}
`
	if pogo.DebugFlag {
		ptrClass += `	public static function check(p:Dynamic):Pointer {
		if(p==null) {
			Scheduler.panicFromHaxe("nil pointer de-reference");
			return null;
		}
		if(Std.is(p,Pointer)) return p;
		if(Std.is(p,Int)) 
			Scheduler.panicFromHaxe("TARDISgo/Haxe implementation cannot convert from uintptr to pointer");
		Scheduler.panicFromHaxe("non-Pointer cannot be used as a pointer");
		return null;
	}
`
	} else { // TODO null test could be removed in some future NoChecking mode maybe?
		/*	ptrClass += `	public inline static function check(p:Pointer):Pointer {
			return p;
		}`*/
	}
	pogo.WriteAsClass("Pointer", ptrClass+
		`	public static function isEqual(p1:Pointer,p2:Pointer):Bool {
		if(p1==p2) return true; // simple case of being the same haxe object
		if(p1==null || p2==null) return false; // one of them is null (if above handles both null)
		if(p1.obj.uniqueRef()==p2.obj.uniqueRef() && p1.off==p2.off) return true; // point to same object & offset
		return false;
	}
	public static inline function make(from:Object):Pointer {
		return new Pointer(from,0);
	} 
	public #if inlinepointers inline #end function addr(byteOffset:Int):Pointer {
		return byteOffset==0?this:new Pointer(obj,off+byteOffset);
	}
	public inline function fieldAddr(byteOffset:Int):Pointer {
		return addr(byteOffset);
	}
	public inline function copy():Pointer {
		return this;
	}
	public #if inlinepointers inline #end function load_object(sz:Int):Object { 
		return obj.get_object(sz,off);
	}
	public #if inlinepointers inline #end function load():Dynamic {
		return obj.get(off);
	}
	public #if inlinepointers inline #end function load_bool():Bool { 
		return obj.get_bool(off);
	}
	public #if inlinepointers inline #end function load_int8():Int { 
		return obj.get_int8(off);
	}
	public #if inlinepointers inline #end function load_int16():Int { 
		return obj.get_int16(off);
	}
	public #if inlinepointers inline #end function load_int32():Int {
		return obj.get_int32(off);
	}
	public #if inlinepointers inline #end function load_int64():GOint64 { 
		return obj.get_int64(off);
	} 
	public #if inlinepointers inline #end function load_uint8():Int { 
		return obj.get_uint8(off);
	}
	public #if inlinepointers inline #end function load_uint16():Int {
		return obj.get_uint16(off);
	}
	public #if inlinepointers inline #end function load_uint32():Int {
		return obj.get_uint32(off);
	}
	public #if inlinepointers inline #end function load_uint64():GOint64 { 
		return obj.get_uint64(off);
	} 
	public #if inlinepointers inline #end function load_uintptr():Dynamic { 
		return obj.get_uintptr(off);
	} 
	public #if inlinepointers inline #end function load_float32():Float { 
		return obj.get_float32(off);
	}
	public #if inlinepointers inline #end function load_float64():Float { 
		return obj.get_float64(off);
	}
	public #if inlinepointers inline #end function load_complex64():Complex {
		return obj.get_complex64(off);
	}
	public #if inlinepointers inline #end function load_complex128():Complex { 
		return obj.get_complex128(off);
	}
	public #if inlinepointers inline #end function load_string():String { 
		return obj.get_string(off);
	}
	public #if inlinepointers inline #end function store_object(sz:Int,v:Object):Void {
		obj.set_object(sz,off,v);
	}
	public #if inlinepointers inline #end function store(v:Dynamic):Void {
		obj.set(off,v);
	}
	public #if inlinepointers inline #end function store_bool(v:Bool):Void { obj.set_bool(off,v); }
	public #if inlinepointers inline #end function store_int8(v:Int):Void { obj.set_int8(off,v); }
	public #if inlinepointers inline #end function store_int16(v:Int):Void { obj.set_int16(off,v); }
	public #if inlinepointers inline #end function store_int32(v:Int):Void { obj.set_int32(off,v); }
	public #if inlinepointers inline #end function store_int64(v:GOint64):Void { obj.set_int64(off,v); }  
	public #if inlinepointers inline #end function store_uint8(v:Int):Void { obj.set_uint8(off,v); }
	public #if inlinepointers inline #end function store_uint16(v:Int):Void { obj.set_uint16(off,v); }
	public #if inlinepointers inline #end function store_uint32(v:Int):Void { obj.set_uint32(off,v); }
	public #if inlinepointers inline #end function store_uint64(v:GOint64):Void { obj.set_uint64(off,v); } 
	public #if inlinepointers inline #end function store_uintptr(v:Dynamic):Void { obj.set_uintptr(off,v); }
	public #if inlinepointers inline #end function store_float32(v:Float):Void { obj.set_float32(off,v); }
	public #if inlinepointers inline #end function store_float64(v:Float):Void { obj.set_float64(off,v); }
	public #if inlinepointers inline #end function store_complex64(v:Complex):Void { obj.set_complex64(off,v); }
	public #if inlinepointers inline #end function store_complex128(v:Complex):Void { obj.set_complex128(off,v); }
	public #if inlinepointers inline #end function store_string(v:String):Void { obj.set_string(off,v); }
	public #if inlinepointers inline #end function toString(sz:Int=-1):String {
		return " &{ "+obj.toString(off,sz)+" } ";
	}
	public function toUniqueVal():String {
		return "&<"+Std.string(obj.uniqueRef())+":"+Std.string(off)+">";
	}
}
`)
	sliceClass := `
@:keep
class Slice {
	private var baseArray:Pointer;
	public var itemSize:Int; // for the size of each item in bytes 
	private var start:Int;
	private var end:Int;
	private var capacity:Int; // of the array, in items

	public var length(get, null):Int;
	inline function get_length():Int {
		return end-start;
	}
	public static #if inlinepointers inline #end function nullLen(s:Slice):Int{
		if(s==null) return 0;
		else return s.length;
	}
	public function new(fromArray:Pointer, low:Int, high:Int, ularraysz:Int, isz:Int) { 
		baseArray = fromArray;
		itemSize = isz;
		if(baseArray==null) {
			start = 0;
			end = 0;
			capacity = 0;
		} else {
			if( low<0 ) Scheduler.panicFromHaxe( "new Slice() low bound -ve"); 
			var ulCap = Math.floor(baseArray.len()/itemSize);
			if( ulCap < ularraysz) {
				ularraysz = ulCap; // ignore the given size & use the actual rather than panic TODO review+tidy
			//	Scheduler.panicFromHaxe("new Slice() internal error: underlying array capacity="+ulCap+
			//		" less than stated slice capacity="+ularraysz); // slices of existing data will have ulCap greater 
			}
			capacity = ularraysz; // the capacity of the array
			if(high==-1) high = ularraysz; //default upper bound is the capacity of the underlying array
			if( high > ularraysz ) Scheduler.panicFromHaxe("new Slice() high bound exceeds underlying array length"); 
			if( low>high ) Scheduler.panicFromHaxe("new Slice() low bound exceeds high bound"); 
			start = low;
			end = high;
		}
	} 
	public static function fromResource(name:String):Slice {
		return fromBytes(haxe.Resource.getBytes(name));
	}
	public static function fromBytes(res:haxe.io.Bytes):Slice {
		var obj = res==null?Object.make(0):Object.make(res.length,res); 
		var ptr = Pointer.make(obj);
		var ret = new Slice(ptr,0,-1,res==null?0:res.length,1); // []byte
		obj=null;ptr=null;
		return ret;
	}
	public function subSlice(low:Int, high:Int):Slice {
		if(high==-1) high = length; //default upper bound is the length of the current slice
		return new Slice(baseArray,low+start,high+start,capacity,itemSize);
	}
	public static function append(oldEnt:Slice,newEnt:Slice):Slice{ // TODO optimize further - heavily used
		if(oldEnt==null && newEnt==null) return null;
		if(newEnt==null || newEnt.len()==0) {
			return oldEnt; // NOTE not a clone as with the line below 
			//return new Slice(oldEnt.baseArray.addr(oldEnt.start*oldEnt.itemSize),0,oldEnt.len(),oldEnt.cap(),oldEnt.itemSize);
		}
		if(oldEnt==null) { // must create a copy rather than just return the new one
			oldEnt=new Slice(Pointer.make(Object.make(0)),0,0,0,newEnt.itemSize); // trigger newObj code below
		}
		if(oldEnt.itemSize!=newEnt.itemSize)
			Scheduler.panicFromHaxe("new Slice() internal error: itemSizes do not match");
		if(oldEnt.cap()>=(oldEnt.len()+newEnt.len())){
			var retEnt=new Slice(oldEnt.baseArray,oldEnt.start,oldEnt.end,oldEnt.capacity,oldEnt.itemSize);
			var offset=retEnt.len();
			for(i in 0...newEnt.len()){
				retEnt.end++; 
				//retEnt.itemAddr(offset+i).store_object(oldEnt.itemSize,newEnt.itemAddr(i).load_object(newEnt.itemSize));
				Object.objBlit(newEnt.baseArray.obj,newEnt.itemOff(i)+newEnt.baseArray.off,
					retEnt.baseArray.obj,retEnt.itemOff(offset+i)+retEnt.baseArray.off,oldEnt.itemSize);
			}
			oldEnt=null;newEnt=null;
			return retEnt;
		}else{
			var newLen = oldEnt.length+newEnt.len();
			var newCap = newLen+(newLen>>2); // NOTE auto-create 50pc new capacity 
			var newObj:Object = Object.make(newCap*oldEnt.itemSize);
			for(i in 0...oldEnt.length) {
				//newObj.set_object(oldEnt.itemSize,i*oldEnt.itemSize,oldEnt.itemAddr(i).load_object(oldEnt.itemSize));
				Object.objBlit(oldEnt.baseArray.obj,oldEnt.itemOff(i)+oldEnt.baseArray.off,
					newObj,i*oldEnt.itemSize,oldEnt.itemSize);
			}
			for(i in 0...newEnt.len()){
				//newObj.set_object(oldEnt.itemSize,
				//	oldEnt.length*oldEnt.itemSize+i*oldEnt.itemSize,newEnt.itemAddr(i).load_object(oldEnt.itemSize));
				Object.objBlit(newEnt.baseArray.obj,newEnt.itemOff(i)+newEnt.baseArray.off,
					newObj,(oldEnt.length*oldEnt.itemSize)+(i*oldEnt.itemSize),oldEnt.itemSize);
			}
			var ptr = Pointer.make(newObj);
			var ret = new Slice(ptr,0,newLen,newCap,oldEnt.itemSize);
			oldEnt=null;newEnt=null;newObj=null;ptr=null;
			return ret;
		}
	}
	public static function copy(target:Slice,source:Slice):Int{ 
		if(target==null) return 0;
		if(source==null) return 0;
		var copySize:Int=target.len();
		if(source.len()<copySize) 
			copySize=source.len(); 
		if(copySize==0) return 0;
		// Optimise not to create any temporary objects
		if(target.baseArray==source.baseArray){ // copy within the same slice
			if(target.start<=source.start){
				for(i in 0...copySize){
					//target.itemAddr(i).store_object(target.itemSize,source.itemAddr(i).load_object(target.itemSize));
					Object.objBlit(source.baseArray.obj,source.itemOff(i)+source.baseArray.off,
						target.baseArray.obj,target.itemOff(i)+target.baseArray.off,
						target.itemSize);
				}
			}else{
				var i = copySize-1;
				while(i>=0){
					//target.itemAddr(i).store_object(target.itemSize,source.itemAddr(i).load_object(target.itemSize));
					Object.objBlit(source.baseArray.obj,source.itemOff(i)+source.baseArray.off,
						target.baseArray.obj,target.itemOff(i)+target.baseArray.off,
						target.itemSize);
					i-=1;
				}
			}
		}else{
			for(i in 0...copySize){
				//target.itemAddr(i).store_object(target.itemSize,source.itemAddr(i).load_object(target.itemSize));
				Object.objBlit(source.baseArray.obj,source.itemOff(i)+source.baseArray.off,
					target.baseArray.obj,target.itemOff(i)+target.baseArray.off,
					target.itemSize);
			}
		}
		return copySize;
	}
	public function param(idx:Int):Dynamic { // special case for .hx pseudo functions
		var ptr=itemAddr(idx);
		var ret=ptr.load();
		ptr=null;
		return ret;
	}
	//public inline function getAt(idx:Int):Dynamic {
	//	//if (idx<0 || idx>=(end-start)) Scheduler.panicFromHaxe("Slice index out of range for getAt()");
	//	return baseArray.addr(idx+start).load();
	//}
	//public inline function setAt(idx:Int,v:Dynamic) {
	//	//if (idx<0 || idx>=(end-start)) Scheduler.panicFromHaxe("Slice index out of range for setAt()");
	//	baseArray.addr(idx+start).store(v); // this code relies on the object reference passing back
	//}
	public function len():Int {
		if(length!=end-start)  Scheduler.panicFromHaxe("Slice internal error: length!=end-start");
		return length;
	}
	public function setLen(n:Int) {
		if(n<0||n>this.cap())  Scheduler.panicFromHaxe("Slice setLen invalid:"+n);
		end = start+n;
	}
	public function cap():Int {
		// TODO remove null and capacity test when stable
		if(baseArray==null){
			if(capacity!=0) Scheduler.panicFromHaxe("Slice interal error: BaseArray==null but capacity="+capacity);
		}else{
			var ulCap = Math.floor(baseArray.len()/itemSize);
			if(capacity>ulCap) // slices of existing data will have ulCap greater 
				Scheduler.panicFromHaxe("Slice interal error: capacity="+capacity+" but underlying capacity="+ulCap);
		}
		return capacity-start;
	}
`
	if pogo.DebugFlag { // Normal range checking should cover this, so only in debug mode
		sliceClass += `
	public function itemAddr(idx:Int):Pointer {
		if (idx<0 || idx>=len()) 
			Scheduler.panicFromHaxe(
				"Slice index "+Std.string(idx)+" out of range 0 <= index < "+Std.string(len())+
				"\nSlice itemSize,capacity,start,end,baseArray: "+
				Std.string(itemSize)+","+Std.string(capacity)+","+
				Std.string(start)+","+Std.string(end)+","+Std.string(baseArray));
`
	} else { // TODO should this function be inline?
		sliceClass += `
		public #if inlinepointers inline #end function itemAddr(idx:Int):Pointer {
	`
	}
	sliceClass += `
		return new Pointer(baseArray.obj,baseArray.off+itemOff(idx));
	}
	private inline function itemOff(idx:Int):Int {
		return (idx+start)*itemSize;
	}
	public function toString():String {
		var ret:String = "Slice{[";
		var ptr:Pointer;
		if(baseArray!=null) 
			for(i in start...end) {
				if(i!=start) ret += ",";
				ptr=baseArray.addr(i*itemSize);
				ret+=ptr.toString(itemSize); // only works for basic types
			}
		ptr=null;
		return ret+"]}";
	}
}
`
	pogo.WriteAsClass("Slice", sliceClass)
	pogo.WriteAsClass("Closure", `

@:keep
class Closure { // "closure" is a keyword in PHP but solved using compiler flag  --php-prefix go  //TODO tidy names
	public var fn:Dynamic; 
	public var bds:Dynamic; // actually an anon struct

	public function new(f:Dynamic,b:Dynamic) {
		if(Std.is(f,Closure))	{
			if(!Reflect.isFunction(f.fn)) Scheduler.panicFromHaxe( "invalid function reference in existing Closure passed to make Closure(): "+f.fn);
			fn=f.fn; 
		} else{
			if(!Reflect.isFunction(f)) Scheduler.panicFromHaxe("invalid function reference passed to make Closure(): "+f); 
	 		fn=f;
		}
		if(fn==null) Scheduler.panicFromHaxe("new Closure() function has become null!"); // error test for flash/cpp TODO remove when issue resolved
		bds=b;
	}
	public function toString():String {
		var ret:String = "Closure{"+fn+",";
		if(bds!=null)
			for(i in 0...bds.length) {
				if(i!=0) ret += ",";
				ret+= bds[i];
			}
		return ret+"}";
	}
	public function methVal(t:Dynamic,v:Dynamic):Dynamic{
		return Reflect.callMethod(null, fn, [[],t,v]);
	}
	public static function callFn(cl:Closure,params:Dynamic):Dynamic {
		if(cl==null) {
			Scheduler.panicFromHaxe("attempt to call via null closure in Closure.callFn()");
			return null;
		}
		if(cl.fn==null) {
		 	Scheduler.panicFromHaxe("attempt to call null function reference in Closure.callFn()");
		 	return null;
		}
		if(!Reflect.isFunction(cl.fn)) {
			Scheduler.panicFromHaxe("invalid function reference in Closure(): "+cl.fn);
			return null;
		}
		return Reflect.callMethod(null, cl.fn, params);
	}
	// This technique is used to create callback functions
	public function buildCallbackFn():Dynamic { 
		//trace("buildCallbackFn");
		function bcf(params:Array<Dynamic>):Dynamic {
			//trace("bcf");
			if(!Go.doneInit) Go.init();
			params.insert(0,bds); // the variables bound in the closure (at final index 1)
			params.insert(0,0); // use goroutine 0 (at final index 0)
			var SF:StackFrame=Reflect.callMethod(null, fn, params); 
			while(SF._incomplete) Scheduler.runAll();
			return SF.res();
		}
		return Reflect.makeVarArgs(bcf); 
	}
}
`)
	pogo.WriteAsClass("Interface", `

class Interface { // "interface" is a keyword in PHP but solved using compiler flag  --php-prefix tgo //TODO tidy names 
	public var typ:Int; // the possibly interface type that has been cast to
	public var val:Dynamic;

	public inline function new(t:Int,v:Dynamic){
		typ=t;
		val=v; 
	}
	public function toString():String {
		var nam:String;
		#if (js || neko || php)
			if(typ==null) 
				nam="nil";
			else
				nam=TypeInfo.getName(typ);
		#else
			nam=TypeInfo.getName(typ);			
		#end
		if(val==null)
			return "Interface{nil:"+nam+"}";
		else
			if(Std.is(val,Pointer))
				return "interface{"+val.toUniqueVal()+":"+nam+"}"; // To stop recursion
			else
				return "Interface{"+Std.string(val)+":"+nam+"}";
	}
	public static function toDynamic(v:Interface):Dynamic {
		if(v==null)
			return null;
		return v.val;
	}
	public static function fromDynamic(v:Dynamic):Interface {
		if(v==null)
			return null;
		if(Std.is(v,Bool))
			return new Interface(TypeInfo.getId("bool"),v); 
		if(Std.is(v,Int))
			return new Interface(TypeInfo.getId("int"),v); 
		if(Std.is(v,Float))
			return new Interface(TypeInfo.getId("float64"),v); 
		if(Std.is(v,String))
			return new Interface(TypeInfo.getId("string"),v); 
		// TODO consider testing for other types here?
		return new Interface(TypeInfo.getId("uintptr"),v); 
	}
	public static function change(t:Int,i:Interface):Interface {
		if(i==null)	
			if(TypeInfo.isConcrete(t))  
				return new Interface(t,TypeZero.zeroValue(t)); 
			else {
				return null; // e.g. error(nil) 
			}
		else 
			if(Std.is(i,Interface)) 	
				if(TypeInfo.isConcrete(t)) 
					return new Interface(t,i.val); 
				else
					return new Interface(i.typ,i.val); // do not allow non-concrete types for Interfaces
			else {
				Scheduler.panicFromHaxe( "Can't change the Interface of a non-Interface type:"+i+" to: "+TypeInfo.getName(t));  
				return new Interface(t,TypeZero.zeroValue(t));	 //dummy value as we have hit the panic button
			}
	}
	public static function isEqual(a:Interface,b:Interface):Bool {		
	// TODO ensure this very wide definition of equality is OK 
	// TODO is another special case required for Slice/Object?
		if(a==null) 
			if(b==null) return true;
			else 		return false;
		if(b==null)		
			return false;
		if(! (TypeInfo.isIdentical(a.typ,b.typ)||TypeAssign.isAssignableTo(a.typ,b.typ)||TypeAssign.isAssignableTo(b.typ,a.typ)) ) 
			return false;	
		return Force.isEqualDynamic(a.val,b.val);
	}			
	/* from the SSA documentation:
	If AssertedType is a concrete type, TypeAssert checks whether the dynamic type in Interface X is equal to it, and if so, 
		the result of the conversion is a copy of the value in the Interface.
	If AssertedType is an Interface, TypeAssert checks whether the dynamic type of the Interface is assignable to it, 
		and if so, the result of the conversion is a copy of the Interface value X. If AssertedType is a superInterface of X.Type(), 
		the operation will fail iff the operand is nil. (Contrast with ChangeInterface, which performs no nil-check.)
	*/
	public static function assert(assTyp:Int,ifce:Interface):Dynamic{
		// TODO add code to deal with overloaded types? i.e. those created by reflect
		if(ifce==null) {
			Scheduler.panicFromHaxe( "Interface.assert null Interface");
		} else {
			if(TypeInfo.isConcrete(assTyp))	{
				if(ifce.typ==assTyp)
					return ifce.val;
				else
					Scheduler.panicFromHaxe( "concrete type assert failed: expected "+TypeInfo.getName(assTyp)+", got "+TypeInfo.getName(ifce.typ) );	
			} else {
				if(assertCache(ifce.typ,assTyp) /*ifce.typ==assTyp||Go_haxegoruntime_assertableTTo.callFromRT(0,ifce.typ,assTyp)*/ ){
					//was:TypeAssert.assertableTo(ifce.typ,assTyp)){
					return new Interface(ifce.typ,ifce.val);
				} else {
					Scheduler.panicFromHaxe( "interface type assert failed: cannot assert to "+TypeInfo.getName(assTyp)+" from "+TypeInfo.getName(ifce.typ) );
				}
			}
		}
		return null;
	}
	public static function assertOk(assTyp:Int,ifce:Interface):{r0:Dynamic,r1:Bool} {
		if(ifce==null) 
			return {r0:TypeZero.zeroValue(assTyp),r1:false};
		if(!assertCache(ifce.typ,assTyp) /*(ifce.typ==assTyp||Go_haxegoruntime_assertableTTo.callFromRT(0,ifce.typ,assTyp))*/ ) //was:TypeAssert.assertableTo(ifce.typ,assTyp)))
			return {r0:TypeZero.zeroValue(assTyp),r1:false};
		if(TypeInfo.isConcrete(assTyp))	
			return {r0:ifce.val,r1:true};
		else	
			return {r0:new Interface(ifce.typ,ifce.val),r1:true};
	}
	static var assertCacheMap = new Map<Int,Bool>();
	public static function assertCache(ifceTyp:Int,assTyp:Int):Bool {
		var key:Int= Force.toUint16(ifceTyp)<<16 | Force.toUint16(assTyp) ; // more than 65k types and we hava a problem...
		var ret:Bool;
		if(assertCacheMap.exists(key)){
			ret=assertCacheMap.get(key);
		}else{
			ret=(ifceTyp==assTyp||Go_haxegoruntime_assertableTTo.callFromRT(0,ifceTyp,assTyp));
			assertCacheMap.set(key,ret);
		}
		return ret;
	}
	static var methodCache = new Map<String,Dynamic>();
	public static function invoke(ifce:Interface,path:String,meth:String,args:Array<Dynamic>):Dynamic {
		if(ifce==null) 
			Scheduler.panicFromHaxe( "Interface.invoke null Interface"); 
		if(!Std.is(ifce,Interface)) 
			Scheduler.panicFromHaxe( "Interface.invoke on non-Interface value"); 
		var key=Std.string(ifce.typ)+":"+path+":"+meth;
		var fn:Dynamic=methodCache.get(key);
		if(fn==null) {
			fn=Go_haxegoruntime_getMMethod.callFromRT(0,ifce.typ,path,meth); //MethodTypeInfo.method(ifce.typ,meth);
			methodCache.set(key,fn);
		}
		var ret=Reflect.callMethod(null, fn, args);
		// set created objects to null for GC
		key=null;
		fn=null;
		// return what was asked for
		return ret;
	}
}
`)
	pogo.WriteAsClass("Channel", `

class Channel { //TODO check close & rangeing over a channel
var entries:Array<Dynamic>;
var max_entries:Int;
var num_entries:Int;
var oldest_entry:Int;	
var closed:Bool;
var capa:Int;
var uniqueId:Int;

static var nextId:Int=0;

public function new(how_many_entries:Int) {
	capa = how_many_entries;
	if(how_many_entries<=0)
		how_many_entries=1;
	entries = new Array<Dynamic>();
	max_entries = how_many_entries;
	oldest_entry = 0;
	num_entries = 0;
	closed = false;
	uniqueId = nextId;
	nextId++;
}
public function hasSpace():Bool {
	if(this==null) return false; // non-existant channels never have space
	if(closed) return false; // closed channels don't have space
	return num_entries < max_entries;
}
public function send(source:Dynamic):Bool {
	if(closed) Scheduler.panicFromHaxe( "attempt to send to closed channel"); 
	var next_element:Int;
	if (this.hasSpace()) {
		next_element = (oldest_entry + num_entries) % max_entries;
		num_entries++;
		entries[next_element]=source;  
		return true;
	} else
		return false;
}
public function hasNoContents():Bool { // used by channel read
	if (this==null) return true; // spec: "Receiving from a nil channel blocks forever."
	if (closed) return false; // spec: "Receiving from a closed channel always succeeds..."
	else return num_entries == 0;
}
public function hasContents():Bool { // used by select
	if (this==null) return false; // spec: "Receiving from a nil channel blocks forever."
	if (closed) return true; // spec: "Receiving from a closed channel always succeeds..."
	return num_entries != 0;
}
public function receive(zero:Dynamic):{r0:Dynamic ,r1:Bool} {
	var ret:Dynamic=zero;
	if (num_entries > 0) {
		ret=entries[oldest_entry];
		oldest_entry = (oldest_entry + 1) % max_entries;
		num_entries--;
		return {r0:ret,r1:true};
	} else
		if(closed)
			return {r0:ret,r1:false}; // spec: "Receiving from a closed channel always succeeds, immediately returning the element type's zero value."
		else {
			Scheduler.panicFromHaxe( "channel receive unreachable code!"); 
			return {r0:ret,r1:false}; //dummy value as we have hit the panic button
		}
}
public inline function len():Int { 
	return num_entries; 
}
public inline function cap():Int { 
	return capa; // give back the cap we were told
}
public function close() {
	if(this==null) Scheduler.panicFromHaxe( "attempt to close a nil channel" ); 
	closed = true;
}
public function toString():String{
	return "<ChanId:"+Std.string(uniqueId)+">";
}
}
`)
	pogo.WriteAsClass("Complex", `

class Complex {
	public var real:Float;
	public var imag:Float;
public function new(r:Float, i:Float) {
	real = r;
	imag = i;
}
public static function neg(x:Complex):Complex {
	return new Complex(0.0-x.real,0.0-x.imag);
}
public static function add(x:Complex,y:Complex):Complex {
	return new Complex(x.real+y.real,x.imag+y.imag);
}
public static function sub(x:Complex,y:Complex):Complex {
	return new Complex(x.real-y.real,x.imag-y.imag);
}
public static function mul(x:Complex,y:Complex):Complex {
	return new Complex( (x.real * y.real) - (x.imag * y.imag), (x.imag * y.real) + (x.real * y.imag));
}
public static function div(x:Complex,y:Complex):Complex {
	if( (y.real == 0.0) && (y.imag == 0.0) ){
		Scheduler.panicFromHaxe( "complex divide by zero");
		return new Complex(0.0,0.0); //dummy value as we have hit the panic button
	} else {
		return new Complex(
			((x.real * y.real) + (x.imag * y.imag)) / ((y.real * y.real) + (y.imag * y.imag)) ,
			((x.imag * y.real) - (x.real * y.imag)) / ((y.real * y.real) + (y.imag * y.imag)) );
	}
}
public static function eq(x:Complex,y:Complex):Bool { // "=="
	return (x.real == y.real) && (x.imag == y.imag);
}
public static function neq(x:Complex,y:Complex):Bool { // "!="
	return (x.real != y.real) || (x.imag != y.imag);
}
public static function toString(x:Complex):String {
	return Std.string(x.real)+"+"+Std.string(x.imag)+"i";
}
}

`)
	pogo.WriteAsClass("GOint64", `

#if ( neko || cpp || cs || java ) 
	typedef HaxeInt64Typedef = haxe.Int64; // these implementations are using native types
#else
	typedef HaxeInt64Typedef = Int64;  // use the copied and modified version of the standard library class below
	// TODO revert to haxe.Int64 when the version below (or better) reaches the released libray
#end

// this abstract type to enable correct handling for Go of HaxeInt64Typedef
abstract HaxeInt64abs(HaxeInt64Typedef) 
from HaxeInt64Typedef to HaxeInt64Typedef 
{ 
public inline function new(v:HaxeInt64Typedef) this=v;

#if !( neko || cpp || cs || java ) // allow casting to/from haxe.Int64 if using own version
  @:from
  static public function fromHI64(v:haxe.Int64) {
  	return HaxeInt64abs.make(v.high,v.low); 
  }
  @:to
  public inline function toHI64():haxe.Int64 {
  	return haxe.Int64.make(Int64.getHigh(this),Int64.getLow(this));
  }
#end

public static inline function getLow(v:HaxeInt64Typedef):Int {
	#if ( neko || cpp || cs || java )
		return v.low;
	#else
		return HaxeInt64Typedef.getLow(v);
	#end
}
public static inline function getHigh(v:HaxeInt64Typedef):Int {
	#if ( neko || cpp || cs || java )
		return v.high;
	#else
		return HaxeInt64Typedef.getHigh(v);
	#end
}

public static inline function toInt(v:HaxeInt64abs):Int {
	return HaxeInt64abs.getLow(v); // NOTE: does not throw an error if value overflows Int
}
public static inline function ofInt(v:Int):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.ofInt(v));
}
public static inline function ofUInt(v:Int):HaxeInt64abs {
	return make(0,v);
}
public static function toFloat(vp:HaxeInt64abs):Float{ // signed int64 to float (TODO auto-cast of Unsigned pos problem)
		//TODO native versions for java & cs
		var v:HaxeInt64Typedef=vp;
		var isNegVal:Bool=false;
		if(isNeg(v)) {
			if(compare(v,make(0x80000000,0))==0) return -9223372036854775808.0; // most -ve value can't be made +ve
			isNegVal=true;
			v=neg(v);	
		}
		var ret:Float=toUFloat(v);
		if(isNegVal) return -ret;
		return ret;
}
public static function toUFloat(vp:HaxeInt64abs):Float{ // unsigned int64 to float
		//TODO native versions for java & cs
		var v:HaxeInt64Typedef=vp;
		var ret:Float=0.0;
		var multiplier:Float=1.0;
		var one:HaxeInt64abs=make(0,1);
		for(i in 0...64) { 
			if(!isZero(and(v,one)))
	 			ret += multiplier;
			multiplier *= 2.0;
			v=ushr(v,1);
		}
		return ret;
}
public static function ofFloat(v:Float):HaxeInt64abs { // float to signed int64 (TODO auto-cast of Unsigned is a posible problem)
		//TODO native versions for java & cs
		if(v==0.0) return make(0,0); 
		if(Math.isNaN(v)) return make(0x80000000,0); // largest -ve number is returned by Go in this situation
		var isNegVal:Bool=false;
		if(v<0.0){
			isNegVal=true;
			v = -v;
		} 
		if(v<2147483647.0) { // optimization: if just a small integer, don't do the full conversion code below
			if(isNegVal) 	return new HaxeInt64abs(HaxeInt64Typedef.neg(HaxeInt64Typedef.ofInt(Math.floor(v)))); // ceil?
			else			return new HaxeInt64abs(HaxeInt64Typedef.ofInt(Math.floor(v)));
		}
		if(v>9223372036854775807.0) { // number too big to encode in 63 bits 
			if(isNegVal)	return new HaxeInt64abs(HaxeInt64Typedef.make(0x80000000,0)); 			// largest -ve number
			else			return new HaxeInt64abs(HaxeInt64Typedef.make(0x7fffffff,0xffffffff)); 	// largest +ve number
		}
		var res:HaxeInt64Typedef = ofUFloat(v);
		if(isNegVal) return new HaxeInt64abs(HaxeInt64Typedef.neg(res));
		return new HaxeInt64abs(res);
}
public static function ofUFloat(v:Float):HaxeInt64abs { // float to un-signed int64 
		//TODO native versions for java & cs
		if(v<0.0){
			//Scheduler.panicFromHaxe("-ve value passed to internal haxe function ofUFloat()");
			return make(0,0); // -ve values are invalid here, so return 0
		} 
		if(Math.isNaN(v)) return make(0x80000000,0); // largest -ve number is returned by Go in this situation
		if(v<2147483647.0) { // optimization: if just a small integer, don't do the full conversion code below
			return ofInt(Math.floor(v));
		}
		if(v>18446744073709551615.0) { // number too big to encode in 64 bits 
			return new HaxeInt64abs(HaxeInt64Typedef.make(0xffffffff,0xffffffff)); 	// largest unsigned number
		}
		var f32:Float = 4294967296.0 ; // the number of combinations in 32-bits
		var f16:Float = 65536.0; // the number of combinations in 16-bits
		v = Math.ffloor(v); // remove any fractional part
		var high:Float = Math.ffloor(v/f32);
		var highTop16:Float = Math.ffloor(high/f16);
		var highBot16:Float = high-(highTop16*f16);
		var highBits:Int = Math.floor(highTop16)<<16 | Math.floor(highBot16);
		var low:Float = v-(high*f32);
		var lowTop16:Float = Math.ffloor(low/f16);
		var lowBot16:Float = low-(lowTop16*f16);
		var lowBits:Int = Math.floor(lowTop16)<<16 | Math.floor(lowBot16);
		return HaxeInt64Typedef.make(highBits,lowBits);
}
public static #if !(cs||java) inline #end function make(h:Int,l:Int):HaxeInt64abs { 
// NOTE cs & java have problems inlining the '0xffffffffL' constant
		return new HaxeInt64abs(HaxeInt64Typedef.make(h,l));
}
public static inline function toString(v:HaxeInt64abs):String {
	return HaxeInt64Typedef.toStr(v);
}
public static inline function toStr(v:HaxeInt64abs):String {
	return HaxeInt64Typedef.toStr(v);
}
public static inline function neg(v:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.neg(v));
}
public static inline function isZero(v:HaxeInt64abs):Bool {
	return HaxeInt64Typedef.isZero(v);
}
public static inline function isNeg(v:HaxeInt64abs):Bool {
	return HaxeInt64Typedef.isNeg(v);
}
public static inline function add(x:HaxeInt64abs,y:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.add(x,y));
}
public static inline function and(x:HaxeInt64abs,y:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.and(x,y));
}
private static function checkDiv(x:HaxeInt64abs,y:HaxeInt64abs,isSigned:Bool):HaxeInt64abs {
	if(HaxeInt64Typedef.isZero(y))
		Scheduler.panicFromHaxe( "attempt to divide 64-bit value by 0"); 
	if(isSigned && (HaxeInt64Typedef.compare(y,HaxeInt64Typedef.ofInt(-1))==0) && (HaxeInt64Typedef.compare(x,HaxeInt64Typedef.make(0x80000000,0))==0) ) 
	{
		//trace("checkDiv 64-bit special case");
		y=HaxeInt64Typedef.ofInt(1); // special case in the Go spec
	}
	return new HaxeInt64abs(y);
}
public static function div(x:HaxeInt64abs,y:HaxeInt64abs,isSigned:Bool):HaxeInt64abs {
	y=checkDiv(x,y,isSigned);
	if(HaxeInt64Typedef.compare(y,HaxeInt64Typedef.ofInt(1))==0) return new HaxeInt64abs(x);
	if(isSigned || (!HaxeInt64Typedef.isNeg(x) && !HaxeInt64Typedef.isNeg(y)))
		return new HaxeInt64abs(HaxeInt64Typedef.div(x,y));
	else {
		if(	HaxeInt64Typedef.isNeg(x) ) {
			if( HaxeInt64Typedef.isNeg(y) ){ // both x and y are "-ve""
				if( HaxeInt64Typedef.compare(x,y) < 0 ) { // x is more "-ve" than y, so the smaller uint   
					return new HaxeInt64abs(HaxeInt64Typedef.ofInt(0));						
				} else {
					return new HaxeInt64abs(HaxeInt64Typedef.ofInt(1));	// both have top bit set & uint(x)>uint(y)
				}
			} else { // only x is -ve
				var pt1:HaxeInt64Typedef = HaxeInt64Typedef.make(0x7FFFFFFF,0xFFFFFFFF); // the largest part of the numerator
				var pt2:HaxeInt64Typedef = HaxeInt64Typedef.and(x,pt1); // the smaller part of the numerator
				var rem:HaxeInt64Typedef = HaxeInt64Typedef.make(0,1); // the left-over bit
				rem = HaxeInt64Typedef.add(rem,HaxeInt64Typedef.mod(pt1,y));
				rem = HaxeInt64Typedef.add(rem,HaxeInt64Typedef.mod(pt2,y));
				if( HaxeInt64Typedef.ucompare(rem,y) >= 0 ) { // the remainder is >= divisor  
					rem = HaxeInt64Typedef.ofInt(1);
				} else {
					rem = HaxeInt64Typedef.ofInt(0);
				}
				pt1 = HaxeInt64Typedef.div(pt1,y);	
				pt2 = HaxeInt64Typedef.div(pt2,y);			
				return new HaxeInt64abs(HaxeInt64Typedef.add(pt1,HaxeInt64Typedef.add(pt2,rem)));	
			}
		}else{ // logically, y is "-ve"" but x is "+ve" so y>x , so any integer divide will yeild 0
				return new HaxeInt64abs(HaxeInt64Typedef.ofInt(0));	
		}
	}
}
public static function mod(x:HaxeInt64abs,y:HaxeInt64abs,isSigned:Bool):HaxeInt64abs {
	y=checkDiv(x,y,isSigned);
	if(HaxeInt64Typedef.compare(y,HaxeInt64Typedef.ofInt(1))==0) return new HaxeInt64abs(HaxeInt64Typedef.ofInt(0));
	if(isSigned)
		return new HaxeInt64abs(HaxeInt64Typedef.mod(x,y));
	else {
		return new HaxeInt64abs(sub(x,mul(div(x,y,false),y)));
	}
}
public static inline function mul(x:HaxeInt64abs,y:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.mul(x,y));
}
public static inline function or(x:HaxeInt64abs,y:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.or(x,y));
}
public static function shl(x:HaxeInt64abs,y:Int):HaxeInt64abs {
	if(y==0) return new HaxeInt64abs(x);
	if(y<0 || y>=64) // this amount of shl is not handled correcty by the underlying code
		return new HaxeInt64abs(HaxeInt64Typedef.ofInt(0));	
	else
		return new HaxeInt64abs(HaxeInt64Typedef.shl(x,y));
}
public static function shr(x:HaxeInt64abs,y:Int):HaxeInt64abs { // note, not inline
	if(y==0) return new HaxeInt64abs(x);
	if(y<0 || y>=64)
		if(isNeg(x))
			return new HaxeInt64abs(HaxeInt64Typedef.ofInt(-1));		
		else
			return new HaxeInt64abs(HaxeInt64Typedef.ofInt(0));		
	return new HaxeInt64abs(HaxeInt64Typedef.shr(x,y));
}
public static function ushr(x:HaxeInt64abs,y:Int):HaxeInt64abs { // note, not inline
	if(y==0) return new HaxeInt64abs(x);
	if(y<0 || y>=64)
		return new HaxeInt64abs(HaxeInt64Typedef.ofInt(0));		
	#if php
	if(y==32){ // error with php on 32 bit right shift for uint64, so do 2x16
		var ret:HaxeInt64Typedef = HaxeInt64Typedef.ushr(x,16);
		return new HaxeInt64abs(HaxeInt64Typedef.ushr(ret,16));
	}
	#end
	return new HaxeInt64abs(HaxeInt64Typedef.ushr(x,y));
}
public static inline function sub(x:HaxeInt64abs,y:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.sub(x,y));
}
public static inline function xor(x:HaxeInt64abs,y:HaxeInt64abs):HaxeInt64abs {
	return new HaxeInt64abs(HaxeInt64Typedef.xor(x,y));
}
public static inline function compare(x:HaxeInt64abs,y:HaxeInt64abs):Int {
	return HaxeInt64Typedef.compare(x,y);
}
public static function ucompare(x:HaxeInt64abs,y:HaxeInt64abs):Int {
	//#if cpp
	 	return HaxeInt64Typedef.ucompare(x,y);
	//#else
	// unsigned compare library code does not work properly for all platforms 
	/*was:
		if(HaxeInt64Typedef.isZero(x)) {
			if(HaxeInt64Typedef.isZero(y)) {
				return 0;
			} else {
				return -1; // any value is larger than x 
			}
		}
		if(HaxeInt64Typedef.isZero(y)) { // if we are here, we know that x is non-zero
				return 1; // any value of x is larger than y 
		}
		if(!HaxeInt64Typedef.isNeg(x)) { // x +ve
			if(!HaxeInt64Typedef.isNeg(y)){ // both +ve so normal comparison
				return HaxeInt64Typedef.compare(x,y);
			}else{ // y -ve and so larger than x
				return -1;
			}
		}else { // x -ve
			if(!HaxeInt64Typedef.isNeg(y)){ // -ve x larger than +ve y
				return 1;
			}else{ // both are -ve so the normal comparison works ok
				return HaxeInt64Typedef.compare(x,y); 
			}
		}
	*/
	//#end
}
}

	typedef GOint64 = HaxeInt64abs;

//**************** rewrite of std Haxe library function haxe.Int64 for PHP integer overflow an other errors
/*
Modify haxe.Int64.hx to work on php and fix other errors
- php integer overflow and ushr are incorrect (for 32-bits Int),
special functions now correct for these faults for Int64.
- both div and mod now have the sign correct when double-negative.
- special cases of div or mod by 0 or 1 now correct.
*/
/*
 * Copyright (C)2005-2012 Haxe Foundation
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */
class Int64 { 

	var high : Int;
	var low : Int;

	inline function new(high, low) {
		this.high = i32(high);
		this.low = i32(low);
	}

	#if php
	/*
		private function to correctly handle 32-bit integer overflow on php 
		see: http://stackoverflow.com/questions/300840/force-php-integer-overflow
	*/
	private static function i32php(value:Int):Int { 
			value = (value & untyped __php__("0xFFFFFFFF"));
 		    if ( (value & untyped __php__("0x80000000"))!=0 )
		        value = -(((~value) & untyped __php__("0xFFFFFFFF")) + 1);
		    return value;
	}
	#end

	/*
		private function to correctly handle 32-bit ushr on php
		see: https://github.com/HaxeFoundation/haxe/commit/1a878aa90708040a41b0dd59f518d83b09ede209
	*/
	private static inline function ushr32(v:Int,n:Int):Int { 
		#if php
		 	return (v >> n) & (untyped __php__("0x7fffffff") >> (n-1));
		#else
			return v>>>n;
		#end
	}

	@:extern static inline function i32(i) {
		return 
		#if !(cpp || java || cs || flash) 
			i==null?0:
		#end
		#if (js || flash8)
			i | 0;
		#elseif php
			i32php(i); // handle overflow of 32-bit integers correctly 
		#else
			i;
		#end
	}

	@:extern static inline function i32mul(a:Int,b:Int) {
		#if (php || js || flash8)
		/*
			We can't simply use i32(a*b) since we might overflow (52 bits precision in doubles)
		*/
		return i32(i32((a * (b >>> 16)) << 16) + (a * (b&0xFFFF)));
		#else
		return a * b;
		#end
	}
	
	#if as3 public #end function toString() {
		if ((high|low) == 0 )
			return "0";
		var str = "";
		var neg = false;
		var i = this;
		if( isNeg(i) ) {
			neg = true;
			i = Int64.neg(i);
		}
		var ten = ofInt(10);
		while( !isZero(i) ) {
			var r = divMod(i, ten);
			str = r.modulus.low + str; 
			i = r.quotient; 
		}
		if( neg ) str = "-" + str;
		return str;
	}

	public static inline function make( high : Int, low : Int ) : Int64 {
		return new Int64(high, low); 
	}

	public static inline function ofInt( x : Int ) : Int64 {
		return new Int64(x >> 31,x);
	}

	public static function toInt( x : Int64 ) : Int {
		if( x.high != 0 ) {
			if( x.high < 0 )
				return -toInt(neg(x));
			throw "Overflow"; //NOTE go panic not used here as it is in the Haxe libary code
		}
		return x.low; 
	}

	public static function getLow( x : Int64 ) : Int {
		return x.low;
	}

	public static function getHigh( x : Int64 ) : Int {
		return x.high;
	}

	public static function add( a : Int64, b : Int64 ) : Int64 {
		var high = i32(a.high + b.high);
		var low = i32(a.low + b.low);
		if( uicompare(low,a.low) < 0 )
			high++;
		return new Int64(high, low);
	}

	public static function sub( a : Int64, b : Int64 ) : Int64 {
		var high = i32(a.high - b.high); // i32() call required to match add
		var low = i32(a.low - b.low); // i32() call required to match add
		if( uicompare(a.low,b.low) < 0 )
			high--;
		return new Int64(high, low);
	}

	public static function mul( a : Int64, b : Int64 ) : Int64 {
		var mask = 0xFFFF;
		var al = a.low & mask, ah = ushr32(a.low , 16); 
		var bl = b.low & mask, bh = ushr32(b.low , 16); 
		var p00 = al * bl;
		var p10 = ah * bl;
		var p01 = al * bh;
		var p11 = ah * bh;
		var low = p00;
		var high = i32(p11 + ushr32(p01 , 16) + ushr32(p10 , 16));
		p01 = i32(p01 << 16); low = i32(low + p01); if( uicompare(low, p01) < 0 ) high = i32(high + 1);
		p10 = i32(p10 << 16); low = i32(low + p10); if( uicompare(low, p10) < 0 ) high = i32(high + 1);
		high = i32(high + i32mul(a.low,b.high));
		high = i32(high + i32mul(a.high,b.low));
		return new Int64(high, low);
	}

	static function divMod( modulus : Int64, divisor : Int64 ) {
		var quotient = new Int64(0, 0);
		var mask = new Int64(0, 1);
		divisor = new Int64(divisor.high, divisor.low);
		while( divisor.high >= 0 ) { 
			var cmp = ucompare(divisor, modulus);
			divisor.high = i32( i32(divisor.high << 1) | ushr32(divisor.low , 31) ); 
			divisor.low = i32(divisor.low << 1); 
			mask.high = i32( i32(mask.high << 1) | ushr32(mask.low , 31) ); 
			mask.low = i32(mask.low << 1);
			if( cmp >= 0 ) break;
		}
		while( i32(mask.low | mask.high) != 0 ) { 
			if( ucompare(modulus, divisor) >= 0 ) {
				quotient.high= i32(quotient.high | mask.high); 
				quotient.low= i32(quotient.low | mask.low); 
				modulus = sub(modulus,divisor);
			}
			mask.low = i32( ushr32(mask.low , 1) | i32(mask.high << 31) ); 
			mask.high = ushr32(mask.high , 1); 

			divisor.low = i32( ushr32(divisor.low , 1) | i32(divisor.high << 31) ); 
			divisor.high = ushr32(divisor.high , 1); 
		}
		return { quotient : quotient, modulus : modulus };
	}

	public static function div( a : Int64, b : Int64 ) : Int64 { 
		if(b.high==0) // handle special cases of 0 and 1
			switch(b.low) {
			case 0:	throw "divide by zero";  //NOTE go panic not used here as it is in the Haxe libary code
			case 1: return new Int64(a.high,a.low);
			} 
		var sign = ((a.high<0) || (b.high<0)) && (!( (a.high<0) && (b.high<0))); // make sure we get the correct sign
		if( a.high < 0 ) a = neg(a);
		if( b.high < 0 ) b = neg(b);
		var q = divMod(a, b).quotient;
		return sign ? neg(q) : q;
	}

	public static function mod( a : Int64, b : Int64 ) : Int64 {
		if(b.high==0) // handle special cases of 0 and 1
			switch(b.low) {
			case 0:	throw "modulus by zero";  //NOTE go panic not used here as it is in the Haxe libary code
			case 1: return ofInt(0);
			}
		var sign = a.high<0; // the sign of a modulus is the sign of the value being mod'ed
		if( a.high < 0 ) a = neg(a);
		if( b.high < 0 ) b = neg(b);
		var m = divMod(a, b).modulus;
		return sign ? neg(m) : m;
	}

	public static inline function shl( a : Int64, b : Int ) : Int64 {
		return if( b & 63 == 0 ) a else if( b & 63 < 32 ) new Int64( (a.high << b) | ushr32(a.low, i32(32-(b&63))), a.low << b ) else new Int64( a.low << i32(b - 32), 0 );
	}

	public static inline function shr( a : Int64, b : Int ) : Int64 {
		return if( b & 63 == 0 ) a else if( b & 63 < 32 ) new Int64( a.high >> b, ushr32(a.low,b) | (a.high << i32(32 - (b&63))) ) else new Int64( a.high >> 31, a.high >> i32(b - 32) );
	}

	public static inline function ushr( a : Int64, b : Int ) : Int64 {
		return if( b & 63 == 0 ) a else if( b & 63 < 32 ) new Int64( ushr32(a.high, b), ushr32(a.low, b) | (a.high << i32(32 - (b&63))) ) else new Int64( 0, ushr32(a.high, i32(b - 32)) );
	}

	public static inline function and( a : Int64, b : Int64 ) : Int64 {
		return new Int64( a.high & b.high, a.low & b.low );
	}

	public static inline function or( a : Int64, b : Int64 ) : Int64 {
		return new Int64( a.high | b.high, a.low | b.low );
	}

	public static inline function xor( a : Int64, b : Int64 ) : Int64 {
		return new Int64( a.high ^ b.high, a.low ^ b.low );
	}

	public static inline function neg( a : Int64 ) : Int64 {
		var high = i32(~a.high); 
		var low = i32(-a.low); 
		if( low == 0 )
			high++;
		return new Int64(high,low);
	}

	public static inline function isNeg( a : Int64 ) : Bool {
		return a.high < 0;
	}

	public static inline function isZero( a : Int64 ) : Bool {
		return (a.high | a.low) == 0;
	}

	static function uicompare( a : Int, b : Int ) {
		return a < 0 ? (b < 0 ? i32(~b - ~a) : 1) : (b < 0 ? -1 : i32(a - b));
	}

	public static inline function compare( a : Int64, b : Int64 ) : Int {
		var v = i32(i32(a.high) - i32(b.high)); 
		return if( v != 0 ) v else uicompare(a.low,b.low);
	}

	/**
		Compare two Int64 in unsigned mode.
	**/
	public static inline function ucompare( a : Int64, b : Int64 ) : Int {
		var v = uicompare(a.high,b.high);
		return if( v != 0 ) v else uicompare(a.low, b.low);
	}

	public static inline function toStr( a : Int64 ) : String {
		return a.toString();
	}

}
//**************** END REWRITE of haxe.Int64 for php and to correct errors

`)
	pogo.WriteAsClass("StackFrameBasis", `

// GoRoutine 
class StackFrameBasis
{
public var _Next:Int=0;
public var _recoverNext:Null<Int>=null;
public var _incomplete:Bool=true;
public var _latestPH:Int=0;
public var _latestBlock:Int=0;
public var _functionPH:Int;
public var _functionName:String;
public var _goroutine(default,null):Int;
public var _bds:Dynamic; // bindings for closures
public var _deferStack:List<StackFrame>;
public var _debugVars:Map<String,Dynamic>;
#if godebug
	var _debugVarsLast:Map<String,Dynamic>;
	static var _debugBP:Map<Int,Bool>;
#end

public function new(gr:Int,ph:Int,name:String){
	_goroutine=gr;
	_functionPH=ph;
	_functionName=name;
	#if godebug
		_debugVars=new Map<String,Dynamic>();
		_debugVarsLast= new Map<String,Dynamic>();
	#end
	this.setPH(ph); // so that we call the debugger, if it is enabled
	// TODO optionally profile function entry here
}

public inline function nullOnExitSF(){
	_functionName=null;
	// the next three items could be optimized to only be set to null on exit if they are used in a Go func
	_bds=null;
	_deferStack=null;
	_debugVars=null;
	#if godebug
		_debugVarsLast=null;
	#end
}

public function setDebugVar(name:String,value:Dynamic){
	if(_debugVars==null) 
		_debugVars=new Map<String,Dynamic>();
	_debugVars.set(name,value);	
}

public function setLatest(ph:Int,blk:Int){ // this can be done inline, but generates too much code
	_latestBlock=blk;
	this.setPH(ph);
	// TODO optionally profile block entry here
}

public function breakpoint(){
	#if (godebug && (cpp || neko))
		trace("GODEBUG: runtime.Breakpoint()");
		_debugBP.set(_latestPH,true); // set a constant debug trap
		setPH(_latestPH); // run the debugger
	#else
		//trace("GODEBUG: runtime.Breakpoint() to run debugger (cpp/neko only) use: haxe -D godebug");
	#end
}

public function setPH(ph:Int){
	_latestPH=ph;
	// TODO optionally profile instruction line entry here	
	// optionally add debugger code here, if the target supports Console.readln()
	#if (godebug && (cpp || neko))
		// TODO add support for: cs || java || php 
		if(_debugBP==null||_debugBP.exists(ph)){
			var stay=true;
			var ln:Null<String>;
			while(stay){
				printDebugState();
				ln=Console.readln();
				if(ln==null)
					stay=false; // effectively step a line 
				else {
					// debugger commands
					var fb=new Array<Dynamic>();
					var bits=ln.split(" ");
					switch(ln.charAt(0)){
					case "S","s","R","r":
						if(bits.length<3)
							fb[0]="please use the format: S/R filename linenumber";
						else{
							if(_debugBP==null){
								_debugBP=new Map<Int,Bool>();
							}
							var base=Go.getStartCPos(bits[1]);
							if(base==-1)
								fb[0]="sorry, can't find file: "+bits[1];
							else{
								var off=Std.parseInt(bits[2]);
								if(off==null)
									fb[0]="sorry, can't parseInt: "+bits[2];
								else{
									fb[0]="break-point ";
									switch(ln.charAt(0)){
									case "S","s":
										fb[1]="set";						
										_debugBP.set(base+off,true);
									case "R","r":
										fb[1]="removed";						
										if(_debugBP.exists(base+off))
											_debugBP.remove(base+off);
									}
									fb[2]=" at: "+Go.CPos(base+off);
								}
							}	
						}
					case "B","b":
						if(_debugBP==null){
							fb[0]="no break-points set";
						} else {
							fb[0]="break-points:\n";
							var ent=1;
							for(b in _debugBP.keys()){
								fb[ent]="\t"+Go.CPos(b)+"\n";
								ent+=1;
							}
						}							
					case "C","c":
						_debugBP=null;
						fb[0]="all break-points cleared";
					case "L","l":
						if(bits.length>=2)
							if(_debugVars.exists(bits[1])){
								var v:String;
								if(bits[1].indexOf(".")!=-1) // global
									v="global: "+_debugVars.get(bits[1]).toString();
								else
								 	v=Std.string(_debugVars.get(bits[1]));
								fb[0]="Local assignment to: "+bits[1]+" = "+v.substr(0,500);
							} else
								fb[0]="Can't find local assignment: "+bits[1];
						else{
							fb[0]="Local assignments:\n";
							var ent=1;
							for(b in _debugVars.keys()){
								if(b.indexOf(".")==-1) { // local
									fb[ent]="\t"+b+" = "+Std.string(_debugVars.get(b)).substr(0,500)+"\n";
									ent+=1;
								}	
							}
						}							
					case "G","g":
						if(bits.length<2)
							fb[0]="please use the format: G globalname ";
						else 
							fb[0]="Global: "+Go.getGlobal(bits[1]).substr(0,500);	
					case "M","m":
						if(bits.length<3)
							fb[0]="please use the format: M objectID offset ";
						else {
							var id=Std.parseInt(bits[1]);
							if(id==null)
									fb[0]="sorry, can't parseInt: "+bits[1];
							else{
								var off=Std.parseInt(bits[2]);
								if(off==null)
									fb[0]="sorry, can't parseInt: "+bits[2];
								else
									fb[0]="Memory: "+Object.memory.get(id).toString(off).substr(0,500);	
							}
						}
					case "D","d":
						fb[0]=Scheduler.stackDump();					
					case "P","p":
						Scheduler.panicFromHaxe("panic from debugger");					
						fb[0]="Panicing from debugger to exit program";
						_debugBP=new Map<Int,Bool>();
						_debugBP.set(-1,true); // unreachable break-point
						stay=false;
					case "X","x":
						//_debugBP=new Map<Int,Bool>();
						//_debugBP.set(-1,true); // unreachable break-point
						fb[0]="eXecute program";
						stay=false;
					default:
						fb[0]="commands: blank=step, B=BrakePointList, S/R=Set/RemoveBP name line, C=ClearAllBP, L=Local name, G=Global name, M=Memory id offset, D=stackDump, X=eXecute program, P=Panic (^C does not work)";
					}
					Console.println(fb);
				}
			}
		}
	#end
}

#if godebug
public function printDebugState():Void{
	var guf=new Array<Dynamic>();
	var gc=1;
	guf[0]="GR:"+_goroutine+" - "+_functionName+" @ "+Go.CPos(_latestPH);
	for(k in _debugVars.keys()){
		if(_debugVars.get(k)!=_debugVarsLast.get(k)){
			if(k.indexOf(".")!=-1) // global
				guf[gc]="\n"+k+" = "+cast(_debugVars.get(k),Pointer).toString().substr(0,500);
			else
				guf[gc]="\n"+k+" = "+Std.string(_debugVars.get(k)).substr(0,500);
			gc+=1;
			_debugVarsLast.set(k,_debugVars.get(k));
		}
	}
	Console.println(guf);
}
#end

public function defer(fn:StackFrame){
	if(_deferStack==null)
		_deferStack=new List<StackFrame>();
	_deferStack.add(fn); // add to the end of the list, so that runDefers() get them in the right order
}

public function runDefers(){
	if(_deferStack!=null)
		while(!_deferStack.isEmpty()){
			Scheduler.push(_goroutine,_deferStack.pop());
		}
}

}
`)
	pogo.WriteAsClass("StackFrame", `

interface StackFrame
{
public var _Next:Int;
public var _recoverNext:Null<Int>;
public var _incomplete:Bool;
public var _latestPH:Int;
public var _latestBlock:Int;
public var _functionPH:Int;
public var _functionName:String;
public var _goroutine(default,null):Int;
public var _bds:Dynamic; // bindings for closures as a anonymous struct
public var _deferStack:List<StackFrame>;
public var _debugVars:Map<String,Dynamic>;
function run():StackFrame; // function state machine (set up by each Go function Haxe class)
function res():Dynamic; // function result (set up by each Go function Haxe class)
function nullOnExitSF():Void; // call this when exiting the function
function setDebugVar(name:String,value:Dynamic):Void;
}
`)
	pogo.WriteAsClass("Scheduler", `
@:keep
class Scheduler { // NOTE this code requires a single-thread, as there is no locking TODO detect deadlocks
// public
public static var doneInit:Bool=false; // flag to limit go-routines to 1 during the init() processing phase
// private
static var grStacks:Array<Array<StackFrame>>=new Array<Array<StackFrame>>(); 
static var grInPanic:Array<Bool>=new Array<Bool>();
static var grPanicMsg:Array<Interface>=new Array<Interface>();
static var panicStackDump:String="";
static var entryCount:Int=0; // this to be able to monitor the re-entrys into this routine for debug
static var currentGR:Int=0; // the current goroutine, used by Scheduler.panicFromHaxe(), NOTE this requires a single thread

// if the scheduler is being run from a timer, this is where it comes to
public static var runLimit:Int=0;
public static function timerEventHandler(dummy:Dynamic) {
	if(runLimit<2) 
		runAll();
	else
		runToStasis(runLimit); 
}

static inline function runToStasis(cycles:Int) {
	var lastHash=new Array<Null<Int>>();
	var thisHash=makeStateHash();
	while( !hashesEqual(lastHash,thisHash) && cycles>0){
		lastHash = thisHash;
		runAll();
		thisHash = makeStateHash();
		cycles -= 1;
	}
	//if(cycles>0)
	//	trace("Stasis achieved at "+cycles);
	//else
	//	trace("Stasis not achieved");
}

static function hashesEqual(a:Array<Null<Int>>,b:Array<Null<Int>>):Bool{
	if(a.length != b.length) 
		return false;
	for(i in 0...a.length)
		if(a[i]!=b[i]) 
			return false;
	return true;
}

static function makeStateHash():Array<Null<Int>> { // TODO this is very ugly, and probably slow, so improve by checking for change on the fly?
	var hash=new Array<Null<Int>>();
	for( gr in 0 ... grStacks.length){  // TODO optimise to not use .length
		for(ent in 0 ... grStacks[gr].length){
			hash[gr] += grStacks[gr][ent]._functionPH ;
		}
	}
	//trace("makeStateHash()="+hash);
	return hash;
}

public static function runAll() { // this must be re-entrant, in order to allow Haxe->Go->Haxe->Go for some runtime functions
	var cg:Int=0; // reentrant current goroutine
	entryCount++;
	if(entryCount>2) { // this is the simple limit to runtime recursion  
		throw "Scheduler.runAll() entryCount exceeded - "+stackDump();
	}

	var thisStack:Array<StackFrame>;
	var thisStackLen:Int;

	// special handling for goroutine 0, which is used in the initialisation phase and re-entrantly, where only one goroutine may operate		
	thisStack=grStacks[0];
	thisStackLen=thisStack.length;
	if(thisStackLen==0) { // check if there is ever likley to be anything to do
		if(grStacks.length<=1) { 
			throw "Scheduler: there is only one goroutine and its stack is empty\n"+stackDump();		
		}
	} else { // run goroutine zero
		runOne(0,entryCount,thisStack,thisStackLen);

	}

	if(doneInit && entryCount==1 ) {	 // don't run extra goroutines when we are re-entrant or have not finished initialistion
									     // NOTE this means that Haxe->Go->Haxe->Go code cannot run goroutines 
		var grStacksLen=grStacks.length;
		for(cg in 1...grStacksLen) { // length may grow during a run through, NOTE goroutine 0 not run again
			thisStack=grStacks[cg];
			thisStackLen=thisStack.length;
			if(thisStackLen>0) {
				runOne(cg,entryCount,thisStack,thisStackLen);
			}
		}

		// prune the list of goroutines only at the end (goroutine numbers are in the stack frames, so can't be altered) 
		grStacksLen=grStacks.length;// there may be more goroutines than we started with
		if(grStacksLen>1) // we must always have goroutine 0
			if(grStacks[grStacksLen-1].length==0) 
				grStacks.pop();
	}
	thisStack=null; // for GC
	entryCount--;
}
static inline function runOne(gr:Int,entryCount:Int,thisStack:Array<StackFrame>,thisStackLen:Int){ // called from above to call individual goroutines TODO: Review for multi-threading
	if(grInPanic[gr]) {
		if(entryCount!=1) { // we are in re-entrant code, so we can't panic again, as this may be part of the panic handling...
				// NOTE this means that Haxe->Go->Haxe->Go code cannot use panic() reliably 
				run1a(gr,thisStack,thisStackLen);
		} else {
			while(grInPanic[gr]){
				if(grStacks[gr].length==0){
					 Console.naclWrite("Panic in goroutine "+gr+"\n"+panicStackDump); // use stored stack dump
					 throw "Go panic";
				} else {
					var sf:StackFrame=grStacks[gr].pop();
					if(sf._deferStack!=null)
						while(!sf._deferStack.isEmpty() && grInPanic[gr]) { 
							// NOTE this will run all of the defered code for a function, 
							// NOTE if recover() is encountered it should set grInPanic[gr] to false.
							// TODO consider merging code with RunDefers()
							var def:StackFrame=sf._deferStack.pop();
							//trace("DEBUG runOne panic defer:",def._functionName);
							Scheduler.push(gr,def);
							while(def._incomplete) 
								runAll(); // with entryCount >1, so run as above 
						}
					if(!grInPanic[gr]){
					 	//trace("DEBUG runOne panic - recovered");
						if(sf._recoverNext != null) {
						 	//trace("DEBUG runOne panic - running recovery code");
							sf._Next = sf._recoverNext; // set the re-entry point
						} 
						grStacks[gr].push(sf); // now run the recovery code
					}
					sf=null; // for GC
				}
			}
		}
	} else {
		run1a(gr,thisStack,thisStackLen);
	}
}
public static inline function run1a(gr:Int,thisStack:Array<StackFrame>,thisStackLen:Int){ 
	currentGR=gr;
	thisStack[thisStackLen-1].run();  
}
public static inline function run1(gr:Int){ // used by callFromRT() for every go function
	run1a(gr,grStacks[gr],grStacks[gr].length); // run() may call haxe which calls these routines recursively 
}
public static function makeGoroutine():Int {
	for (r in 1 ... grStacks.length) // goroutine zero is reserved for init activities, main.main() and Haxe call-backs
		if(grStacks[r].length==0)
		{
			grInPanic[r]=false;
			grPanicMsg[r]=null;
			return r;	// reuse a previous goroutine number if possible
		}
	var l:Int=grStacks.length;
	grStacks[l]=new Array<StackFrame>(); 
	grInPanic[l]=false;
	grPanicMsg[l]=null;
	return l;
}
public static inline function pop(gr:Int):StackFrame {
	return grStacks[gr].pop(); // NOTE removing old object pointer does not improve GC
}
public static inline function push(gr:Int,sf:StackFrame){
	grStacks[gr].push(sf);
}
public static inline function NumGoroutine():Int {
	return grStacks.length;
}
public static inline function ThisGoroutine():Int {
	return currentGR;
}

public static function stackDump():String {
	var ret:String = "";
	var gr:Int;
	ret += "runAll() entryCount="+entryCount+"\n";
	for(gr in 0...grStacks.length) {
		ret += "---\nGoroutine " + gr + " "+grPanicMsg[gr]+"\n"; //may need to unpack the interface
		if(grStacks[gr].length==0) {
			ret += "Stack is empty\n";
		} else {
			ret += "Stack has " +grStacks[gr].length+ " entries:\n";
			var e = grStacks[gr].length -1;
			while( e >= 0){
				var ent = grStacks[gr][e];
				if(ent==null) {
					ret += "\tStack entry is null\n";
				} else {
					ret += "\t"+ent._functionName+" starting at "+Go.CPos(ent._functionPH);
					ret += " latest position "+Go.CPos(ent._latestPH);
					ret += " latest block "+ent._latestBlock+"\n";
					if(ent._debugVars!=null){
						for(k in ent._debugVars.keys()) {
							if(k.indexOf(".")==-1){ // not a global assignment, so showing only locals
								var t:Dynamic=ent._debugVars.get(k);
								if(t==null) t="nil";
								if(Std.is(t,Pointer)) t=t.toUniqueVal();
								ret += "\t\tvar "+k+" = "+t+"\n";
								t=null; // for GC
							}
						}
					}
				}
				ent=null; // for GC
				e -= 1;
			}
		}
	}
	return ret;
}

public static function getNumCallers(gr:Int):Int {
	if(grStacks[gr].length==0) {
		return 0;
	} else {
		return grStacks[gr].length;
	}
}

public static function getCallerX(gr:Int,x:Int):Int {
	if(grStacks[gr].length==0) {
		return 0; // error
	} else {
		var e = grStacks[gr].length -1;
		while(e >= 0){
			var ent=grStacks[gr][e];
			if(x==0) {
				if(ent==null) {
					return 0; // this is an error 
				} else {
					return ent._latestPH;
				}
			}
			ent=null; // for GC
			x -= 1;
			e -= 1;
		}
	}
	return 0; // error
}

public static function traceStackDump() {trace(stackDump());}

public static function panic(gr:Int,err:Interface){
	if(gr>=grStacks.length||gr<0)
		throw "Scheduler.panic() invalid goroutine";
	if(grInPanic[gr]) { // if we are already in a panic, not much we can do...
		//trace("Scheduler.panic() panic within panic for goroutine "+Std.string(gr)+" message: "+err.toString());		
	}else{
		grInPanic[gr]=true;
		grPanicMsg[gr]=err;
		panicStackDump=stackDump();
		#if godebug
			trace("GODEBUG: panic in goroutine "+Std.string(gr)+" message: "+err.toString());
			var top = grStacks[gr][grStacks[gr].length-1] //grStacks[gr].first();
			if(top!=null)
				cast(top,StackFrameBasis).breakpoint();
		#end
	} 
}
public static function recover(gr:Int):Interface{
	if(gr>=grStacks.length||gr<0)
		throw "Scheduler.recover() invalid goroutine";
	if(grInPanic[gr]==false)
		return null;
	#if godebug
		trace("GODEBUG: recover in goroutine "+Std.string(gr)+" message: "+grPanicMsg[gr]);
		var top = grStacks[gr][grStacks[gr].length-1] //grStacks[gr].first();
		if(top!=null)
			cast(top,StackFrameBasis).breakpoint();
	#end
	grInPanic[gr]=false;
	var t = grPanicMsg[gr];
	grPanicMsg[gr]=null;
	return t;
}
public static function panicFromHaxe(err:String) { 
	if(currentGR>=grStacks.length||currentGR<0) 
		// if current goroutine is -ve, or out of range, always panics in goroutine 0
		panic(0,new Interface(TypeInfo.getId("string"),"Runtime panic, unknown goroutine, "+err+" "));
	else
		panic(currentGR,new Interface(TypeInfo.getId("string"),"Runtime panic, "+err+" "));
	Console.naclWrite(panicStackDump); 
	throw "Haxe panic"; // NOTE can't be recovered!
}
public static function bbi() {
	panicFromHaxe("bad block ID (internal phi error)");
}
public static function ioor() {
	panicFromHaxe("index out of range");
}
public static function htc(c:Dynamic,pos:Int) {
	panicFromHaxe("Haxe try-catch exception <"+Std.string(c)+"> position "+Std.string(pos)+
		" at or before: "+Go.CPos(pos));
}
public static #if inlinepointers inline #end function wraprangechk(val:Int,sz:Int) {
	if((val<0)||(val>=sz)) ioor();
}
public static function unt():Dynamic {
		panicFromHaxe("nil interface target for method");	
		return null;
}
static function unp() {
		panicFromHaxe("unexpected nil pointer (ssa:wrapnilchk)");	
}
public static function wrapnilchk(p:Pointer):Pointer {
	if(p==null) unp();
	return p;
}
}
`)
	pogo.WriteAsClass("GOmap", `

class GOmap {
	// TODO write a more sophisticated (and hopefully faster) version of this code 
	// TODO in Haxe, the keys can be Int, String or "object" (by reference)
	// TODO there is also a very sophisticated go implementation in runtime

	public var baseMap:Map<String,{key:Dynamic,val:Dynamic}>;
	public var kz:Dynamic;
	public var vz:Dynamic;

	public function new (kDef:Dynamic,vDef:Dynamic) {
		//trace("DEBUG new",kDef,vDef);
		baseMap = new Map<String,{key:Dynamic,val:Dynamic}>();
		kz = kDef;
		vz = vDef;
	}

	public static function makeKey(a:Dynamic):String{
		//trace("DEBUG makeKey",a);
		if(a==null) return "<<<<NULL>>>>"; // TODO how can this be more unique?
		if(Reflect.isObject(a)){
			if(Std.is(a,String))
				return a;
			if(Std.is(a,Pointer))
				return cast(a,Pointer).toUniqueVal();  
			// NOTE Object could be an abstract
			#if !abstractobjects
			if(Std.is(a,Object)){
				//trace("DEBUG makeKey Object found");
				var r=cast(a,Object).toString(); 
				//trace("DEBUG makeKey Object="+r);
				return r;
			}
			#end
			if(Std.is(a,Complex)){
				return Complex.toString(a);
			}
			if(Std.is(a,Interface))
				return cast(a,Interface).toString();
			if(Std.is(a,Slice)) 
				return a.toString();
			if(Std.is(a,Channel))
				return a.toString();
			if(Std.is(a,GOmap)||Std.is(a,Closure)) {
				Scheduler.panicFromHaxe("haxeruntime.GOmap.makeKey() unsupported haxe type: "+a);
				return "";
			}
			#if abstractobjects
				return a.toString(); // must be an Object or Int64
			#else
				return GOint64.toString(a);
			#end
		}
		#if (cpp || cs)	
			// in cpp & cs, Std.string(1.9999999999999998) => "2"
			// TODO consider how to deal with this issue in the compound types above
			if(Std.is(a,Float)) {
				return GOint64.toString(Go_haxegoruntime_FFloat64bits.callFromRT(0,a));
			}
		#end
		return Std.string(a);
	}

	public function set(realKey:Dynamic,value:Dynamic){
		var sKey = makeKey(realKey);
		//trace("DEBUG set",sKey,realKey);
		if(baseMap.exists(sKey)){
			if(!Force.isEqualDynamic(baseMap.get(sKey).key,realKey))
				Scheduler.panicFromHaxe("haxeruntime.GOmap non-unique key for: "+sKey);
		}
		baseMap.set(sKey,{key:realKey,val:value});
	}

	public function get(rKey:Dynamic):Dynamic {
		var sKey = makeKey(rKey);
		//trace("DEBUG get",sKey,rKey);		
		if(baseMap.exists(sKey))	return baseMap.get(sKey).val;
		else 						return vz; // the zero value
	}

	public function exists(rKey:Dynamic):Bool {
		var sKey = makeKey(rKey);
		//trace("DEBUG exists",sKey,rKey);		
		return baseMap.exists(sKey);
	}

	public function remove(r:Dynamic){
		var s = makeKey(r);
		//trace("DEBUG remove",s,r);		
		baseMap.remove(s);
	}

	public function len():Int {
		var _l:Int=0;
		var _it=baseMap.iterator();
		while(_it.hasNext()) {_l++; _it.next();};
		//trace("DEBUG len",_l);		
		return _l;
	}

	public function range():GOmapRange {
		var keys = new Array<String>();
		var k = baseMap.keys(); // in C# and Java, this may not work if new items are added to the map
		while(k.hasNext()) 
			keys.push(k.next());
		return new GOmapRange(keys,this);
	}

}
`)
	pogo.WriteAsClass("GOmapRange", `

class GOmapRange {
	private var k:Array<String>;
	private var m:GOmap;

	public function new(kv:Array<String>, mv:GOmap){
		k=kv;
		m=mv;
	}

	public function next():{r0:Bool,r1:Dynamic,r2:Dynamic} {
		var _hn:Bool=k.length>0;
		if(_hn){
			var _nxt=k.pop();
			if(m.baseMap.exists(_nxt))
				return {r0:true,r1:m.baseMap.get(_nxt).key,r2:m.baseMap.get(_nxt).val};
			else
				return next(); // recurse if this key is not found (deleted in-between?)
		}else{
			return {r0:false,r1:m.kz,r2:m.vz};
		}
	}
}
`)
	pogo.WriteAsClass("GOstringRange", `

class GOstringRange {
	private var g:Int;
	private var k:Int;
	private var v:Slice;

	public function new(gr:Int,s:String){
		g=gr;
		k=0;
		v=Force.toUTF8slice(gr,s);
	}

	public function next():{r0:Bool,r1:Int,r2:Int} {
		var _thisK:Int=k;
		if(k>=v.len())
			return {r0:false,r1:0,r2:0};
		else {
			var _dr:{r0:Int,r1:Int}=Go_unicode_47_utf8_DDecodeRRune.callFromRT(g,v.subSlice(_thisK,-1));
			k+=_dr.r1;
			return {r0:true,r1:_thisK,r2:_dr.r0};
		}
	}
}


`)

	return ""
}
