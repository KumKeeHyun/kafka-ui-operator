package funcutil

func GroupBy[S ~[]E, E any, K comparable](slice S, keyFunc func(e E) K) map[K]E {
	res := make(map[K]E, len(slice))
	for _, elem := range slice {
		res[keyFunc(elem)] = elem
	}
	return res
}
