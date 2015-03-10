# run the working unit tests using the fastest/most accurate method: C++, JS or neko/interp
for onelevel in errors path
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
for twolevels in container/heap container/list crypto/aes crypto/cipher encoding/ascii85 encoding/base32 image/color text/tabwriter unicode/utf8 unicode/utf16 
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
for onelevel in bufio bytes flag fmt html io math regexp sort strings unicode 
do
	echo "========================================="
	echo "Unit Test (via js): " $onelevel 
	echo "========================================="
	cd $onelevel
	tardisgo -haxe js -test $onelevel
	if [ "$?" != "0" ]; then
		cd .. 
		exit $?
	fi
	cd .. 
done
for twolevels in archive/tar compress/bzip2 compress/gzip compress/lzw container/ring crypto/des crypto/md5 crypto/rc4 crypto/sha1 crypto/sha256 crypto/sha512 encoding/base64 encoding/csv encoding/hex go/format go/scanner hash/adler32 hash/crc32 hash/crc64 hash/fnv image/draw index/suffixarray io/ioutil math/cmplx path/filepath text/scanner
do
	echo "========================================="
	echo "Unit Test (via js): " $twolevels 
	echo "========================================="
	cd $twolevels
	tardisgo -haxe js -test $twolevels
	if [ "$?" != "0" ]; then
		cd ../.. 
		exit $?
	fi
	cd ../.. 
done
for threelevels in database/sql/driver
do
	echo "========================================="
	echo "Unit Test (via js): " $threelevels 
	echo "========================================="
	cd $threelevels
	tardisgo -haxe js -test $threelevels
	if [ "$?" != "0" ]; then
		cd ../../.. 
		exit $?
	fi
	cd ../../.. 
done
for onelevel in runtime strconv
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
for twolevels in archive/zip compress/flate compress/zlib regexp/syntax sync/atomic
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
echo "====================="
echo "All Unit Tests Passed" 
echo "====================="
