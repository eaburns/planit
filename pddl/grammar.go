package pddl

import (
	"io"
)

// ParseDomain returns a *Domain parsed from
// a PDDL file, or an error if the parse fails.
func ParseDomain(file string, r io.Reader) (d *Domain, err error) {
	p, err := newParser(file, r)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(parseError)
			if !ok {
				panic(r)
			}
		}
	}()
	p.expect(tokOpen)
	p.expectId("define")
	d = new(Domain)
	d.Name = parseDomainName(p)
	d.Requirements = parseReqsDef(p)
	d.Types = parseTypesDef(p)
	d.Constants = parseConstsDef(p)
	d.Predicates = parsePredsDef(p)

	// Ignore :functions for now
	if p.acceptNamedList(":functions") {
		for nesting := 1; nesting > 0; {
			switch p.next().typ {
			case tokClose:
				nesting--
			case tokOpen:
				nesting++
			}
		}
	}
	for p.peek().typ == tokOpen {
		d.Actions = append(d.Actions, parseActionDef(p))
	}

	p.expect(tokClose)
	return d, nil
}

func parseDomainName(p *parser) string {
	p.expect(tokOpen)
	p.expectId("domain")
	n := p.expect(tokId)
	p.expect(tokClose)
	return n.text
}

func parseReqsDef(p *parser) (reqs []string) {
	if p.acceptNamedList(":requirements") {
		for t, ok := p.accept(tokCid); ok; t, ok = p.accept(tokCid) {
			reqs = append(reqs, t.text)
		}
		p.expect(tokClose)
	}
	return
}

func parseTypesDef(p *parser) (types []TypedName) {
	if p.acceptNamedList(":types") {
		types = parseTypedListString(p, tokId)
		p.expect(tokClose)
	}
	return
}

func parseConstsDef(p *parser) (consts []TypedName) {
	if p.acceptNamedList(":constants") {
		consts = parseTypedListString(p, tokId)
		p.expect(tokClose)
	}
	return
}

func parsePredsDef(p *parser) (Predicates []Predicate) {
	if p.acceptNamedList(":predicates") {
		Predicates = append(Predicates, parseAtomicFormSkele(p))
		for p.peek().typ == tokOpen {
			Predicates = append(Predicates, parseAtomicFormSkele(p))
		}
		p.expect(tokClose)
	}
	return
}

func parseAtomicFormSkele(p *parser) Predicate {
	p.expect(tokOpen)
	pred := Predicate{
		Name:       parseName(p, p.expect(tokId).text),
		Parameters: parseTypedListString(p, tokQid),
	}
	p.expect(tokClose)
	return pred
}

func parseActionDef(p *parser) Action {
	p.expect(tokOpen)
	p.expectId(":action")

	act := Action{Name: p.expect(tokId).text, Parameters: parseActParms(p)}

	if p.peek().text == ":precondition" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.Precondition = parsePreGd(p)
		}
	}
	if p.peek().text == ":effect" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.Effect = parseEffect(p)
		}
	}

	p.expect(tokClose)
	return act
}

func parseActParms(p *parser) []TypedName {
	p.expectId(":parameters")
	p.expect(tokOpen)
	res := parseTypedListString(p, tokQid)
	p.expect(tokClose)
	return res
}

func parsePreGd(p *parser) (res Formula) {
	switch {
	case p.acceptNamedList("and"):
		res = parseAndGd(p, parsePreGd)
	case p.acceptNamedList("forall"):
		res = parseForallGd(p, parsePreGd)
	default:
		res = parsePrefGd(p)
	}
	return
}

func parsePrefGd(p *parser) Formula {
	return parseGd(p)
}

func parseGd(p *parser) (res Formula) {
	switch {
	case p.acceptNamedList("and"):
		res = parseAndGd(p, parseGd)
	case p.acceptNamedList("or"):
		res = parseOrGd(p, parseGd)
	case p.acceptNamedList("not"):
		res = &NotNode{UnaryNode{parseGd(p)}}
		p.expect(tokClose)
	case p.acceptNamedList("imply"):
		res = &ImplyNode{BinaryNode{parseGd(p), parseGd(p)}}
		p.expect(tokClose)
	case p.acceptNamedList("exists"):
		res = parseExistsGd(p, parseGd)
	case p.acceptNamedList("forall"):
		res = parseForallGd(p, parseGd)
	case p.acceptNamedList("not"):
		res = &NotNode{UnaryNode{parseProposition(p)}}
		p.expect(tokClose)
	default:
		res = parseProposition(p)
	}
	return
}

func parseProposition(p *parser) *PropositionNode {
	p.expect(tokOpen)
	defer p.expect(tokClose)
	return &PropositionNode{
		Predicate:  parseName(p, p.expect(tokId).text),
		Parameters: parseTerms(p),
	}
}

func parseTerms(p *parser) (lst []Term) {
	for {
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, Term{Name: parseName(p, t.text)})
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, Term{Name: parseName(p, t.text), Variable: true})
			continue
		}
		break
	}
	return
}

func parseAndGd(p *parser, nested func(*parser) Formula) Formula {
	var fs []Formula
	for p.peek().typ == tokOpen {
		fs = append(fs, nested(p))
	}
	defer p.expect(tokClose)
	return &AndNode{MultiNode{fs}}
}

func parseOrGd(p *parser, nested func(*parser) Formula) Formula {
	var fs []Formula
	for p.peek().typ == tokOpen {
		fs = append(fs, nested(p))
	}
	defer p.expect(tokClose)
	return &OrNode{MultiNode{fs}}
}

func parseForallGd(p *parser, nested func(*parser) Formula) Formula {
	p.expect(tokOpen)
	vars := parseTypedListString(p, tokQid)
	p.expect(tokClose)

	defer p.expect(tokClose)
	return &ForallNode{QuantNode{vars, UnaryNode{Formula: nested(p)}}}
}

func parseExistsGd(p *parser, nested func(*parser) Formula) Formula {
	p.expect(tokOpen)
	vars := parseTypedListString(p, tokQid)
	p.expect(tokClose)

	defer p.expect(tokClose)
	return &ExistsNode{QuantNode{vars, UnaryNode{Formula: nested(p)}}}
}

func parseEffect(p *parser) Formula {
	if p.acceptNamedList("and") {
		return parseAndEffect(p, parseCeffect)
	}
	return parseCeffect(p)
}

func parseAndEffect(p *parser, nested func(*parser) Formula) Formula {
	var fs []Formula
	for p.peek().typ == tokOpen {
		fs = append(fs, nested(p))
	}
	defer p.expect(tokClose)
	return &AndNode{MultiNode{fs}}
}

func parseCeffect(p *parser) (res Formula) {
	switch {
	case p.acceptNamedList("forall"):
		res = parseForallEffect(p, parseEffect)
	case p.acceptNamedList("when"):
		res = parseWhen(p, parseCondEffect)
	default:
		res = parsePeffect(p)
	}
	return
}

func parseForallEffect(p *parser, nested func(*parser) Formula) Formula {
	p.expect(tokOpen)
	vars := parseTypedListString(p, tokQid)
	p.expect(tokClose)

	defer p.expect(tokClose)
	return &ForallNode{QuantNode{vars, UnaryNode{Formula: nested(p)}}}
}

func parseWhen(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(tokClose)
	return &WhenNode{parseGd(p), UnaryNode{Formula: nested(p)}}
}

func parsePeffect(p *parser) Formula {
	if _, ok := AssignOps[p.peekn(2).text]; ok && p.peek().typ == tokOpen {
		return parseAssign(p)
	}
	if p.acceptNamedList("not") {
		defer p.expect(tokClose)
		return &NotNode{UnaryNode{parseProposition(p)}}
	}
	return parseProposition(p)
}

func parseAssign(p *parser) Formula {
	p.expect(tokOpen)
	res := &AssignNode{
		Op:   p.expect(tokId).text,
		Lval: parseFhead(p),
		Rval: parseFexp(p),
	}
	p.expect(tokClose)
	return res
}

func parseCondEffect(p *parser) Formula {
	if p.acceptNamedList("and") {
		return parseAndEffect(p, parsePeffect)
	}
	return parsePeffect(p)
}

func parseFhead(p *parser) string {
	if _, ok := p.accept(tokOpen); !ok {
		return p.expect(tokId).text
	}
	name := p.expect(tokId).text
	p.expect(tokClose)
	return name
}

func parseFexp(p *parser) string {
	return p.expect(tokNum).text
}

// ParseProblem returns a Problem parsed from
// the an io.Reader.  Returns the problem or an
// error.
func ParseProblem(file string, r io.Reader) (prob *Problem, err error) {
	p, err := newParser(file, r)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(parseError)
			if !ok {
				panic(r)
			}
		}
	}()
	p.expect(tokOpen)
	p.expectId("define")
	prob = new(Problem)
	prob.Name = parseProbName(p)
	prob.Domain = parseProbDomain(p)
	prob.Requirements = parseReqsDef(p)
	prob.Objects = parseObjsDecl(p)
	prob.Init = parseInit(p)
	prob.Goal = parseGoal(p)
	prob.Metric = parseMetric(p)

	p.expect(tokClose)
	return prob, nil
}

func parseProbName(p *parser) string {
	p.expect(tokOpen)
	p.expectId("problem")
	name := p.expect(tokId).text
	p.expect(tokClose)
	return name
}

func parseProbDomain(p *parser) string {
	p.expect(tokOpen)
	p.expectId(":domain")
	name := p.expect(tokId).text
	p.expect(tokClose)
	return name
}

func parseObjsDecl(p *parser) (objs []TypedName) {
	if p.acceptNamedList(":objects") {
		objs = parseTypedListString(p, tokId)
		p.expect(tokClose)
	}
	return
}

func parseInit(p *parser) (els []Formula) {
	p.expect(tokOpen)
	p.expectId(":init")
	for p.peek().typ == tokOpen {
		els = append(els, parseInitEl(p))
	}
	p.expect(tokClose)
	return
}

func parseInitEl(p *parser) (res Formula) {
	switch {
	case p.acceptNamedList("="):
		res = &AssignNode{
			Op:   "=",
			Lval: parseFhead(p),
			Rval: p.expect(tokNum).text,
		}
		p.expect(tokClose)
	case p.acceptNamedList("not"):
		res = &NotNode{UnaryNode{parseProposition(p)}}
		p.expect(tokClose)
	default:
		res = parseProposition(p)
	}
	return
}

func parseGoal(p *parser) Formula {
	p.expect(tokOpen)
	p.expectId(":goal")
	g := parsePreGd(p)
	p.expect(tokClose)
	return g
}

func parseMetric(p *parser) (m Metric) {
	if p.acceptNamedList(":metric") {
		m = MetricMinCost
		p.expectId("minimize")
		p.expect(tokOpen)
		p.expectId("total-cost")
		p.expect(tokClose)
		p.expect(tokClose)
	}
	return
}

func parseTypedListString(p *parser, typ tokenType) (lst []TypedName) {
	for {
		names := parseStrings(p, typ)
		if len(names) == 0 {
			break
		}
		typ := parseType(p)
		for _, n := range names {
			name := parseName(p, n)
			lst = append(lst, TypedName{Name: name, Types: typ})
		}
	}
	return lst
}

func parseType(p *parser) (typ []Name) {
	if _, ok := p.accept(tokMinus); !ok {
		return []Name{}
	}
	if _, ok := p.accept(tokOpen); ok {
		p.expectId("either")
		for _, s := range parseStringPlus(p, tokId) {
			typ = append(typ, parseName(p, s))
		}
		p.expect(tokClose)
		return typ
	}
	t := p.expect(tokId)
	return []Name{parseName(p, t.text)}
}

func parseStringPlus(p *parser, typ tokenType) []string {
	lst := []string{p.expect(typ).text}
	lst = append(lst, parseStrings(p, typ)...)
	return lst
}

func parseStrings(p *parser, typ tokenType) (lst []string) {
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		lst = append(lst, t.text)
	}
	return lst
}

func parseName(p *parser, text string) Name {
	return Name{text, p.loc()}
}
