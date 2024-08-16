package main

type Term interface{ isTerm() }

func (*TermLit) isTerm() {}
func (*TermVar) isTerm() {}
func (*TermLam) isTerm() {}
func (*TermApp) isTerm() {}
func (*TermRec) isTerm() {}
func (*TermSel) isTerm() {}
func (*TermLet) isTerm() {}

type TermLit struct{ value any }
type TermVar struct{ name string }
type TermLam struct {
	name string
	rhs  Term
}
type TermApp struct {
	lhs Term
	rhs Term
}
type TermRec struct {
	fields map[string]Term
}
type TermSel struct {
	receiver  Term
	fieldName string
}
type TermLet struct {
	isRec bool
	name  string
	rhs   Term
	body  Term
}

type Type interface{ isType() }

func (*TypeTop) isType()       {}
func (*TypeBot) isType()       {}
func (*TypeUnion) isType()     {}
func (*TypeInter) isType()     {}
func (*TypeFunction) isType()  {}
func (*TypeRecord) isType()    {}
func (*TypeRecursive) isType() {}
func (*TypePrimitive) isType() {}
func (*TypeVariable) isType()  {}

type TypeTop struct{}
type TypeBot struct{}
type TypeUnion struct {
	lhs Type
	rhs Type
}
type TypeInter struct {
	lhs Type
	rhs Type
}
type TypeFunction struct {
	lhs Type
	rhs Type
}
type TypeRecord struct {
	fields map[string]Type
}
type TypeRecursive struct {
	uv   Type
	body Type
}
type TypePrimitive struct {
	name string
}
type TypeVariable struct {
	nameHint string
	hash     int
}
