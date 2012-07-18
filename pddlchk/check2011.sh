#!/bin/sh

dirs=/home/aifs2/group/data/pddl/ipc2011/seq-sat

for dir in $dirs/*; do
	for prob in $dir/problems/*.pddl ; do
		pnum=$(echo $(basename $prob) | sed 's/.pddl//g')

		dom=$dir/domain/${pnum}-domain.pddl
		if ! test -e $dom; then
			dom=$(ls $dir/domain/*.pddl | head -n 1)
		fi

		echo -n "$(basename $dir) $(basename $dom) $(basename $prob)… "
		time -f "%E %M kB" ./pddlchk $dom $prob || {
			echo ./pddlchk $dom $prob
			exit 1
		}
	done
done