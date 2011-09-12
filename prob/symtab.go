package prob

import "bytes"

type Symtab struct {
	preds  []*Pred
	predNums   map[string]int
	constNames []string
	constNums  map[string]int
	varNames   []string
}

type Pred struct {
	Name string
	Parms []string
}

func (s *Symtab) ConstNum(name string) int {
	if n, ok := s.constNums[name]; ok {
		return n
	}
	n := len(s.constNames) + 1
	s.constNums[name] = n
	s.constNames = append(s.constNames, name)
	return n
}

func (s *Symtab) ConstName(n int) (string, bool) {
	if n <= 0 {
		panic("Non-positive constant number")
	}
	n--
	if n > len(s.constNames) {
		return "", false
	}
	return s.constNames[n], true
}

func (s *Symtab) VarNum(f *Frame, name string) (int, *Frame) {
	if n, ok := f.lookup(name); ok {
		return n, f
	}
	n := len(s.varNames) + 1
	s.varNames = append(s.varNames, name)
	return -n, f.push(name, -n)
}

func (s *Symtab) VarName(n int) (string, bool) {
	if n >= 0 {
		panic("Non-negative variable number")
	}
	n = -(n + 1)
	if n > len(s.varNames) {
		return "", false
	}
	return s.varNames[n], true
}

func (s *Symtab) PredNum(name string, parms []string) int {
	str := predString(name, parms)
	if n, ok := s.predNums[str]; ok {
		return n
	}
	n := len(s.preds) + 1
	s.predNums[str] = n
	s.preds = append(s.preds, &Pred{Name: name, Parms: parms})
	return n
}

func predString(name string, parms []string) string {
	buf := bytes.NewBufferString(name)
	for _, p := range parms {
		buf.WriteByte(' ')
		buf.WriteString(p)
	}
	return buf.String()
}

type Frame struct {
	name   string
	number int
	up     *Frame
}

func (f *Frame) push(vr string, num int) *Frame {
	return &Frame{name: vr, number: num, up: f}
}

func (f *Frame) lookup(vr string) (int, bool) {
	if f == nil {
		return 0, false
	}
	if f.name == vr {
		return f.number, true
	}
	return f.up.lookup(vr)
}
