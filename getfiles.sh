echo -e "\e[1;33m_,  _, _, _ __,  _, _ _, _ __,  "
echo -e "\e[1;33m(_  /_\ |\/| |_) (_  | |\/| |_) "
echo -e "\e[1;33m, ) | | |  | |   , ) | |  | |   "
echo -e "\e[1;33m ~  ~ ~ ~  ~ ~    ~  ~ ~  ~ ~   "
echo -e "\e[0m"
KSEARCH="/media/kayos/1tbssd/Samples-v2/"
OUTD="/media/kayos/1tbssd/Samples-v2/001-LINKED_SORTED_DIRECTORIES"
KSTART=1
DRUMSTR="Drums"
count() {
        echo -e -n "\e[4mFile count: "
	export KSTART=0
        KCNT=$(find ${OUTD}/$1 -type l | wc -l)
        if [ ${KCNT} == 0 ]; then 
        	echo -e "\e[2mNone\e[0m";
        	sleep 1
        else
        	echo -e -n "\e[1m "; echo $KCNT; echo -e "\e[0m";
	        if [ ${KSTART} == 1 ]; then
        		echo "Directory not empty, continue?"
        		read -r -p "Are you sure? [y/N] " response
        		case "$response" in
	       		[yY][eE][sS]|[yY])
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
	echo "usage ${0} key [key1] [key2] [key3] [key4] OR drums";
else
	INKEY=$1
	if [ "$INKEY" == "$DRUMSTR" ]; then
		echo "Finding files in:"
		echo -e "\e[1;33m${KSEARCH}\e[0m"
		echo "to link into appropridate subdirectories of:"
		echo -e "\e[1;33m${OUTD}/Drums/\e[0m"
		echo ""
		count "Drums/"
		
		echo "Finding kicks and creating symlinks..."
		fdfind -i --type f --glob "*kick*" \
			--exec ln -s {} ${OUTD}/Drums/Kicks/{/} \; \
			${KSEARCH}  2>error.log;
			
		echo "Finding snares and creating symlinks..."		
		fdfind -i --type f --glob "*snare*" \
			--exec ln -s {} ${OUTD}/Drums/Snares/{/} \; \
			${KSEARCH}  2>error.log;

		echo "Finding claps and creating symlinks..."
		fdfind -i --type f --glob "*clap*" \
			--exec ln -s {} ${OUTD}/Drums/Claps/{/} \; \
			${KSEARCH}  2>error.log;

		echo "Finding hi-hats and creating symlinks..."
		fdfind -i --type f --glob "*hat*open*" \
			--exec ln -s {} ${OUTD}/Drums/HiHats-Open/{/} \; \
			${KSEARCH}  2>error.log;
		fdfind -i --type f --glob "*open*hat*" \
			--exec ln -s {} ${OUTD}/Drums/HiHats-Open/{/} \; \
			${KSEARCH}  2>error.log;
		fdfind -i --type f --glob "*hat*closed*" \
			--exec ln -s {} ${OUTD}/Drums/HiHats-Closed/{/} \; \
			${KSEARCH}  2>error.log;
		fdfind -i --type f --glob "*closed*hat*" \
			--exec ln -s {} ${OUTD}/Drums/HiHats-Closed/{/} \; \
			${KSEARCH}  2>error.log;
			
		echo "Finding 808s and creating symlinks..."
		fdfind -i --type f --glob "*808*" \
			--exec ln -s {} ${OUTD}/Drums/808s/{/} \; \
			${KSEARCH}  2>error.log;

		echo ""
		echo -e "\e[32mDone!\e[0m"
					
		count Drums
	else
		echo "Finding files in:"
		echo -e "\e[1;33m${KSEARCH}\e[0m"
		echo "to link into appropridate subdirectories of:"
		echo -e "\e[1;33m${OUTD}/Key/\e[0m"
		count "Key/"
		sleep 1
		for arg; do
			echo -n "Finding files in key $arg and creating symlinks..."
			fdfind --type f --glob "*_${arg}_*.wav" \
				--exec ln -s {} ${OUTD}/Key/$arg/{/} \; \
				${KSEARCH} 2>error.log;
			echo -n -e "\e[32mDone! \e[0m"
			count "/Key/$arg" 
		done
	fi
fi

echo ""
