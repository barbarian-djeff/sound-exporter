package main

const max = 5

func add(arr []Minute, t Minute) []Minute {
	res := []Minute{t}
	if len(arr) == max {
		return append(res, arr[:4]...)
	}
	return append(res, arr...)
}
