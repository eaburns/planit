package lifted

type Symtab struct {
	constNums  map[string]int
	constNames []string

	predNums  map[string]int
	predNames []string
	predInertia []inertia

	typeNums  map[string]int
	typeNames []string
	typeObjs  [][]int

	varNames []string
}

type inertia int

const (
	noInertia inertia = iota
	posInertia
	negInertia
)

func NewSymtab() *Symtab {
	return &Symtab{
		constNums: make(map[string]int),
		predNums: make(map[string]int),
		typeNums: make(map[string]int),
	}
}