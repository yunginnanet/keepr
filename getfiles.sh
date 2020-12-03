echo -e "\e[1;33m_,  _, _, _ __,  _, _ _, _ __,  "
echo -e "\e[1;33m(_  /_\ |\/| |_) (_  | |\/| |_) "
echo -e "\e[1;33m, ) | | |  | |   , ) | |  | |   "
echo -e "\e[1;33m ~  ~ ~ ~  ~ ~    ~  ~ ~  ~ ~   "
echo -e "\e[0m"
KSEARCH="/media/kayos/1tbssd/Samples-v2/"
KOUTD="/media/kayos/1tbssd/Samples-v2/001-LINKED_SORTED_DIRECTORIES/Key"
KSTART=1

count() {
        echo -e -n "\e[4mCurrent file count: "
        KCNT=$(ls ${KOUTD}/$INKEY | wc -l)
        if [ ${KCNT} == 0 ]; then 
        	echo -e "\e[2mNone\e[0m";
        	sleep 1
        else
        	echo -e -n "\e[1m "; echo $KCNT; echo -e "\e[0m";
	        if [ $KSTART == 1 ]; then
        		echo "Directory not empty, continue?"
        		read -r -p "Are you sure? [y/N] " response
        		echo ""
        		case "$response" in
	       		[yY][eE][sS]|[yY])
	       			export KSTART=0
        			echo "Starting...";
        			;;
        		*)
        			exit
        			;;
        		esac
        	fi
	fi
}



if [ -z "$1" ]; then
	echo "usage ${0} Key";
else
	INKEY=$1
	echo "Finding files in: ${KSEARCH}"
	echo "To link into:     ${KOUTD}/${INKEY}/"
	echo ""
	echo "In key:           ${INKEY}"
	echo ""
	count
	sleep 1
	echo "Finding files in key ${INKEY} and creating symlinks..."
	sleep 1
	fdfind --type f --glob "* ${INKEY} *.wav" \
	--exec ln -i {} ${KOUTD}/${INKEY}/{/} \; \
	fdfind --type f --glob "*_${INKEY}_*.wav" \
	--exec ln -i {} ${KOUTD}/${INKEY}/{/} \; \
	/media/kayos/1tbssd/Samples-v2/ 2>/dev/null;
	echo ""
	echo -e "\e[32mDone!\e[0m"
	count
fi
echo ""
