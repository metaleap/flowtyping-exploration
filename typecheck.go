package ft

import (
	"fmt"
)

func flattenTy(ty *Ty) *Ty {
	var flatten_ands, flatten_ors, unwrap_trivial func(*Ty) *Ty

	flatten_ands = func(t *Ty) *Ty {
		switch t.Tag {
		case TyTuple, TyNot, TyOr:
			return &Ty{Tag: t.Tag, Of: listMap(t.Of, flatten_ands)}
		case TyAnd:
			return &Ty{Tag: TyAnd, Of: listFold(listMap(t.Of, flatten_ands), Tys{}, func(it *Ty, accum Tys) Tys {
				if it.Tag == TyAnd {
					return accum.setUnion(it.Of...)
				}
				return accum.setAdd(it)
			})}
		}
		return t
	}

	flatten_ors = func(t *Ty) *Ty {
		switch t.Tag {
		case TyTuple, TyNot, TyAnd:
			return &Ty{Tag: t.Tag, Of: listMap(t.Of, flatten_ors)}
		case TyOr:
			return &Ty{Tag: TyOr, Of: listFold(listMap(t.Of, flatten_ors), Tys{}, func(it *Ty, accum Tys) Tys {
				if it.Tag == TyOr {
					return accum.setUnion(it.Of...)
				}
				return accum.setAdd(it)
			})}
		}
		return t
	}

	unwrap_trivial = func(t *Ty) *Ty {
		switch t.Tag {
		case TyTuple, TyNot:
			return &Ty{Tag: t.Tag, Of: listMap(t.Of, unwrap_trivial)}
		case TyAnd:
			tys := listMap(t.Of, unwrap_trivial)
			return If(len(tys) == 1, tys[0], &Ty{Tag: TyAnd, Of: tys})
		case TyOr:
			if tys := Tys(listMap(t.Of, unwrap_trivial)).setRemove(tyNever); len(tys) == 0 {
				return tyNever
			} else if len(tys) == 1 {
				return tys[0]
			} else {
				return &Ty{Tag: TyOr, Of: tys}
			}
		}
		return t
	}

	return unwrap_trivial(flatten_ors(flatten_ands(ty)))
}

type Atom interface{ isTypeAtom() }

func (AtomPos) isTypeAtom()  {}
func (AtomNeg) isTypeAtom()  {}
func (AtomStar) isTypeAtom() {}

type AtomPos struct {
	Kind    TyTag
	IfTuple []AtomPos
}

type AtomNeg struct {
	Not AtomPos
}

type AtomStar struct {
	Pos *AtomPos
	Neg *AtomNeg
}

type AtomInter struct {
	Pos *AtomPos // if nil, that AtomInter means "Never"
}

func atomOfTy(ty *Ty) Atom {
	var pos func(*Ty) AtomPos
	pos = func(t *Ty) AtomPos {
		switch t.Tag {
		case TyTuple:
			return AtomPos{Kind: TyTuple, IfTuple: listMap(t.Of, pos)}
		case TyInt, TyAny:
			return AtomPos{Kind: t.Tag}
		}
		panic("notPositive:" + fmt.Sprintf("%#v", t))
	}
	switch ty.Tag {
	case TyNever:
		return AtomNeg{Not: AtomPos{Kind: TyAny}}
	case TyNot:
		if ty.Tag == TyNever {
			return AtomPos{Kind: TyAny}
		} else {
			return AtomNeg{Not: pos(ty.Of[0])}
		}
	}
	return pos(ty)
}

func tyOfPos(atom AtomPos) *Ty {
	switch atom.Kind {
	case TyAny:
		return &Ty{Tag: TyAny}
	case TyInt:
		return &Ty{Tag: TyInt}
	case TyTuple:
		return &Ty{Tag: TyTuple, Of: listMap(atom.IfTuple, tyOfPos)}
	}
	panic("unreachable")
}

func tyOfNeg(atom AtomNeg) *Ty {
	return &Ty{Tag: TyNot, Of: Tys{tyOfPos(atom.Not)}}
}

func tyOfAtom(atom Atom) *Ty {
	switch atom := atom.(type) {
	case AtomNeg:
		return tyOfNeg(atom)
	case AtomPos:
		return tyOfPos(atom)
	}
	panic("unreachable")
}

func atomSub(s Atom, t Atom) bool {
	if atomEq(s, t) {
		return true
	} else if t_pos, t_is := t.(AtomPos); t_is && t_pos.Kind == TyAny {
		return true
	} else if s_pos, s_is := s.(AtomPos); s_is && t_is && s_pos.Kind == TyTuple && t_pos.Kind == TyTuple {
		return listForAll2(s_pos.IfTuple, t_pos.IfTuple, func(a1 AtomPos, a2 AtomPos) bool {
			return atomSub(a1, a2)
		})
	}
	return false
}
func atomSup(s Atom, t Atom) bool    { return atomSub(t, s) }
func atomSupNot(s Atom, t Atom) bool { return !atomSup(s, t) }

func atomInter(t1 AtomPos, t2 AtomPos) AtomInter {
	switch {
	case atomEq(t1, t2):
		return AtomInter{Pos: &t1}
	case t1.Kind == TyAny:
		return AtomInter{Pos: &t2}
	case t2.Kind == TyAny:
		return AtomInter{Pos: &t1}
	case (t1.Kind == TyInt && t2.Kind == TyTuple) || (t1.Kind == TyTuple && t2.Kind == TyInt):
		return AtomInter{Pos: nil} // never
	case t1.Kind == TyTuple && t2.Kind == TyTuple:
		if len(t1.IfTuple) != len(t2.IfTuple) || listExists2(t1.IfTuple, t2.IfTuple, func(ti AtomPos, si AtomPos) bool {
			return atomInter(ti, si).Pos == nil // == never
		}) {
			return AtomInter{Pos: nil} // never
		} else {
			return AtomInter{Pos: &AtomPos{
				Kind: TyTuple,
				IfTuple: listMap2(t1.IfTuple, t2.IfTuple, func(ti AtomPos, si AtomPos) AtomPos {
					inter := atomInter(ti, si)
					if inter.Pos == nil {
						panic("impossible")
					}
					return *inter.Pos
				}),
			}}
		}
	}
	panic("unreachable")
}

func splitAt(sep *Ty, list Tys) (Tys, Tys) {
	var walk func(Tys, Tys) (Tys, Tys)
	walk = func(before Tys, list Tys) (Tys, Tys) {
		if len(list) == 0 {
			panic("splitting element not found")
		}
		if list[0].eq(sep) {
			return listRev(before), list[1:]
		}
		return walk(append(Tys{list[0]}, before...), list[1:])
	}
	return walk(nil, list)
}

func atomPos(it Atom) *AtomPos {
	ret, is := it.(AtomPos)
	return If(is, &ret, nil)
}
func atomNeg(it Atom) *AtomNeg {
	ret, is := it.(AtomNeg)
	return If(is, &ret, nil)
}

type DnfInter = []AtomStar
type DnfForm = []DnfInter

func dnfOfTy(ty *Ty) DnfForm {
	dnf_inter := func(ty *Ty) DnfInter {
		switch ty.Tag {
		case TyAnd:
			atoms := listMap(ty.Of, atomOfTy)
			ret := make(DnfInter, len(atoms))
			for i, atom := range atoms {
				ret[i] = AtomStar{Pos: atomPos(atom), Neg: atomNeg(atom)}
			}
		case TyOr:
			panic("not_dnf")
		case TyNot, TyTuple, TyAny, TyNever, TyInt:
			atom := atomOfTy(ty)
			return DnfInter{{Pos: atomPos(atom), Neg: atomNeg(atom)}}
		}
		panic("impossible")
	}
	var dnf_form func(ty *Ty) DnfForm
	dnf_form = func(ty *Ty) DnfForm {
		switch ty.Tag {
		case TyOr:
			return listMap(ty.Of, dnf_inter)
		case TyAnd:
			return dnf_form(&Ty{Tag: TyOr, Of: Tys{ty}})
		case TyNot, TyTuple, TyAny, TyNever, TyInt:
			return dnf_form(&Ty{Tag: TyAnd, Of: Tys{ty}})
		}
		panic("impossible")
	}
	return dnf_form(ty)
}

func tyOfDnf(inters DnfForm) *Ty {
	tys := listMap(inters, func(atoms DnfInter) *Ty {
		return &Ty{Tag: TyAnd, Of: listMap(atoms, func(it AtomStar) *Ty { return tyOfAtom(it) })}
	})
	return flattenTy(&Ty{Tag: TyOr, Of: tys})
}

func dnfStep(ty *Ty) *Ty {
	switch {
	case ty.Tag == TyNot && ty.Of[0].Tag == TyNot:
		return ty.Of[0].Of[0]
	case ty.Tag == TyNot && ty.Of[0].Tag == TyOr:
		return &Ty{Tag: TyAnd, Of: listMap(ty.Of[0].Of, func(t *Ty) *Ty { return &Ty{Tag: TyNot, Of: Tys{t}} })}
	case ty.Tag == TyNot && ty.Of[0].Tag == TyAnd:
		return &Ty{Tag: TyOr, Of: listMap(ty.Of[0].Of, func(t *Ty) *Ty { return &Ty{Tag: TyNot, Of: Tys{t}} })}
	case ty.Tag == TyAnd && ty.Of.setExists(func(t *Ty) bool { return t.Tag == TyOr }):
		// Factor unions out of intersections
		factor_out := ty.Of.setFindFirst(func(t *Ty) bool { return t.Tag == TyOr })
		rest_inters := ty.Of.setRemove(factor_out)
		return &Ty{Tag: TyOr, Of: listMap(factor_out.Of, func(s_i *Ty) *Ty { return &Ty{Tag: TyOr, Of: rest_inters.setAdd(s_i)} })}
	case ty.Tag == TyTuple && ty.Of.setExists(func(t *Ty) bool { return t.Tag == TyOr }):
		// Factor unions out of tuples
		factor_out := listFind(ty.Of, func(t *Ty) bool { return t.Tag == TyOr })
		before, after := splitAt(factor_out, ty.Of)
		return &Ty{Tag: TyOr, Of: listMap(factor_out.Of, func(t_i *Ty) *Ty { return &Ty{Tag: TyTuple, Of: append(before, append(Tys{t_i}, after...)...)} })}
	case ty.Tag == TyTuple && ty.Of.setExists(func(t *Ty) bool { return t.Tag == TyAnd }):
		// Factor intersections out of tuples
		factor_out := listFind(ty.Of, func(t *Ty) bool { return t.Tag == TyAnd })
		before, after := splitAt(factor_out, ty.Of)
		return &Ty{Tag: TyAnd, Of: listMap(factor_out.Of, func(t_i *Ty) *Ty { return &Ty{Tag: TyTuple, Of: append(before, append(Tys{t_i}, after...)...)} })}
	case ty.Tag == TyTuple && ty.Of.setExists(func(t *Ty) bool { return t.Tag == TyNot }):
		// Factor negations out of tuples
		factor_out := listFind(ty.Of, func(t *Ty) bool { return t.Tag == TyNot })
		before, after := splitAt(factor_out, ty.Of)
		t := factor_out.Of[0]
		return &Ty{Tag: TyAnd, Of: Tys{
			&Ty{Tag: TyTuple, Of: append(before, append(Tys{{Tag: TyAny}}, after...)...)},
			&Ty{Tag: TyNot, Of: Tys{{Tag: TyTuple, Of: append(before, append(Tys{t}, after...)...)}}},
		}}
	}
	return ty
}
