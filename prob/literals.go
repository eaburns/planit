package prob

// A hash table for interning literals

import "math/rand"

const literalTabSz = 104729 // prime, should be large enough

type literalTable struct {
	hashVec [][]uint64
	interns [literalTabSz][]*Literal
	byNum   []*Literal
}

func (tab *literalTable) intern(l *Literal) *Literal {
	hash := tab.hash(l)
	ind := hash % literalTabSz

	for i := range tab.interns[ind] {
		if tab.interns[ind][i].eq(l) {
			return tab.interns[ind][i]
		}
	}

	l.Num = len(tab.byNum)
	tab.byNum = append(tab.byNum, l)
	tab.interns[ind] = append(tab.interns[ind], l)
	return l
}

func (a *Literal) eq(b *Literal) bool {
	if a.Predicate.Num != b.Predicate.Num || a.Positive != b.Positive {
		return false
	}
	for i := range a.Parameters {
		if a.Parameters[i].Variable != b.Parameters[i].Variable ||
			a.Parameters[i].Num != b.Parameters[i].Num {
			return false
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
		hash ^= tab.hvec(i+2, l.Parameters[i].Number())
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
