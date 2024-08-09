package ft

type TyTag int

var tVarCount int

const (
	_ TyTag = iota
	Any
	Never
	Int
	Tuple
	Not
	And
	Or
	// Infer: any that's < 0
)

type Ty struct {
	Tag TyTag // if < 0, is inference variable
	Of  []*Ty // for Tuple,Not,And,Or
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
	Params []*Param
	Body   Term
	In     Term
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

func shouldWrapTy(tyChild, tyParent *Ty) bool {
	prec := func(ty *Ty) int {
		switch ty.Tag {
		case Or:
			return 1
		case And:
			return 2
		case Not:
			return 3
		}
		panic("unreachable")
	}
	if tyChild.Tag == Any || tyChild.Tag == Never || tyChild.Tag == Int || tyChild.Tag == Tuple || tyChild.Tag < 0 ||
		tyParent.Tag == Any || tyParent.Tag == Never || tyParent.Tag == Int || tyParent.Tag == Tuple || tyParent.Tag < 0 {
		return false
	}
	return prec(tyChild) < prec(tyParent)
}
