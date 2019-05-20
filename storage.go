package main

const max = 5

func add(arr []Minute, t Minute) []Minute {
	if len(arr) == max {
		res := make([]Minute, 5)
		res[0] = t
		return append(res, arr[:4]...)
	}
	return append(arr, t)
}
