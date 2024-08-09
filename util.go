package ft

func listMap[TIn any, TOut any](list []TIn, f func(TIn) TOut) (ret []TOut) {
	ret = make([]TOut, len(list))
	for i, item := range list {
		ret[i] = f(item)
	}
	return
}
