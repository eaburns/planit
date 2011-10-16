package prob

// A hash table for interning literals

import "rand"

const literalTabSz = 104729	// prime, should be large enough

type literalTable struct{
	hashVec [][]uint64
	buckets [literalTabSz][]*Literal
	byNum []*Literal
}

func (tab *literalTable) intern(l *Literal) *Literal {
	hash := tab.hash(l)
	ind := hash % literalTabSz

	for i := range tab.buckets[ind] {
		if tab.buckets[ind][i].eq(l) {
			return tab.buckets[ind][i]
		}
	}

	l.Num = len(tab.byNum)
	tab.byNum = append(tab.byNum, l)
	tab.buckets[ind] = append(tab.buckets[ind], l)
	return l
}

func (a *Literal) eq(b *Literal) bool {
	if a.Predicate.Num != b.Predicate.Num || a.Positive != b.Positive {
		return false
	}
	for i := range a.Parameters {
		switch a := a.Parameters[i].(type) {
		case *Constant:
			if b, ok := b.Parameters[i].(*Constant); !ok || a.Num != b.Num {
				return false
			}
		case *Variable:
			if b, ok := b.Parameters[i].(*Variable); !ok || a.Num != b.Num {
				return false
			}
		default:
			panic("unexpected term type")
		}
	}
	return true
}

func (tab *literalTable) hash(l *Literal) uint64 {
	hash := tab.hvec(0, l.Predicate.Num)

	if l.Positive {
		hash ^= tab.hvec(1, 1)
	} else {
		hash ^= tab.hvec(1, 0)
	}

	for i := range l.Parameters {
		hash ^= tab.hvec(i + 2, l.Parameters[i].Number())
	}

	return hash
}

func (tab *literalTable) hvec(a, b int) uint64 {
	if a >= len(tab.hashVec) {
		toadd := a - len(tab.hashVec) + 1
		tab.hashVec = append(tab.hashVec, make([][]uint64, toadd)...)
	}
	if b >= len(tab.hashVec[a]) {
		toadd := b - len(tab.hashVec[a]) + 1
		for i := 0; i < toadd; i++ {
			tab.hashVec[a] = append(tab.hashVec[a], uint64(rand.Int63()))
		}
	}
	return tab.hashVec[a][b]
}