# run the working unit test using neko --interp
	echo "========================================="
	echo "Unit Test: " $1
	echo "========================================="
	cd $1
	tardisgo -test $1
	if [ "$?" != "0" ]; then
		exit $?
	fi
	haxe -main tardis.Go --interp
	if [ "$?" != "0" ]; then
		exit $?
	fi
	cd ..
	