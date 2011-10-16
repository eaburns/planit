package prob

type Symtab struct {
	constNums  map[string]int
	constNames []string

	predNums    map[string]int
	predNames   []string
	predInertia []byte // Inertia bit-flags

	typeNums  map[string]int
	typeNames []string
	typeObjs  [][]int // Objects of the given type

	varNames []string
}

const (
	// Inertia bit flags
	posInertia = 1 << 0
	negInertia = 1 << 1
)

func NewSymtab() *Symtab {
	return &Symtab{
		constNums: make(map[string]int),
		predNums:  make(map[string]int),
		typeNums:  make(map[string]int),
	}
}
