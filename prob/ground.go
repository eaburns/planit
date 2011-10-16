package prob

func Ground(d *Domain, p *Problem) (ops []Oper) {
	syms := newSymtab()
	d.assignNums(syms)
	p.assignNums(syms)

	d.findInertia(syms)

	acts := d.expandQuants(syms)
	for i := range acts {
		ops = append(ops, acts[i].operators()...)
	}

	return
}
