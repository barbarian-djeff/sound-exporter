package main

const max = 5

func addPeak(arr []Peak, p Peak) []Peak {
	res := []Peak{p}
	if len(arr) == max {
		return append(res, arr[:max-1]...)
	}
	return append(res, arr...)
}

func addMinute(arr []Minute, m Minute) []Minute {
	res := []Minute{m}
	if len(arr) == max {
		return append(res, arr[:max-1]...)
	}
	return append(res, arr...)
}
