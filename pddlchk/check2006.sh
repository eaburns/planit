#!/bin/sh

dirs=/home/aifs2/group/data/pddl/ipc2006/

for dir in $dirs/*; do
	for prob in $dir/Propositional/p*.pddl ; do
		pnum=$(echo $(basename $prob) | sed 's/.pddl//g')

		dom=$dir/Propositional/domain_${pnum}.pddl
		if ! test -e $dom; then
			dom=$dir/Propositional/domain.pddl
		fi

		echo -n "$(basename $dir) $(basename $dom) $(basename $prob)… "
		args=
		if test "$(basename $dir)" = "pathways"; then
			args="-ignore-multiple-objects"
		fi
		time -f "%E %M kB" ./pddlchk $args $dom $prob || {
			echo ./pddlchk $args $dom $prob
			exit 1
		}
	done
done