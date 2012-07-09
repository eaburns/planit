package pddl

import (
	"io"
	"log"
)

// Parse returns either a domain, a problem or
// a parse error.
func Parse(file string, r io.Reader) (dom *Domain, prob *Problem, err error) {
	p, err := newParser(file, r)
	if err != nil {
		return nil, nil, err
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
	if p.peekn(4).text == "domain" {
		dom = parseDomain(p)
	} else {
		prob = parseProblem(p)
	}
	return
}

// parseDomain parses a domain.
func parseDomain(p *parser) *Domain {
	p.expect(tokOpen)
	p.expectId("define")
	d := new(Domain)
	d.Name = parseDomainName(p)
	d.Requirements = parseReqsDef(p)
	d.Types = parseTypesDef(p)
	d.Constants = parseConstsDef(p)
	d.Predicates = parsePredsDef(p)
	d.Functions = parseFuncsDef(p)

	// Ignore :functions for now
	if p.acceptNamedList(":functions") {
		log.Println(p.Loc(), "ignoring functions declaration")
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
	return d
}

func parseDomainName(p *parser) string {
	p.expect(tokOpen)
	p.expectId("domain")
	n := p.expect(tokId)
	p.expect(tokClose)
	return n.text
}

func parseReqsDef(p *parser) (reqs []Name) {
	if p.acceptNamedList(":requirements") {
		for t, ok := p.accept(tokCid); ok; t, ok = p.accept(tokCid) {
			reqs = append(reqs, makeName(p, t.text))
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

func parsePredsDef(p *parser) (predicates []Predicate) {
	if p.acceptNamedList(":predicates") {
		predicates = append(predicates, parseAtomicFormSkele(p))
		for p.peek().typ == tokOpen {
			predicates = append(predicates, parseAtomicFormSkele(p))
		}
		p.expect(tokClose)
	}
	return
}

func parseAtomicFormSkele(p *parser) Predicate {
	p.expect(tokOpen)
	defer p.expect(tokClose)
	return Predicate{
		Name:       makeName(p, p.expect(tokId).text),
		Parameters: parseTypedListString(p, tokQid),
	}
}

func parseAtomicFuncSkele(p *parser) Function {
	p.expect(tokOpen)
	defer p.expect(tokClose)
	return Function{
		Name:       makeName(p, p.expect(tokId).text),
		Parameters: parseTypedListString(p, tokQid),
	}
}

func parseFuncsDef(p *parser) (funcs []Function) {
	if p.acceptNamedList(":functions") {
		funcs = append(funcs, parseFunctionTypedList(p)...)
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
			name := makeName(p, n)
			lst = append(lst, TypedName{Name: name, Types: typ})
		}
	}
	return
}

func parseType(p *parser) (typ []Type) {
	if _, ok := p.accept(tokMinus); !ok {
		return []Type{}
	}
	if _, ok := p.accept(tokOpen); ok {
		p.expectId("either")
		for _, s := range parseStringPlus(p, tokId) {
			typ = append(typ, Type{Name: makeName(p, s)})
		}
		p.expect(tokClose)
		return typ
	}
	t := p.expect(tokId)
	return []Type{{Name: makeName(p, t.text)}}
}

func parseFunctionTypedList(p *parser) (funcs []Function) {
	for {
		var fs []Function
		for p.peek().typ == tokOpen {
			fs = append(fs, parseAtomicFuncSkele(p))
		}
		if len(fs) == 0 {
			break
		}
		typ := parseFunctionType(p)
		for i, _ := range fs {
			fs[i].Types = typ
		}
		funcs = append(funcs, fs...)
	}
	return
}

func parseFunctionType(p *parser) (typ []Type) {
	if _, ok := p.accept(tokMinus); !ok {
		return []Type{}
	}
	t := p.expectId("number")
	return []Type{{Name: makeName(p, t.text)}}
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
		res = &NotNode{UnaryNode{Node{p.Loc()}, parseGd(p)}}
		p.expect(tokClose)
	case p.acceptNamedList("imply"):
		res = &ImplyNode{BinaryNode{Node{p.Loc()}, parseGd(p), parseGd(p)}}
		p.expect(tokClose)
	case p.acceptNamedList("exists"):
		res = parseExistsGd(p, parseGd)
	case p.acceptNamedList("forall"):
		res = parseForallGd(p, parseGd)
	case p.acceptNamedList("not"):
		res = &NotNode{UnaryNode{Node{p.Loc()}, parseProposition(p)}}
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
		Node:      Node{p.Loc()},
		Predicate: p.expect(tokId).text,
		Arguments: parseTerms(p),
	}
}

func parseTerms(p *parser) (lst []Term) {
	for {
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, Term{Name: makeName(p, t.text)})
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, Term{Name: makeName(p, t.text), Variable: true})
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
	return &AndNode{MultiNode{Node{p.Loc()}, fs}}
}

func parseOrGd(p *parser, nested func(*parser) Formula) Formula {
	var fs []Formula
	for p.peek().typ == tokOpen {
		fs = append(fs, nested(p))
	}
	defer p.expect(tokClose)
	return &OrNode{MultiNode{Node{p.Loc()}, fs}}
}

func parseForallGd(p *parser, nested func(*parser) Formula) Formula {
	p.expect(tokOpen)
	vars := parseTypedListString(p, tokQid)
	p.expect(tokClose)

	defer p.expect(tokClose)
	return &ForallNode{QuantNode{vars, UnaryNode{Formula: nested(p)}}, false}
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
	return &AndNode{MultiNode{Node{p.Loc()}, fs}}
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
	return &ForallNode{QuantNode{vars, UnaryNode{Formula: nested(p)}}, true}
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
		return &NotNode{UnaryNode{Node{p.Loc()}, parseProposition(p)}}
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

func parseFhead(p *parser) Name {
	if _, ok := p.accept(tokOpen); !ok {
		return makeName(p, p.expect(tokId).text)
	}
	defer p.expect(tokClose)
	return makeName(p, p.expect(tokId).text)
}

func parseFexp(p *parser) string {
	return p.expect(tokNum).text
}

// parseProblem parses a problem
func parseProblem(p *parser) *Problem {
	p.expect(tokOpen)
	p.expectId("define")
	prob := new(Problem)
	prob.Name = parseProbName(p)
	prob.Domain = parseProbDomain(p)
	prob.Requirements = parseReqsDef(p)
	prob.Objects = parseObjsDecl(p)
	prob.Init = parseInit(p)
	prob.Goal = parseGoal(p)
	prob.Metric = parseMetric(p)

	p.expect(tokClose)
	return prob
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
		res = &NotNode{UnaryNode{Node{p.Loc()}, parseProposition(p)}}
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

func makeName(p *parser, text string) Name {
	return Name{text, p.Loc()}
}
