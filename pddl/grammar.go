package pddl

import (
	"io"
)

// Parse returns either a domain, a problem or
// a parse error.
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

// parseDomain parses a domain.
func parseDomain(p *parser) (d *Domain, err error) {
	if _, err = p.expect(tokOpen, "define"); err != nil {
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

	_, err = p.expect(tokClose)
	return
}

func parseDomainName(p *parser) (n Name, err error) {
	if _, err = p.expect(tokOpen, "domain"); err != nil {
		return
	}
	if n, err = parseName(p, tokId); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseReqsDef(p *parser) (reqs []Name, err error) {
	if p.acceptNamedList(":requirements") {
		for p.peek().typ == tokCid {
			var n Name
			if n, err = parseName(p, tokCid); err != nil {
				return
			}
			reqs = append(reqs, n)
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parseTypesDef(p *parser) (types []TypedName, err error) {
	if p.acceptNamedList(":types") {
		if types, err = parseTypedListString(p, tokId); err != nil {
			return
		}
		_, err = p.expect(tokClose)
	}
	return
}

func parseConstsDef(p *parser) (consts []TypedName, err error) {
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
	if pred.Name, err = parseName(p, tokId); err != nil {
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
	if fun.Name, err = parseName(p, tokId); err != nil {
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

func parseTypedListString(p *parser, typ tokenType) (lst []TypedName, err error) {
	for {
		names := parseStrings(p, typ)
		if len(names) == 0 {
			break
		}
		var t []Type
		if t, err = parseType(p); err != nil {
			return
		}
		for _, n := range names {
			lst = append(lst, TypedName{Name: n, Types: t})
		}
	}
	return
}

func parseType(p *parser) (typ []Type, err error) {
	if _, ok := p.accept(tokMinus); !ok {
		return
	}
	if _, ok := p.accept(tokOpen); ok {
		if _, err = p.expect("either"); err != nil {
			return
		}
		var names []Name
		if names, err = parseStringPlus(p, tokId); err != nil {
			return
		}
		for _, n := range names  {
			typ = append(typ, Type{Name: n})
		}
		_, err = p.expect(tokClose)
		return
	}
	n, err := parseName(p, tokId)
	return []Type{{Name: n}}, err
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
		var typ []Type
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

func parseFunctionType(p *parser) (typ []Type, err error) {
	if _, ok := p.accept(tokMinus); !ok {
		return
	}
	var t []token
	l := p.Loc()
	if t, err = p.expect("number"); err != nil {
		return
	}
	return []Type{{Name: Name{t[0].text, l}}}, nil
}

func parseActionDef(p *parser) (act Action, err error) {
	if _, err = p.expect(tokOpen, ":action"); err != nil {
		return
	}
	if act.Name, err = parseName(p, tokId); err != nil {
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
			if act.Effect, err= parseEffect(p); err != nil {
				return
			}
		}
	}

	_, err = p.expect(tokClose)
	return
}

func parseActParms(p *parser) (parms []TypedName, err error) {
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
		nd := Node{p.Loc()}
		var gd Formula
		if gd, err = parseGd(p); err != nil {
			return
		}
		form = &NotNode{UnaryNode{nd, gd}}
		_, err = p.expect(tokClose)
	case p.acceptNamedList("imply"):
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
	case p.acceptNamedList("exists"):
		form, err = parseExistsGd(p, parseGd)
	case p.acceptNamedList("forall"):
		form, err = parseForallGd(p, parseGd)
	case p.acceptNamedList("not"):
		nd := Node{p.Loc()}
		var prop *PropositionNode
		if prop, err = parseProposition(p); err != nil {
			return
		}
		form = &NotNode{UnaryNode{nd, prop}}
		_, err = p.expect(tokClose)
	default:
		form, err = parseProposition(p)
	}
	return
}

func parseProposition(p *parser) (prop *PropositionNode, err error) {
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	prop = new(PropositionNode)
	prop.Node = Node{p.Loc()}
	var n Name
	if n, err = parseName(p, tokId); err != nil {
		return
	}
	prop.Predicate = n
	prop.Arguments = parseTerms(p)

	_, err = p.expect(tokClose)
	return
}

func parseTerms(p *parser) (lst []Term) {
	for {
		l := p.Loc()
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, Term{Name: Name{ t.text, l }})
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, Term{Name: Name{ t.text, l }, Variable: true})
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

func parseForallGd(p *parser, nested nested) (form Formula, err error) {
	nd := Node{p.Loc()}
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	var vars []TypedName
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
	var vars []TypedName
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
	var vars []TypedName
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
	if p.acceptNamedList("not") {
		nd := Node{p.Loc()}
		var prop *PropositionNode
		if prop, err = parseProposition(p); err != nil {
			return
		}
		form = &NotNode{UnaryNode{nd, prop}}
		_, err = p.expect(tokClose)
	}
	return parseProposition(p)
}

func parseAssign(p *parser) (a *AssignNode, err error) {
	if _, err = p.expect(tokOpen); err != nil {
		return
	}
	a = new(AssignNode)
	if a.Op, err = parseName(p, tokId); err != nil {
		return
	}
	if a.Lval, err = parseFhead(p); err != nil {
		return
	}
	if a.Rval, err = parseFexp(p); err != nil {
		return
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

func parseFhead(p *parser) (n Name, err error) {
	_, open := p.accept(tokOpen)
	if n, err = parseName(p, tokId); err != nil {
		return
	}
	if open {
		_, err = p.expect(tokClose)
	}
	return
}

func parseFexp(p *parser) (string, error) {
	n, err := p.expect(tokNum)
	if err != nil {
		return "", err
	}
	return n[0].text, nil
}

// parseProblem parses a problem
func parseProblem(p *parser) (prob *Problem, err error) {
	if _, err = p.expect(tokOpen, "define"); err != nil {
		return
	}
	prob = new(Problem)
	if prob.Name, err  = parseProbName(p); err != nil {
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

func parseProbName(p *parser) (n Name, err error) {
	if _, err = p.expect(tokOpen, "problem"); err != nil {
		return
	}
	if n, err = parseName(p, tokId); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseProbDomain(p *parser) (n Name, err error) {
	if _, err = p.expect(tokOpen, ":domain"); err != nil {
		return
	}
	if n, err = parseName(p, tokId); err != nil {
		return
	}
	_, err = p.expect(tokClose)
	return
}

func parseObjsDecl(p *parser) (objs []TypedName, err error) {
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
	switch {
	case p.acceptNamedList("="):
		nd := Node{p.Loc()}
		asn := &AssignNode{ Op:   Name{"=", nd.Loc() }, Node: nd }
		if asn.Lval, err = parseFhead(p); err != nil {
			return
		}
		var t []token
		if t, err = p.expect(tokNum); err != nil {
			return
		}
		asn.Rval = t[0].text
		form = asn
		_, err = p.expect(tokClose)
	case p.acceptNamedList("not"):
		nd := Node{p.Loc()}
		var prop *PropositionNode
		if prop, err = parseProposition(p); err != nil {
			return
		}
		form = &NotNode{UnaryNode{nd, prop}}
		_, err = p.expect(tokClose)
	default:
		form, err = parseProposition(p)
	}
	return
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

func parseStringPlus(p *parser, typ tokenType) (ns []Name, err error) {
	var n Name
	if n, err = parseName(p, typ); err != nil {
		return
	}
	ns = []Name{ n }
	ns = append(ns, parseStrings(p, typ)...)
	return
}

func parseStrings(p *parser, typ tokenType) (lst []Name) {
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		l := p.Loc()
		lst = append(lst, Name{ t.text, l })
	}
	return lst
}

func parseName(p *parser, typ tokenType) (Name, error) {
	l := p.Loc()
	n, err := p.expect(typ)
	if err != nil {
		return Name{}, err
	}
	return Name{ n[0].text, l }, nil
}