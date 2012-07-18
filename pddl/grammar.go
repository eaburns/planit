package pddl

import (
	"io"
	"log"
)

// Parse returns either a domain, a problem or a parse error.
func Parse(file string, r io.Reader) (ast interface{}, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(Error); ok {
			err = e
		} else {
			panic(r)
		}
	}()
	p, err := newParser(file, r)
	if err != nil {
		return
	}
	p.expect("(", "define")
	defer p.expect(")")

	if p.peekn(2).text == "domain" {
		return parseDomain(p), nil
	}
	return parseProblem(p), nil
}

func parseDomain(p *parser) *Domain {
	return &Domain{
		Name: parseDomainName(p),
		Requirements: parseReqsDef(p),
		Types: parseTypesDef(p),
		Constants: parseConstsDef(p),
		Predicates: parsePredsDef(p),
		Functions: parseFuncsDef(p),
		Actions: parseActionsDef(p),
	}
}

func parseDomainName(p *parser) Name {
	p.expect("(", "domain")
	defer p.expect(")")
	return parseName(p, tokName)
}

func parseReqsDef(p *parser) (reqs []Name) {
	if p.accept("(", ":requirements") {
		defer p.expect(")")
		for p.peek().typ == tokCname {
			reqs = append(reqs, parseName(p, tokCname))
		}
	}
	return
}

func parseTypesDef(p *parser) (types []Type) {
	if p.accept("(", ":types") {
		defer p.expect(")")
		for _, t := range parseTypedListString(p, tokName) {
			types = append(types, Type{TypedEntry: t})
		}
	}
	return
}

func parseConstsDef(p *parser) []TypedEntry {
	if p.accept("(", ":constants") {
		defer p.expect(")")
		return parseTypedListString(p, tokName)
	}
	return nil
}

func parsePredsDef(p *parser) []Predicate {
	if p.accept("(", ":predicates") {
		defer p.expect(")")
		preds := []Predicate{parseAtomicFormSkele(p)}
		for p.peek().typ == tokOpen {
			preds = append(preds, parseAtomicFormSkele(p))
		}
		return preds
	}
	return nil
}

func parseAtomicFormSkele(p *parser) Predicate {
	p.expect("(")
	defer p.expect(")")
	return Predicate{
		Name: parseName(p, tokName),
		Parameters: parseTypedListString(p, tokQname),
	}
}

func parseAtomicFuncSkele(p *parser) Function {
	p.expect("(")
	defer p.expect(")")
	return Function{
		Name: parseName(p, tokName),
		Parameters: parseTypedListString(p, tokQname),
	}
}

func parseFuncsDef(p *parser) []Function {
	if p.accept("(", ":functions") {
		defer p.expect(")")
		return parseFunctionTypedList(p)
	}
	return nil
}

func parseActionsDef(p *parser) (acts []Action) {
	for p.peek().typ == tokOpen {
		acts = append(acts, parseActionDef(p))
	}
	return
}

func parseTypedListString(p *parser, typ tokenType) (lst []TypedEntry) {
	for {
		ids := parseNames(p, typ)
		if len(ids) == 0 && p.peek().typ == tokMinus {
			log.Println("Parser hack: allowing an empty name list in front of a type in a typed list")
			log.Println("This seems to be required for IPC 2008 woodworking-strips/p11-domain.pddl")
		} else if len(ids) == 0 {
			break
		}
		t := parseType(p)
		for _, id := range ids {
			lst = append(lst, TypedEntry{Name: id, Types: t})
		}
	}
	return
}

func parseType(p *parser) (typ []TypeName) {
	if !p.accept("-") {
		return
	}
	if !p.accept("(") {
		return []TypeName{{Name: parseName(p, tokName)}}
	}
	p.expect("either")
	defer p.expect(")")
	for _, id := range parseNamesPlus(p, tokName) {
		typ = append(typ, TypeName{Name: id})
	}
	return
}

func parseFunctionTypedList(p *parser) (funs []Function) {
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
		funs = append(funs, fs...)
	}
	return
}

func parseFunctionType(p *parser) (typ []TypeName) {
	if !p.accept("-") {
		return
	}
	return []TypeName{TypeName{
		Name: Name{
			Location: p.Loc(),
			Str: p.expectText("number").text,
		},
	}}
}

func parseActionDef(p *parser) (act Action) {
	p.expect("(", ":action")
	defer p.expect(")")
	act.Name = parseName(p, tokName)
	act.Parameters = parseActParms(p)
	if p.accept(":precondition") {
		if !p.accept("(", ")") {
			act.Precondition = parsePreGd(p)
		}
	}
	if p.accept(":effect") {
		if !p.accept("(", ")") {
			act.Effect = parseEffect(p)
		}
	}
	return
}

func parseActParms(p *parser) (parms []TypedEntry) {
	p.expect(":parameters", "(")
	defer p.expect(")")
	return parseTypedListString(p, tokQname)
}

func parsePreGd(p *parser) Formula {
	switch {
	case p.accept("(", "and"):
		return parseAndGd(p, parsePreGd)
	case p.accept("(", "forall"):
		return parseForallGd(p, parsePreGd)
	}
	return parsePrefGd(p)
}

func parsePrefGd(p *parser) Formula {
	return parseGd(p)
}

func parseGd(p *parser) Formula {
	switch {
	case p.accept("(", "and"):
		return parseAndGd(p, parseGd)
	case p.accept("(", "or"):
		return parseOrGd(p, parseGd)
	case p.accept("(", "not"):
		form := parseNotGd(p)
		if lit, ok := form.(*LiteralNode); ok {
			lit.Negative = !lit.Negative
			return lit
		}
		return form
	case p.accept("(", "imply"):
		return parseImplyGd(p)
	case p.accept("(", "exists"):
		return parseExistsGd(p, parseGd)
	case p.accept("(", "forall"):
		return parseForallGd(p, parseGd)
	}
	return parseLiteral(p, false)
}

func parseLiteral(p *parser, eff bool) *LiteralNode {
	lit := new(LiteralNode)
	if p.accept("(", "not") {
		lit.Negative = true
		defer p.expect(")")
	}
	p.expect("(")
	defer p.expect(")")

	lit.IsEffect = eff
	lit.Node = Node{p.Loc()}
	if p.accept("=") {
		lit.Predicate = Name{"=", lit.Node.Loc()}
	} else {
		lit.Predicate = parseName(p, tokName)
	}
	lit.Arguments = parseTerms(p)
	return lit
}

func parseTerms(p *parser) (lst []Term) {
	for {
		l := p.Loc()
		if t, ok := p.acceptToken(tokName); ok {
			lst = append(lst, Term{Name: Name{t.text, l}})
			continue
		}
		if t, ok := p.acceptToken(tokQname); ok {
			lst = append(lst, Term{Name: Name{t.text, l}, Variable: true})
			continue
		}
		break
	}
	return
}

func parseAndGd(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")
	return &AndNode{MultiNode{
		Node: Node{ p.Loc() },
		Formula: parseFormulaStar(p, nested),
	}}
}

func parseFormulaStar(p *parser, nested func(*parser) Formula) (fs []Formula) {
	for p.peek().typ == tokOpen {
		fs = append(fs, nested(p))
	}
	return
}

func parseOrGd(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")
	return &OrNode{MultiNode{
		Node: Node{ p.Loc() },
		Formula: parseFormulaStar(p, nested),
	}}
}

func parseNotGd(p *parser) Formula {
	defer p.expect(")")
	return &NotNode{UnaryNode{
		Node: Node{p.Loc()},
		Formula: parseGd(p),
	}}
}

func parseImplyGd(p *parser) Formula {
	defer p.expect(")")
	return &ImplyNode{BinaryNode{
		Node: Node{p.Loc()},
		Left: parseGd(p),
		Right: parseGd(p),
	}}
}

func parseForallGd(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")

	loc := p.Loc()
	return &ForallNode{
		QuantNode: QuantNode{
			Variables: parseQuantVariables(p),
			UnaryNode: UnaryNode{Node{ loc }, nested(p)},
		},
		IsEffect: false,
	}
}

func parseQuantVariables(p *parser) []TypedEntry {
	p.expect("(")
	defer p.expect(")")
	return parseTypedListString(p, tokQname)
}

func parseExistsGd(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")
	loc := p.Loc()
	return &ExistsNode{QuantNode{
		Variables: parseQuantVariables(p),
		UnaryNode: UnaryNode{Node{ loc }, nested(p)},
	}}
}

func parseEffect(p *parser) Formula {
	if p.accept("(", "and") {
		return parseAndEffect(p, parseCeffect)
	}
	return parseCeffect(p)
}

func parseAndEffect(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")
	return &AndNode{MultiNode{
		Node: Node{ p.Loc() },
		Formula: parseFormulaStar(p, nested),
	}}
}

func parseCeffect(p *parser) Formula {
	switch {
	case p.accept("(", "forall"):
		return parseForallEffect(p, parseEffect)
	case p.accept("(", "when"):
		return parseWhen(p, parseCondEffect)
	}
	return parsePeffect(p)
}

func parseForallEffect(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")
	loc := p.Loc()
	return &ForallNode{
		QuantNode: QuantNode{
			Variables: parseQuantVariables(p),
			UnaryNode: UnaryNode{Node{ loc }, nested(p)},
		},
		IsEffect: true,
	}
}

func parseWhen(p *parser, nested func(*parser) Formula) Formula {
	defer p.expect(")")
	loc := p.Loc()
	return &WhenNode{
		Condition: parseGd(p),
		UnaryNode: UnaryNode{Node{ loc }, nested(p)},
	}
}

func parsePeffect(p *parser) Formula {
	if _, ok := AssignOps[p.peekn(2).text]; ok && p.peek().typ == tokOpen {
		return parseAssign(p)
	}
	return parseLiteral(p, true)
}

func parseAssign(p *parser) *AssignNode {
	p.expect("(")
	defer p.expect(")")
	a := new(AssignNode)
	a.Op = parseName(p, tokName)
	a.Lval = parseFhead(p)

	// f-exp:
	// We support :action-costs, which means that
	// an Fexp can be either a non-negative number
	// (non-negativity is checked during semantic
	// analysis) or it can be of the form:
	// 	(<function-symbol> <term>*)
	// i.e., an f-head

	if n, ok := p.acceptToken(tokNum); ok {
		a.IsNumber = true
		a.Number = n.text
	} else {
		a.Fhead = parseFhead(p)
	}
	return a
}

func parseCondEffect(p *parser) Formula {
	if p.accept("(", "and") {
		return parseAndEffect(p, parsePeffect)
	}
	return parsePeffect(p)
}

func parseFhead(p *parser) (head Fhead) {
	open := p.accept("(")
	head.Name = parseName(p, tokName)
	if open {
		head.Arguments = parseTerms(p)
		p.expect(")")
	}
	return
}

func parseProblem(p *parser) *Problem {
	return &Problem {
		Name: parseProbName(p),
		Domain: parseProbDomain(p),
		Requirements: parseReqsDef(p),
		Objects: parseObjsDecl(p),
		Init: parseInit(p),
		Goal: parseGoal(p),
		Metric: parseMetric(p),
	}
}

func parseProbName(p *parser) Name {
	p.expect("(", "problem")
	defer p.expect(")")
	return parseName(p, tokName)
}

func parseProbDomain(p *parser) Name {
	p.expect("(", ":domain")
	defer p.expect(")")
	return parseName(p, tokName)
}

func parseObjsDecl(p *parser) []TypedEntry {
	if p.accept("(", ":objects") {
		defer p.expect(")")
		return parseTypedListString(p, tokName)
	}
	return nil
}

func parseInit(p *parser) (els []Formula) {
	p.expect("(", ":init")
	defer p.expect(")")
	for p.peek().typ == tokOpen {
		els = append(els, parseInitEl(p))
	}
	return
}

func parseInitEl(p *parser) Formula {
	loc := p.Loc()
	if p.accept("(", "=") {
		defer p.expect(")")
		return &AssignNode{
			Node: Node{loc},
			Op: Name{"=", p.Loc()},
			Lval: parseFhead(p),
			IsNumber: true,
			Number: p.expectType(tokNum).text,
			IsInit: true,
		}
	}
	return parseLiteral(p, false)
}

func parseGoal(p *parser) Formula {
	p.expect("(", ":goal")
	defer p.expect(")")
	return parsePreGd(p)
}

func parseMetric(p *parser) Metric {
	if p.accept("(", ":metric") {
		p.expect("minimize", "(", "total-cost", ")", ")")
		return MetricMinCost
	}
	return MetricMakespan
}

func parseNamesPlus(p *parser, typ tokenType) []Name {
	return append([]Name{parseName(p, typ)}, parseNames(p, typ)...)
}

func parseNames(p *parser, typ tokenType) (ids []Name) {
	for t, ok := p.acceptToken(typ); ok; t, ok = p.acceptToken(typ) {
		l := p.Loc()
		ids = append(ids, Name{t.text, l})
	}
	return
}

func parseName(p *parser, typ tokenType) Name {
	return Name{
		Location: p.Loc(),
		Str: p.expectType(typ).text,
	}
}
