package ft

func listMap[TIn any, TOut any](list []TIn, f func(TIn) TOut) (ret []TOut) {
	ret = make([]TOut, len(list))
	for i, item := range list {
		ret[i] = f(item)
	}
	return
}

func listFold[TListItem any, TAccum any](list []TListItem, accum TAccum, f func(TListItem, TAccum) TAccum) TAccum {
	for _, t := range list {
		accum = f(t, accum)
	}
	return accum
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
