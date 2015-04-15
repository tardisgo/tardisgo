echo "========================================="
echo "Unit Test ( all ) : " $* 
echo "========================================="
cd $1
tardisgo -haxe all -test $*
exit $?