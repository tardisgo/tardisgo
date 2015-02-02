# run the working unit tests using neko --interp
for onelevel in bytes errors path sort strings unicode 
do
	echo "Unit Test:" $onelevel 
	cd $onelevel
	tardisgo -test $onelevel
	haxe -main tardis.Go --interp
	cd .. 
done
for twolevels in container/heap container/list container/ring encoding/ascii85 encoding/base32 encoding/base64 encoding/hex text/tabwriter unicode/utf8 unicode/utf16 
do
	echo "Unit Test:" $twolevels 
	cd $twolevels
	tardisgo  -test $twolevels
	haxe -main tardis.Go --interp
	cd ../.. 
done