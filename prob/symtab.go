package prob

type symtab struct {
	constNums  map[string]int
	constNames []string

	predNums    map[string]int
	predNames   []string
	inertia []byte

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

func newSymtab() *symtab {
	return &symtab{
		constNums: make(map[string]int),
		predNums:  make(map[string]int),
		typeNums:  make(map[string]int),
	}
}
