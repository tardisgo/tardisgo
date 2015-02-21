# run the working unit tests using the fastest/most accurate method: C++ or neko/interp
for onelevel in bufio bytes fmt math runtime sort strconv
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
for twolevels in hash/adler32 hash/crc32 hash/crc64 hash/fnv math/cmplx sync/atomic
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
for onelevel in errors flag path strings unicode 
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
for twolevels in container/heap container/list container/ring encoding/ascii85 encoding/base32 encoding/base64 encoding/csv encoding/hex text/scanner text/tabwriter unicode/utf8 unicode/utf16 
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
