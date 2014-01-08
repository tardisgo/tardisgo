
// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tardisgo;

class Interface { // "interface" is a keyword in PHP but solved using compiler flag  --php-prefix go //TODO tidy names 
	public var typ:Int; // the possibly interface type that has been cast to
	public var val:Dynamic;

	public function new(t:Int,v:Dynamic){
		typ=t;
		val=Deep.copy(v); 
	}
	public function toString():String {
		if(val==null)
			return "Interface{null:"+TypeInfo.getName(typ)+"}";
		else
			return "Interface{"+val+":"+TypeInfo.getName(typ)+"}";
	}
	public static function change(t:Int,i:Interface):Interface {
		if(i==null)	
			if(TypeInfo.isConcrete(t)) 
				return new Interface(t,TypeInfo.zeroValue(t)); 
			else {
				Scheduler.panicFromHaxe( "can't change the Interface of a nil value to Interface type: " +TypeInfo.getName(t));  
				return new Interface(t,TypeInfo.zeroValue(t));	 //dummy value as we have hit the panic button
			}
		else 
			if(Std.is(i,Interface)) 	
				if(TypeInfo.isConcrete(t)) 
					return new Interface(t,Deep.copy(i.val)); 
				else
					return new Interface(i.typ,Deep.copy(i.val)); // do not allow non-concrete types for Interfaces
			else {
				Scheduler.panicFromHaxe( "Can't change the Interface of a non-Interface type:"+i+" to: "+TypeInfo.getName(t));  
				return new Interface(t,TypeInfo.zeroValue(t));	 //dummy value as we have hit the panic button
			}
	}
	public static function isEqual(a:Interface,b:Interface):Bool {		// TODO ensure this very wide definition of equality is OK 
		if(a==null) 
			if(b==null) return true;
			else 		return false;
		if(b==null)		
			return false;
		if(! (TypeInfo.isIdentical(a.typ,b.typ)||TypeInfo.isAssignableTo(a.typ,b.typ)||	TypeInfo.isAssignableTo(b.typ,a.typ)) ) 
			return false;
		else
			if(a.val==b.val) 
				return true; // simple equality
			else // could still be equal underneath a pointer    //TODO is another special case required for Slice?
				if(Std.is(a.val,Pointer) && Std.is(b.val,Pointer))
					return a.val.load() == b.val.load();
				else
					return false;	
	}			
	/* from the SSA documentation:
	If AssertedType is a concrete type, TypeAssert checks whether the dynamic type in Interface X is equal to it, and if so, 
		the result of the conversion is a copy of the value in the Interface.
	If AssertedType is an Interface, TypeAssert checks whether the dynamic type of the Interface is assignable to it, 
		and if so, the result of the conversion is a copy of the Interface value X. If AssertedType is a superInterface of X.Type(), 
		the operation will fail iff the operand is nil. (Contrast with ChangeInterface, which performs no nil-check.)
	*/
	public static function assert(assTyp:Int,ifce:Interface):Dynamic{
		if(ifce==null) 
			Scheduler.panicFromHaxe( "Interface.assert null Interface");
		if(!(TypeInfo.isAssignableTo(ifce.typ,assTyp)||TypeInfo.isIdentical(assTyp,ifce.typ))) // TODO review need for isIdentical 
			Scheduler.panicFromHaxe( "type assert failed: expected "+TypeInfo.getName(assTyp)+", got "+TypeInfo.getName(ifce.typ) );
		if(TypeInfo.isConcrete(assTyp))	
			return Deep.copy(ifce.val);
		else	
			return new Interface(ifce.typ,ifce.val);
	}
	public static function assertOk(assTyp:Int,ifce:Interface):{r0:Dynamic,r1:Bool} {
		if(ifce==null) 
			return {r0:TypeInfo.zeroValue(assTyp),r1:false};
		if(!(TypeInfo.isAssignableTo(ifce.typ,assTyp)||TypeInfo.isIdentical(assTyp,ifce.typ))) // TODO review need for isIdentical 
			return {r0:TypeInfo.zeroValue(assTyp),r1:false};
		if(TypeInfo.isConcrete(assTyp))	
			return {r0:Deep.copy(ifce.val),r1:true};
		else	
			return {r0:new Interface(ifce.typ,ifce.val),r1:true};
	}
	public static function invoke(ifce:Interface,meth:String,args:Array<Dynamic>):Dynamic {
		if(ifce==null) 
			Scheduler.panicFromHaxe( "Interface.invoke null Interface"); 
		//trace("Invoke:"+ifce+":"+meth);
		if(!Std.is(ifce,Interface)) 
			Scheduler.panicFromHaxe( "Interface.invoke on non-Interface value"); 
		//return Reflect.callMethod(o:Dynamic, func:Dynamic, args:Array<Dynamic>);
		var fn:Dynamic=TypeInfo.method(ifce.typ,meth);
		//trace("Invoke:"+TypeInfo.getName(ifce.typ)+":"+meth+":"+ifce.val+":"+fn);
		//return fn([],Deep.copy(ifce.val));
		return Reflect.callMethod(null, fn, args);
	}
}
