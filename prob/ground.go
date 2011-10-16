package prob

import "fmt"

func Ground(d *Domain, p *Problem) (ops []Operator) {
	syms := makeSymtab()
	d.assignNums(&syms)
	p.assignNums(&syms)

	d.findInertia(&syms)

	acts := d.expandQuants(&syms)
	for i := range acts {
		ops = append(ops, acts[i].operators()...)
	}

	fmt.Printf("%d actions\n", len(d.Actions))
	fmt.Printf("%d types\n", len(syms.typeNames))
	fmt.Printf("%d predicates\n", len(syms.predNames))
	fmt.Printf("%d constants\n", len(syms.constNames))
	fmt.Printf("%d literals\n", len(syms.lits.byNum))
	fmt.Printf("%d grounded operators\n", len(ops))

	return
}
