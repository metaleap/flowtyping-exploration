package ft

type TyTag int

var tVarCount int

const (
	_ TyTag = iota
	TyAny
	TyNever
	TyInt
	TyTuple
	TyNot
	TyAnd
	TyOr
	// Infer: any that's < 0
)

type Ty struct {
	Tag TyTag // if < 0, is inference variable
	Of  Tys   // for Tuple,Not,And,Or
}

type Param struct {
	Name string
	Ty   *Ty
}

type Term struct {
	Ty  *Ty
	Val TermVal
}

type TermVal interface{ isTermVal() }

func (termValNum) isTermVal() {}
func (termValVar) isTermVal() {}
func (termValTup) isTermVal() {}
func (termValApp) isTermVal() {}
func (termValDec) isTermVal() {}
func (termValIf) isTermVal()  {}

type termValNum int
type termValVar string
type termValTup []Term
type termValApp struct {
	Callee string
	Args   []Term
}
type termValDec struct {
	Name   string
	Params []Param
	Body   Term
	Cont   Term
}
type termValIf struct {
	Var  string
	IsTy *Ty
	Then Term
	Else Term
}

func needsInference(ty *Ty) bool {
	return ty.Tag < 0
}

func freshInferTy() *Ty {
	tVarCount--
	return &Ty{Tag: TyTag(tVarCount)}
}

var prec = map[TyTag]int{
	TyOr:  1,
	TyAnd: 2,
	TyNot: 3,
}

func shouldWrapTy(tyChild, tyParent *Ty) bool {
	if tyChild.Tag == TyAny || tyChild.Tag == TyNever || tyChild.Tag == TyInt || tyChild.Tag == TyTuple || tyChild.Tag < 0 ||
		tyParent.Tag == TyAny || tyParent.Tag == TyNever || tyParent.Tag == TyInt || tyParent.Tag == TyTuple || tyParent.Tag < 0 {
		return false
	}
	return prec[tyChild.Tag] < prec[tyParent.Tag]
}
