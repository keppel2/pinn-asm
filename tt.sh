./pp.sh pparse && for i in {1..7}; do ./mt.sh em$i || exit 1 ; done
