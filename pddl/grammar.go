package pddl

import (
	"io"
)

// Parse returns either a domain, a problem or a parse error.
func Parse(file string, r io.Reader) (dom *Domain, prob *Problem, err error) {
	var p *parser
	if p, err = newParser(file, r); err != nil {
		return
	}
	if p.peekn(4).text == "domain" {
		dom, err = parseDomain(p)
	} else {
		prob, err = parseProblem(p)
	}
	return
}

func parseDomain(p *parser) (d *Domain, err error) {
	if _, err = p.expect(tokOpen, "define"); err != nil {
		return
	}
	d = new(Domain)
	if d.Identifier, err = parseDomainName(p); err != nil {
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
	_, err = p.expect(tokClose)
	return
}

func parseDomainName(p *parser) (id Identifier, err error) {
	if _, err = p.expect(tokOpen, "domain"); err != nil {
		return
	}
	if id, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseReqsDef(p *parser) (reqs []Identifier, err error) {
	if p.acceptNamedList(":requirements") {
		for p.peek().typ == tokCid {
			var id Identifier
			if id, err = parseIdentifier(p, tokCid); err != nil {
				return
			}
			reqs = append(reqs, id)
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parseTypesDef(p *parser) (types []Type, err error) {
	if p.acceptNamedList(":types") {
		var tlist []TypedIdentifier
		if tlist, err = parseTypedListString(p, tokId); err != nil {
			return
		}
		for _, t := range tlist {
			types = append(types, Type{TypedIdentifier: t})
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parseConstsDef(p *parser) (consts []TypedIdentifier, err error) {
	if p.acceptNamedList(":constants") {
		if consts, err = parseTypedListString(p, tokId); err != nil {
			return
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parsePredsDef(p *parser) (preds []Predicate, err error) {
	if p.acceptNamedList(":predicates") {
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
		_, err = p.expect(tokClose)
	}
	return preds, err
}

func parseAtomicFormSkele(p *parser) (pred Predicate, err error) {
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	if pred.Identifier, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	if pred.Parameters, err = parseTypedListString(p, tokQid); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseAtomicFuncSkele(p *parser) (fun Function, err error) {
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	if fun.Identifier, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	if fun.Parameters, err = parseTypedListString(p, tokQid); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseFuncsDef(p *parser) (funs []Function, err error) {
	if p.acceptNamedList(":functions") {
		if funs, err = parseFunctionTypedList(p); err != nil {
			return
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parseTypedListString(p *parser, typ tokenType) (lst []TypedIdentifier, err error) {
	for {
		ids := parseIdentifiers(p, typ)
		if len(ids) == 0 {
			break
		}
		var t []TypeName
		if t, err = parseType(p); err != nil {
			return
		}
		for _, id := range ids {
			lst = append(lst, TypedIdentifier{Identifier: id, Types: t})
		}
	}
	return
}

func parseType(p *parser) (typ []TypeName, err error) {
	if _, ok := p.accept(tokMinus); !ok {
		return
	}
	if _, ok := p.accept(tokOpen); ok {
		if _, err = p.expect("either"); err != nil {
			return
		}
		var ids []Identifier
		if ids, err = parseIdentifiersPlus(p, tokId); err != nil {
			return
		}
		for _, id := range ids {
			typ = append(typ, TypeName{Identifier: id})
		}
		_, err = p.expect(tokClose)
		return
	}
	id, err := parseIdentifier(p, tokId)
	return []TypeName{{Identifier: id}}, err
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
	if _, ok := p.accept(tokMinus); !ok {
		return
	}
	var t []token
	l := p.Loc()
	if t, err = p.expect("number"); err != nil {
		return
	}
	return []TypeName{{Identifier: Identifier{t[0].text, l}}}, nil
}

func parseActionDef(p *parser) (act Action, err error) {
	if _, err = p.expect(tokOpen, ":action"); err != nil {
		return
	}
	if act.Identifier, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	if act.Parameters, err = parseActParms(p); err != nil {
		return
	}
	if p.peek().text == ":precondition" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			if act.Precondition, err = parsePreGd(p); err != nil {
				return
			}
		}
	}
	if p.peek().text == ":effect" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			if act.Effect, err = parseEffect(p); err != nil {
				return
			}
		}
	}
	_, err = p.expect(tokClose)
	return
}

func parseActParms(p *parser) (parms []TypedIdentifier, err error) {
	if _, err = p.expect(":parameters", tokOpen); err != nil {
		return
	}
	if parms, err = parseTypedListString(p, tokQid); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parsePreGd(p *parser) (form Formula, err error) {
	switch {
	case p.acceptNamedList("and"):
		form, err = parseAndGd(p, parsePreGd)
	case p.acceptNamedList("forall"):
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
	case p.acceptNamedList("and"):
		form, err = parseAndGd(p, parseGd)
	case p.acceptNamedList("or"):
		form, err = parseOrGd(p, parseGd)
	case p.acceptNamedList("not"):
		form, err = parseNotGd(p)
		if err != nil {
			return
		}
		if lit, ok := form.(*LiteralNode); ok {
			lit.Negative = !lit.Negative
			form = lit
		}
	case p.acceptNamedList("imply"):
		form, err = parseImplyGd(p)
	case p.acceptNamedList("exists"):
		form, err = parseExistsGd(p, parseGd)
	case p.acceptNamedList("forall"):
		form, err = parseForallGd(p, parseGd)
	default:
		form, err = parseLiteral(p, false)
	}
	return
}

func parseLiteral(p *parser, eff bool) (lit *LiteralNode, err error) {
	lit = new(LiteralNode)
	lit.IsEffect = eff
	if p.acceptNamedList("not") {
		lit.Negative = true
	}
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	lit.Node = Node{p.Loc()}
	var id Identifier
	if id, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	lit.Predicate = id
	lit.Arguments = parseTerms(p)
	if _, err = p.expect(tokClose); err != nil {
		return
	}
	if lit.Negative {
		_, err = p.expect(tokClose)
	}
	return
}

func parseTerms(p *parser) (lst []Term) {
	for {
		l := p.Loc()
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, Term{Identifier: Identifier{t.text, l}})
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, Term{Identifier: Identifier{t.text, l}, Variable: true})
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
	_, err = p.expect(tokClose)
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
	_, err = p.expect(tokClose)
	return
}

func parseNotGd(p *parser) (form Formula, err error) {
	nd := Node{p.Loc()}
	var gd Formula
	if gd, err = parseGd(p); err != nil {
		return
	}
	form = &NotNode{UnaryNode{nd, gd}}
	_, err = p.expect(tokClose)
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
	_, err = p.expect(tokClose)
	return
}

func parseForallGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	var vars []TypedIdentifier
	if vars, err = parseTypedListString(p, tokQid); err != nil {
		return
	}
	if _, err = p.expect(tokClose); err != nil {
		return
	}
	var n Formula
	if n, err = nested(p); err != nil {
		return
	}
	form = &ForallNode{QuantNode{vars, UnaryNode{nd, n}}, false}
	_, err = p.expect(tokClose)
	return
}

func parseExistsGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	var vars []TypedIdentifier
	if vars, err = parseTypedListString(p, tokQid); err != nil {
		return
	}
	if _, err = p.expect(tokClose); err != nil {
		return
	}
	var n Formula
	if n, err = nested(p); err != nil {
		return
	}
	form = &ExistsNode{QuantNode{vars, UnaryNode{nd, n}}}
	_, err = p.expect(tokClose)
	return
}

func parseEffect(p *parser) (Formula, error) {
	if p.acceptNamedList("and") {
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
	_, err = p.expect(tokClose)
	return
}

func parseCeffect(p *parser) (form Formula, err error) {
	switch {
	case p.acceptNamedList("forall"):
		form, err = parseForallEffect(p, parseEffect)
	case p.acceptNamedList("when"):
		form, err = parseWhen(p, parseCondEffect)
	default:
		form, err = parsePeffect(p)
	}
	return
}

func parseForallEffect(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	var vars []TypedIdentifier
	if vars, err = parseTypedListString(p, tokQid); err != nil {
		return
	}
	if _, err = p.expect(tokClose); err != nil {
		return
	}
	var n Formula
	if n, err = nested(p); err != nil {
		return
	}
	form = &ForallNode{QuantNode{vars, UnaryNode{nd, n}}, true}
	_, err = p.expect(tokClose)
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
	_, err = p.expect(tokClose)
	return
}

func parsePeffect(p *parser) (form Formula, err error) {
	if _, ok := AssignOps[p.peekn(2).text]; ok && p.peek().typ == tokOpen {
		return parseAssign(p)
	}
	return parseLiteral(p, true)
}

func parseAssign(p *parser) (a *AssignNode, err error) {
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	a = new(AssignNode)
	if a.Op, err = parseIdentifier(p, tokId); err != nil {
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

	if n, ok := p.accept(tokNum); ok {
		a.IsNumber = true
		a.Number = n.text
	} else {
		if a.Fhead, err = parseFhead(p); err != nil {
			return
		}
	}
	_, err = p.expect(tokClose)
	return
}

func parseCondEffect(p *parser) (Formula, error) {
	if p.acceptNamedList("and") {
		return parseAndEffect(p, parsePeffect)
	}
	return parsePeffect(p)
}

func parseFhead(p *parser) (head Fhead, err error) {
	_, open := p.accept(tokOpen)
	if head.Identifier, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	if open {
		head.Arguments = parseTerms(p)
		_, err = p.expect(tokClose)
	}
	return
}

func parseProblem(p *parser) (prob *Problem, err error) {
	if _, err = p.expect(tokOpen, "define"); err != nil {
		return
	}
	prob = new(Problem)
	if prob.Identifier, err = parseProbName(p); err != nil {
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
	_, err = p.expect(tokClose)
	return
}

func parseProbName(p *parser) (id Identifier, err error) {
	if _, err = p.expect(tokOpen, "problem"); err != nil {
		return
	}
	if id, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseProbDomain(p *parser) (id Identifier, err error) {
	if _, err = p.expect(tokOpen, ":domain"); err != nil {
		return
	}
	if id, err = parseIdentifier(p, tokId); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseObjsDecl(p *parser) (objs []TypedIdentifier, err error) {
	if p.acceptNamedList(":objects") {
		if objs, err = parseTypedListString(p, tokId); err != nil {
			return
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parseInit(p *parser) (els []Formula, err error) {
	if _, err = p.expect(tokOpen, ":init"); err != nil {
		return
	}
	for p.peek().typ == tokOpen {
		var el Formula
		if el, err = parseInitEl(p); err != nil {
			return
		}
		els = append(els, el)
	}
	_, err = p.expect(tokClose)
	return
}

func parseInitEl(p *parser) (form Formula, err error) {
	if p.acceptNamedList("=") {
		nd := Node{p.Loc()}
		asn := &AssignNode{Op: Identifier{"=", nd.Loc()}, Node: nd}
		if asn.Lval, err = parseFhead(p); err != nil {
			return
		}
		var t []token
		if t, err = p.expect(tokNum); err != nil {
			return
		}
		asn.IsNumber = true
		asn.Number = t[0].text
		form = asn
		_, err = p.expect(tokClose)
		return
	}
	return parseLiteral(p, false)
}

func parseGoal(p *parser) (form Formula, err error) {
	if _, err = p.expect(tokOpen, ":goal"); err != nil {
		return
	}
	if form, err = parsePreGd(p); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseMetric(p *parser) (m Metric, err error) {
	if p.acceptNamedList(":metric") {
		m = MetricMinCost
		_, err = p.expect("minimize", tokOpen, "total-cost", tokClose, tokClose)
	}
	return
}

func parseIdentifiersPlus(p *parser, typ tokenType) (ids []Identifier, err error) {
	var id Identifier
	if id, err = parseIdentifier(p, typ); err != nil {
		return
	}
	ids = []Identifier{id}
	ids = append(ids, parseIdentifiers(p, typ)...)
	return
}

func parseIdentifiers(p *parser, typ tokenType) (ids []Identifier) {
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		l := p.Loc()
		ids = append(ids, Identifier{t.text, l})
	}
	return
}

func parseIdentifier(p *parser, typ tokenType) (Identifier, error) {
	l := p.Loc()
	id, err := p.expect(typ)
	if err != nil {
		return Identifier{}, err
	}
	return Identifier{id[0].text, l}, nil
}
