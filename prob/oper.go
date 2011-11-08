package prob

// Ground operators

import "fmt"

type Operator struct {
	Name     string
	Parms    []Name
	Cost     float32
	Precond  []Literal
	Effect   []Literal
	CondEffs []CondEffect
}

type CondEffect struct {
	Cond   []Literal
	Effect []Literal
}

func (a *Action) operators() (ops []Operator) {
	a.dnf()
	a.ensureDnf()

	parms := make([]Name, len(a.Parameters))
	for i := range a.Parameters {
		parms[i] = a.Parameters[i].Name
	}

	conds := gatherOrs(a.Precondition)
	ueffs, ceffs := gatherEffects(a.Effect)

	for i := range conds {
		ops = append(ops, Operator{
			Name:     a.Name,
			Parms:    parms,
			Cost:     float32(1), // For now
			Precond:  gatherLits(conds[i]),
			Effect:   ueffs,
			CondEffs: ceffs,
		})
	}
	return
}

// Get the unconditional and conditional effects
func gatherEffects(f Formula) (ueffs []Literal, ceffs []CondEffect) {
	for _, eff := range gatherOrs(f) {
		switch n := eff.(type) {
		case *WhenNode:
			ceffs = append(ceffs, CondEffect{
				Cond:   gatherLits(n.Condition),
				Effect: gatherLits(n.Formula),
			})
		case *AndNode:
			ueffs = append(ueffs, gatherLits(n)...)
		case *Literal:
			if n.Num < 0 {
				panic("A Literal that was no assigned a number")
			}
			ueffs = append(ueffs, *n)
		case *AssignNode:
			// Ignore assignment for now
		case TrueNode:
			// Ignore
		case FalseNode:
			// Ignore
		default:
			panic(fmt.Sprintf("gatherEffects: unexpected node type: %T", f))
		}
	}
	return
}

func gatherLits(f Formula) (lits []Literal) {
	switch n := f.(type) {
	case *AndNode:
		lits = append(lits, gatherLits(n.Left)...)
		lits = append(lits, gatherLits(n.Right)...)
	case *Literal:
		if n.Num < 0 {
			panic("A Literal that was no assigned a number")
		}
		lits = append(lits, *n)
	case *AssignNode:
		// Ignore an assignment
	default:
		panic(fmt.Sprintf("gatherLits: unexpected node type: %T", f))
	}
	return
}
