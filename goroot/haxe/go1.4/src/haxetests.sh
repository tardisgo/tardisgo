# run the working unit tests using the fastest/most accurate method C++ or neko/interp
for onelevel in bufio bytes sort
do
	echo "========================================="
	echo "Unit Test (via C++): " $onelevel 
	echo "========================================="
	cd $onelevel
	tardisgo -haxe cpp -test $onelevel
	if [ "$?" != "0" ]; then
		cd .. 
		exit $?
	fi
	cd .. 
done
for twolevels in container/ring  # this item is a place-holder, it does not run long enough to require C++
do
	echo "========================================="
	echo "Unit Test (via C++): " $twolevels 
	echo "========================================="
	cd $twolevels
	tardisgo -haxe cpp -test $twolevels
	if [ "$?" != "0" ]; then
		cd ../.. 
		exit $?
	fi
	cd ../.. 
done
for onelevel in errors path strings unicode 
do
	echo "========================================="
	echo "Unit Test (via interpreter): " $onelevel 
	echo "========================================="
	cd $onelevel
	tardisgo -haxe interp -test $onelevel
	if [ "$?" != "0" ]; then
		cd .. 
		exit $?
	fi
	cd .. 
done
for twolevels in container/list container/heap encoding/ascii85 encoding/base32 encoding/base64 encoding/hex text/tabwriter unicode/utf8 unicode/utf16 
do
	echo "========================================="
	echo "Unit Test (via interpreter): " $twolevels 
	echo "========================================="
	cd $twolevels
	tardisgo -haxe interp -test $twolevels
	if [ "$?" != "0" ]; then
		cd ../.. 
		exit $?
	fi
	cd ../.. 
done
echo "====================="
echo "All Unit Tests Passed" 
echo "====================="
