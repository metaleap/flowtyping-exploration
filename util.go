package ft

var tyNever = &Ty{Tag: TyNever}

func assert(b bool) {
	if !b {
		panic("assertion failure")
	}
}

func If[T any](b bool, t T, f T) T {
	if b {
		return t
	}
	return f
}

func listForAll2[T any](list1 []T, list2 []T, f func(T, T) bool) bool {
	if len(list1) == len(list2) {
		for i, item := range list1 {
			if !f(item, list2[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func listExists2[T any](list1 []T, list2 []T, f func(T, T) bool) bool {
	if len(list1) == len(list2) {
		for i, item := range list1 {
			if f(item, list2[i]) {
				return true
			}
		}
	}
	return false
}

func listFind[T any](list []T, pred func(T) bool) (ret T) {
	for _, item := range list {
		if pred(item) {
			return item
		}
	}
	return
}

func listMap[TIn any, TOut any](list []TIn, f func(TIn) TOut) (ret []TOut) {
	ret = make([]TOut, len(list))
	for i, item := range list {
		ret[i] = f(item)
	}
	return
}

// caller assures same len for both lists
func listMap2[TIn any, TOut any](list1 []TIn, list2 []TIn, f func(TIn, TIn) TOut) (ret []TOut) {
	ret = make([]TOut, len(list1))
	for i, item := range list1 {
		ret[i] = f(item, list2[i])
	}
	return
}

func listFilter[T any](list []T, pred func(T) bool) []T {
	ret := make([]T, 0, len(list))
	for _, item := range list {
		if pred(item) {
			ret = append(ret, item)
		}
	}
	return ret
}

func listFold[TListItem any, TAccum any](list []TListItem, accum TAccum, f func(TListItem, TAccum) TAccum) TAccum {
	for _, t := range list {
		accum = f(t, accum)
	}
	return accum
}

func listRev[T any](list []T) []T {
	ret := make([]T, len(list))
	for i, item := range list {
		ret[len(ret)-(1+i)] = item
	}
	return ret
}

type Tys []*Ty

func (me Tys) setAdd(ty *Ty) Tys {
	return me.setUnion(ty)
}

func (me Tys) setChoose() *Ty { return me[0] }

func (me Tys) setExists(pred func(*Ty) bool) bool {
	return (me.setFindFirst(pred) != nil)
}

func (me Tys) setFindFirst(pred func(*Ty) bool) *Ty {
	for _, t := range me {
		if pred(t) {
			return t
		}
	}
	return nil
}

func (me Tys) setRemove(ty *Ty) Tys {
	for i, t := range me {
		if t.eq(ty) {
			return append(me[:i], me[i+1:]...)
		}
	}
	return me
}

func (me Tys) setUnion(with ...*Ty) Tys {
	for _, t2 := range with {
		var exists bool
		for _, t1 := range me {
			if exists = t1.eq(t2); exists {
				break
			}
		}
		if !exists {
			me = append(me, t2)
		}
	}
	return me
}

func atomEq(a1 Atom, a2 Atom) bool {
	p1, isp1 := a1.(AtomPos)
	p2, isp2 := a2.(AtomPos)
	if isp1 && isp2 && p1.Kind == p2.Kind && listForAll2(p1.IfTuple, p2.IfTuple, func(tp1 AtomPos, tp2 AtomPos) bool { return atomEq(tp1, tp2) }) {
		return true
	}
	n1, isn1 := a1.(AtomNeg)
	n2, isn2 := a2.(AtomNeg)
	if isn1 && isn2 {
		return atomEq(n1.Not, n2.Not)
	}
	s1, iss1 := a1.(AtomStar)
	s2, iss2 := a2.(AtomStar)
	if iss1 && iss2 {
		return (s1.Neg != nil && s2.Neg != nil && atomEq(s1.Neg, s2.Neg)) ||
			(s1.Pos != nil && s2.Pos != nil && atomEq(s1.Pos, s2.Pos))
	}
	panic("unreachable")
}

func (me *Ty) eq(to *Ty) bool {
	if me == to {
		return true
	} else if me.Tag == to.Tag && len(me.Of) == len(to.Of) {
		for i, t := range me.Of {
			if !t.eq(to.Of[i]) {
				return false
			}
		}
		return true
	}
	return false
}
