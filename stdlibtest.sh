# run the working unit test using neko --interp
	echo "========================================="
	echo "Unit Test: " $1
	echo "========================================="
	cd goroot/haxe/go1.4/src/$1
	tardisgo -test $1
	if [ "$?" != "0" ]; then
		exit $?
	fi
	haxe -main tardis.Go --interp
	if [ "$?" != "0" ]; then
		exit $?
	fi
