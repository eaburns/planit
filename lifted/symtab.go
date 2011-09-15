package lifted

type Symtab struct {
	consts   CoatCheck
	preds    CoatCheck
	types    CoatCheck
	varNames []string
}

func NewSymtab() *Symtab {
	return &Symtab{
		consts: MakeCoatCheck(),
		preds:  MakeCoatCheck(),
		types:  MakeCoatCheck(),
	}
}

// Assign unique small numbers to strings
type CoatCheck struct {
	nums    map[string]int
	strings []string
}

func MakeCoatCheck() CoatCheck {
	return CoatCheck{nums: make(map[string]int)}
}

func (n *CoatCheck) Number(name *Name) {
	if num, ok := n.nums[name.String]; ok {
		name.Number = num
	}
	num := len(n.strings)
	n.nums[name.String] = num
	n.strings = append(n.strings, name.String)
	name.Number = num
}
