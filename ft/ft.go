package ft

func flattenTyInTerm(t Term) Term {
	switch val := t.Val.(type) {
	case termValNum, termValVar:
		return t
	case termValTup:
		return Term{Ty: t.Ty, Val: termValTup(listMap(val, flattenTyInTerm))}
	case termValApp:
		return Term{Ty: t.Ty, Val: termValApp{Callee: val.Callee, Args: listMap(val.Args, flattenTyInTerm)}}
	case termValDec:
		return Term{Ty: t.Ty, Val: termValDec{
			Name: val.Name,
			Params: listMap(val.Params, func(it Param) Param {
				if !needsInference(it.Ty) {
					it.Ty = flattenTy(it.Ty)
				}
				return it
			}),
			Body: flattenTyInTerm(val.Body),
			Cont: flattenTyInTerm(val.Cont),
		}}
	case termValIf:
		return Term{Ty: t.Ty, Val: termValIf{Var: val.Var, IsTy: flattenTy(val.IsTy), Then: flattenTyInTerm(val.Then), Else: flattenTyInTerm(val.Else)}}
	}
	panic("unreachable")
}
