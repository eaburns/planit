package pddl

import (
	"io"
	"log"
)

// Parse returns either a domain, a problem or a parse error.
func Parse(file string, r io.Reader) (ast interface{}, err error) {
	var p *parser
	if p, err = newParser(file, r); err != nil {
		return
	}
	if p.peekn(4).text == "domain" {
		ast, err = parseDomain(p)
	} else {
		ast, err = parseProblem(p)
	}
	return
}

func parseDomain(p *parser) (d *Domain, err error) {
	if err = p.expect("(", "define"); err != nil {
		return
	}
	d = new(Domain)
	if d.Name, err = parseDomainName(p); err != nil {
		return
	}
	if d.Requirements, err = parseReqsDef(p); err != nil {
		return
	}
	if d.Types, err = parseTypesDef(p); err != nil {
		return
	}
	if d.Constants, err = parseConstsDef(p); err != nil {
		return
	}
	if d.Predicates, err = parsePredsDef(p); err != nil {
		return
	}
	if d.Functions, err = parseFuncsDef(p); err != nil {
		return
	}
	for p.peek().typ == tokOpen {
		var act Action
		if act, err = parseActionDef(p); err != nil {
			return
		}
		d.Actions = append(d.Actions, act)
	}
	err = p.expect(")")
	return
}

func parseDomainName(p *parser) (id Name, err error) {
	if err = p.expect("(", "domain"); err != nil {
		return
	}
	if id, err = parseName(p, tokName); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseReqsDef(p *parser) (reqs []Name, err error) {
	if p.accept("(", ":requirements") {
		for p.peek().typ == tokCname {
			var id Name
			if id, err = parseName(p, tokCname); err != nil {
				return
			}
			reqs = append(reqs, id)
		}
		err = p.expect(")")
	}
	return
}

func parseTypesDef(p *parser) (types []Type, err error) {
	if p.accept("(", ":types") {
		var tlist []TypedEntry
		if tlist, err = parseTypedListString(p, tokName); err != nil {
			return
		}
		for _, t := range tlist {
			types = append(types, Type{TypedEntry: t})
		}
		err = p.expect(")")
	}
	return
}

func parseConstsDef(p *parser) (consts []TypedEntry, err error) {
	if p.accept("(", ":constants") {
		if consts, err = parseTypedListString(p, tokName); err != nil {
			return
		}
		err = p.expect(")")
	}
	return
}

func parsePredsDef(p *parser) (preds []Predicate, err error) {
	if p.accept("(", ":predicates") {
		var pred Predicate
		if pred, err = parseAtomicFormSkele(p); err != nil {
			return
		}
		preds = append(preds, pred)
		for p.peek().typ == tokOpen {
			if pred, err = parseAtomicFormSkele(p); err != nil {
				return
			}
			preds = append(preds, pred)
		}
		err = p.expect(")")
	}
	return preds, err
}

func parseAtomicFormSkele(p *parser) (pred Predicate, err error) {
	if err = p.expect("("); err != nil {
		return
	}
	if pred.Name, err = parseName(p, tokName); err != nil {
		return
	}
	if pred.Parameters, err = parseTypedListString(p, tokQname); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseAtomicFuncSkele(p *parser) (fun Function, err error) {
	if err = p.expect("("); err != nil {
		return
	}
	if fun.Name, err = parseName(p, tokName); err != nil {
		return
	}
	if fun.Parameters, err = parseTypedListString(p, tokQname); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseFuncsDef(p *parser) (funs []Function, err error) {
	if p.accept("(", ":functions") {
		if funs, err = parseFunctionTypedList(p); err != nil {
			return
		}
		err = p.expect(")")
	}
	return
}

func parseTypedListString(p *parser, typ tokenType) (lst []TypedEntry, err error) {
	for {
		ids := parseNames(p, typ)
		if len(ids) == 0 && p.peek().typ == tokMinus {
			log.Println("Parser hack: allowing an empty name list in front of a type in a typed list")
			log.Println("This seems to be required for IPC 2008 woodworking-strips/p11-domain.pddl")
		} else if len(ids) == 0 {
			break
		}
		var t []TypeName
		if t, err = parseType(p); err != nil {
			return
		}
		for _, id := range ids {
			lst = append(lst, TypedEntry{Name: id, Types: t})
		}
	}
	return
}

func parseType(p *parser) (typ []TypeName, err error) {
	if !p.accept(tokMinus) {
		return
	}
	if p.accept("(") {
		if err = p.expect("either"); err != nil {
			return
		}
		var ids []Name
		if ids, err = parseNamesPlus(p, tokName); err != nil {
			return
		}
		for _, id := range ids {
			typ = append(typ, TypeName{Name: id})
		}
		err = p.expect(")")
		return
	}
	id, err := parseName(p, tokName)
	return []TypeName{{Name: id}}, err
}

func parseFunctionTypedList(p *parser) (funs []Function, err error) {
	for {
		var fs []Function
		for p.peek().typ == tokOpen {
			var fun Function
			if fun, err = parseAtomicFuncSkele(p); err != nil {
				return
			}
			fs = append(fs, fun)
		}
		if len(fs) == 0 {
			break
		}
		var typ []TypeName
		if typ, err = parseFunctionType(p); err != nil {
			return
		}
		for i, _ := range fs {
			fs[i].Types = typ
		}
		funs = append(funs, fs...)
	}
	return
}

func parseFunctionType(p *parser) (typ []TypeName, err error) {
	if !p.accept(tokMinus) {
		return
	}
	var t []token
	l := p.Loc()
	if t, err = p.expectTokens("number"); err != nil {
		return
	}
	return []TypeName{{Name: Name{t[0].text, l}}}, nil
}

func parseActionDef(p *parser) (act Action, err error) {
	if err = p.expect("(", ":action"); err != nil {
		return
	}
	if act.Name, err = parseName(p, tokName); err != nil {
		return
	}
	if act.Parameters, err = parseActParms(p); err != nil {
		return
	}
	if p.accept(":precondition") {
		if !p.accept("(", ")") {
			if act.Precondition, err = parsePreGd(p); err != nil {
				return
			}
		}
	}
	if p.accept(":effect") {
		if !p.accept("(", ")") {
			if act.Effect, err = parseEffect(p); err != nil {
				return
			}
		}
	}
	err = p.expect(")")
	return
}

func parseActParms(p *parser) (parms []TypedEntry, err error) {
	if err = p.expect(":parameters", "("); err != nil {
		return
	}
	if parms, err = parseTypedListString(p, tokQname); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parsePreGd(p *parser) (form Formula, err error) {
	switch {
	case p.accept("(", "and"):
		form, err = parseAndGd(p, parsePreGd)
	case p.accept("(", "forall"):
		form, err = parseForallGd(p, parsePreGd)
	default:
		form, err = parsePrefGd(p)
	}
	return
}

func parsePrefGd(p *parser) (Formula, error) {
	return parseGd(p)
}

func parseGd(p *parser) (form Formula, err error) {
	switch {
	case p.accept("(", "and"):
		form, err = parseAndGd(p, parseGd)
	case p.accept("(", "or"):
		form, err = parseOrGd(p, parseGd)
	case p.accept("(", "not"):
		form, err = parseNotGd(p)
		if err != nil {
			return
		}
		if lit, ok := form.(*LiteralNode); ok {
			lit.Negative = !lit.Negative
			form = lit
		}
	case p.accept("(", "imply"):
		form, err = parseImplyGd(p)
	case p.accept("(", "exists"):
		form, err = parseExistsGd(p, parseGd)
	case p.accept("(", "forall"):
		form, err = parseForallGd(p, parseGd)
	default:
		form, err = parseLiteral(p, false)
	}
	return
}

func parseLiteral(p *parser, eff bool) (lit *LiteralNode, err error) {
	lit = new(LiteralNode)
	lit.IsEffect = eff
	if p.accept("(", "not") {
		lit.Negative = true
	}
	if err = p.expect("("); err != nil {
		return
	}
	lit.Node = Node{p.Loc()}
	var id Name
	if p.accept(tokEq) {
		id = Name{"=", lit.Node.Loc()}
	} else if id, err = parseName(p, tokName); err != nil {
		return
	}
	lit.Predicate = id
	lit.Arguments = parseTerms(p)
	if err = p.expect(")"); err != nil {
		return
	}
	if lit.Negative {
		err = p.expect(")")
	}
	return
}

func parseTerms(p *parser) (lst []Term) {
	for {
		l := p.Loc()
		if t, ok := p.acceptTokens(tokName); ok {
			lst = append(lst, Term{Name: Name{t[0].text, l}})
			continue
		}
		if t, ok := p.acceptTokens(tokQname); ok {
			lst = append(lst, Term{Name: Name{t[0].text, l}, Variable: true})
			continue
		}
		break
	}
	return
}

type nested func(*parser) (Formula, error)

func parseAndGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	var fs []Formula
	for p.peek().typ == tokOpen {
		var n Formula
		if n, err = nested(p); err != nil {
			return
		}
		fs = append(fs, n)
	}
	form = &AndNode{MultiNode{nd, fs}}
	err = p.expect(")")
	return
}

func parseOrGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	var fs []Formula
	for p.peek().typ == tokOpen {
		var n Formula
		if n, err = nested(p); err != nil {
			return
		}
		fs = append(fs, n)
	}
	form = &OrNode{MultiNode{nd, fs}}
	err = p.expect(")")
	return
}

func parseNotGd(p *parser) (form Formula, err error) {
	nd := Node{p.Loc()}
	var gd Formula
	if gd, err = parseGd(p); err != nil {
		return
	}
	form = &NotNode{UnaryNode{nd, gd}}
	err = p.expect(")")
	return
}

func parseImplyGd(p *parser) (form Formula, err error) {
	nd := Node{p.Loc()}
	var left, right Formula
	if left, err = parseGd(p); err != nil {
		return
	}
	if right, err = parseGd(p); err != nil {
		return
	}
	form = &ImplyNode{BinaryNode{nd, left, right}}
	err = p.expect(")")
	return
}

func parseForallGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if err = p.expect("("); err != nil {
		return
	}
	var vars []TypedEntry
	if vars, err = parseTypedListString(p, tokQname); err != nil {
		return
	}
	if err = p.expect(")"); err != nil {
		return
	}
	var n Formula
	if n, err = nested(p); err != nil {
		return
	}
	form = &ForallNode{QuantNode{vars, UnaryNode{nd, n}}, false}
	err = p.expect(")")
	return
}

func parseExistsGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if err = p.expect("("); err != nil {
		return
	}
	var vars []TypedEntry
	if vars, err = parseTypedListString(p, tokQname); err != nil {
		return
	}
	if err = p.expect(")"); err != nil {
		return
	}
	var n Formula
	if n, err = nested(p); err != nil {
		return
	}
	form = &ExistsNode{QuantNode{vars, UnaryNode{nd, n}}}
	err = p.expect(")")
	return
}

func parseEffect(p *parser) (Formula, error) {
	if p.accept("(", "and") {
		return parseAndEffect(p, parseCeffect)
	}
	return parseCeffect(p)
}

func parseAndEffect(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	var fs []Formula
	for p.peek().typ == tokOpen {
		var f Formula
		if f, err = nested(p); err != nil {
			return
		}
		fs = append(fs, f)
	}
	form = &AndNode{MultiNode{nd, fs}}
	err = p.expect(")")
	return
}

func parseCeffect(p *parser) (form Formula, err error) {
	switch {
	case p.accept("(", "forall"):
		form, err = parseForallEffect(p, parseEffect)
	case p.accept("(", "when"):
		form, err = parseWhen(p, parseCondEffect)
	default:
		form, err = parsePeffect(p)
	}
	return
}

func parseForallEffect(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if err = p.expect("("); err != nil {
		return
	}
	var vars []TypedEntry
	if vars, err = parseTypedListString(p, tokQname); err != nil {
		return
	}
	if err = p.expect(")"); err != nil {
		return
	}
	var n Formula
	if n, err = nested(p); err != nil {
		return
	}
	form = &ForallNode{QuantNode{vars, UnaryNode{nd, n}}, true}
	err = p.expect(")")
	return
}

func parseWhen(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	var gd, n Formula
	if gd, err = parseGd(p); err != nil {
		return
	}
	if n, err = nested(p); err != nil {
		return
	}
	form = &WhenNode{gd, UnaryNode{nd, n}}
	err = p.expect(")")
	return
}

func parsePeffect(p *parser) (form Formula, err error) {
	if _, ok := AssignOps[p.peekn(2).text]; ok && p.peek().typ == tokOpen {
		return parseAssign(p)
	}
	return parseLiteral(p, true)
}

func parseAssign(p *parser) (a *AssignNode, err error) {
	if err = p.expect("("); err != nil {
		return
	}
	a = new(AssignNode)
	if a.Op, err = parseName(p, tokName); err != nil {
		return
	}
	if a.Lval, err = parseFhead(p); err != nil {
		return
	}

	// f-exp:
	// We support :action-costs, which means that
	// an Fexp can be either a non-negative number
	// (non-negativity is checked during semantic
	// analysis) or it can be of the form:
	// 	(<function-symbol> <term>*)
	// i.e., an f-head

	if n, ok := p.acceptTokens(tokNum); ok {
		a.IsNumber = true
		a.Number = n[0].text
	} else {
		if a.Fhead, err = parseFhead(p); err != nil {
			return
		}
	}
	err = p.expect(")")
	return
}

func parseCondEffect(p *parser) (Formula, error) {
	if p.accept("(", "and") {
		return parseAndEffect(p, parsePeffect)
	}
	return parsePeffect(p)
}

func parseFhead(p *parser) (head Fhead, err error) {
	open := p.accept("(")
	if head.Name, err = parseName(p, tokName); err != nil {
		return
	}
	if open {
		head.Arguments = parseTerms(p)
		err = p.expect(")")
	}
	return
}

func parseProblem(p *parser) (prob *Problem, err error) {
	if err = p.expect("(", "define"); err != nil {
		return
	}
	prob = new(Problem)
	if prob.Name, err = parseProbName(p); err != nil {
		return
	}
	if prob.Domain, err = parseProbDomain(p); err != nil {
		return
	}
	if prob.Requirements, err = parseReqsDef(p); err != nil {
		return
	}
	if prob.Objects, err = parseObjsDecl(p); err != nil {
		return
	}
	if prob.Init, err = parseInit(p); err != nil {
		return
	}
	if prob.Goal, err = parseGoal(p); err != nil {
		return
	}
	if prob.Metric, err = parseMetric(p); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseProbName(p *parser) (id Name, err error) {
	if err = p.expect("(", "problem"); err != nil {
		return
	}
	if id, err = parseName(p, tokName); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseProbDomain(p *parser) (id Name, err error) {
	if err = p.expect("(", ":domain"); err != nil {
		return
	}
	if id, err = parseName(p, tokName); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseObjsDecl(p *parser) (objs []TypedEntry, err error) {
	if p.accept("(", ":objects") {
		if objs, err = parseTypedListString(p, tokName); err != nil {
			return
		}
		err = p.expect(")")
	}
	return
}

func parseInit(p *parser) (els []Formula, err error) {
	if err = p.expect("(", ":init"); err != nil {
		return
	}
	for p.peek().typ == tokOpen {
		var el Formula
		if el, err = parseInitEl(p); err != nil {
			return
		}
		els = append(els, el)
	}
	err = p.expect(")")
	return
}

func parseInitEl(p *parser) (form Formula, err error) {
	if p.accept("(", "=") {
		nd := Node{p.Loc()}
		asn := &AssignNode{Op: Name{"=", nd.Loc()}, Node: nd}
		if asn.Lval, err = parseFhead(p); err != nil {
			return
		}
		var t []token
		if t, err = p.expectTokens(tokNum); err != nil {
			return
		}
		asn.IsNumber = true
		asn.Number = t[0].text
		asn.IsInit = true
		form = asn
		err = p.expect(")")
		return
	}
	return parseLiteral(p, false)
}

func parseGoal(p *parser) (form Formula, err error) {
	if err = p.expect("(", ":goal"); err != nil {
		return
	}
	if form, err = parsePreGd(p); err != nil {
		return
	}
	err = p.expect(")")
	return
}

func parseMetric(p *parser) (m Metric, err error) {
	if p.accept("(", ":metric") {
		m = MetricMinCost
		err = p.expect("minimize", "(", "total-cost", ")", ")")
	}
	return
}

func parseNamesPlus(p *parser, typ tokenType) (ids []Name, err error) {
	var id Name
	if id, err = parseName(p, typ); err != nil {
		return
	}
	ids = []Name{id}
	ids = append(ids, parseNames(p, typ)...)
	return
}

func parseNames(p *parser, typ tokenType) (ids []Name) {
	for t, ok := p.acceptTokens(typ); ok; t, ok = p.acceptTokens(typ) {
		l := p.Loc()
		ids = append(ids, Name{t[0].text, l})
	}
	return
}

func parseName(p *parser, typ tokenType) (Name, error) {
	l := p.Loc()
	id, err := p.expectTokens(typ)
	if err != nil {
		return Name{}, err
	}
	return Name{id[0].text, l}, nil
}
