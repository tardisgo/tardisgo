# run the working unit tests
for onelevel in   sort
do
	echo "Unit Test:" $onelevel 
	cd $onelevel
	tardisgo -testall -test $onelevel
	cd .. 
done
for twolevels in  unicode/utf8 container/heap container/list container/ring
do
	echo "Unit Test:" $twolevels 
	cd $twolevels
	tardisgo -testall -test $twolevels
	cd ../.. 
done