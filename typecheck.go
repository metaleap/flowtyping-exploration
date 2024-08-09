package ft

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
		return t
	}

	unwrap_trivial = func(t *Ty) *Ty {
		return t
	}

	return unwrap_trivial(flatten_ors(flatten_ands(ty)))
}
