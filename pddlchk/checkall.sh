#!/bin/sh

dirs=/home/aifs2/group/data/pddl/ipc2008/seq-sat

for dir in $dirs/*; do
	for dom in $dir/*-domain.pddl; do
		prob=$(echo $dom | sed 's/-domain//')
		echo -n "$(basename $dir) $(basename $dom) $(basename $prob)… "
		time -f "%E %M kB" ./pddlchk $dom $prob || {
			echo ./pddlchk $dom $prob
			exit 1
		}
	done
done