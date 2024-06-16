package funcutil

import "errors"

var (
	NotFound = errors.New("not found")
)

func IsNotFound(err error) bool {
	return errors.Is(err, NotFound)
}

func Map[S ~[]E, E any, ER any](slice S, mapFunc func(E) ER) []ER {
	res := make([]ER, 0, len(slice))
	for _, elem := range slice {
		res = append(res, mapFunc(elem))
	}
	return res
}

func Filter[S ~[]E, E any](slice S, filterFunc func(E) bool) []E {
	res := make([]E, 0)
	for _, elem := range slice {
		if filterFunc(elem) {
			res = append(res, elem)
		}
	}
	return res
}

func Split[S ~[]E, E any](slice S, filterFunc func(E) bool) ([]E, []E) {
	resLeft, resRight := make([]E, 0), make([]E, 0)
	for _, elem := range slice {
		if filterFunc(elem) {
			resLeft = append(resLeft, elem)
		} else {
			resRight = append(resRight, elem)
		}
	}
	return resLeft, resRight
}

func FindFirst[S ~[]E, E any](slice S, filterFunc func(E) bool) (E, error) {
	for _, elem := range slice {
		if filterFunc(elem) {
			return elem, nil
		}
	}
	var empty E
	return empty, NotFound
}

func MaxFunc[S ~[]E, E any](slice S, cmpFunc func(a, b E) int) E {
	if len(slice) < 1 {
		panic("MaxFunc: empty list")
	}
	maxElem := slice[0]
	for i := 1; i < len(slice); i++ {
		if cmpFunc(slice[i], maxElem) > 0 {
			maxElem = slice[i]
		}
	}
	return maxElem
}
