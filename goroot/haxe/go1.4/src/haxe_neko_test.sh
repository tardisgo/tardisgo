# run the working unit tests using neko --interp
for onelevel in bufio bytes errors path sort strings unicode 
do
	echo "========================================="
	echo "Unit Test: " $onelevel 
	echo "========================================="
	cd $onelevel
	tardisgo -test $onelevel
	if [ "$?" != "0" ]; then
		cd .. 
		exit $?
	fi
	haxe -main tardis.Go --interp
	if [ "$?" != "0" ]; then
		cd .. 
		exit $?
	fi
	cd .. 
done
for twolevels in container/heap container/list container/ring encoding/ascii85 encoding/base32 encoding/base64 encoding/hex text/tabwriter unicode/utf8 unicode/utf16 
do
	echo "========================================="
	echo "Unit Test: " $twolevels 
	echo "========================================="
	cd $twolevels
	tardisgo  -test $twolevels
	if [ "$?" != "0" ]; then
		cd ../.. 
		exit $?
	fi
	haxe -main tardis.Go --interp
	if [ "$?" != "0" ]; then
		cd ../.. 
		exit $?
	fi
	cd ../.. 
done
echo "====================="
echo "All Unit Tests Passed" 
echo "====================="
