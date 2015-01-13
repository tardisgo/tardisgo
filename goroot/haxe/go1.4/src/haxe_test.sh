# run the working unit tests
for onelevel in path unicode sort
do
	echo "Unit Test:" $onelevel 
	cd $onelevel
	tardisgo -runall -test $onelevel
	cd .. 
done
for twolevels in  unicode/utf8 container/heap container/list container/ring
do
	echo "Unit Test:" $twolevels 
	cd $twolevels
	tardisgo -runall -test $twolevels
	cd ../.. 
done