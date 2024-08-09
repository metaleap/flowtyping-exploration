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

func dnfOfTy(ty *Ty) {
	dnf_inter := func(ty *Ty) {

	}
	_ = dnf_inter // TODO: cont. here
}
