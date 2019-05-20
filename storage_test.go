package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_add(t *testing.T) {
	minutes = []Minute{}
	minutes = addMinute(minutes, m("1"))
	equal(t, []string{"1"}, minutes)

	minutes = addMinute(minutes, m("2"))
	minutes = addMinute(minutes, m("3"))
	minutes = addMinute(minutes, m("4"))
	equal(t, []string{"4", "3", "2", "1"}, minutes)

	minutes = addMinute(minutes, m("5"))
	equal(t, []string{"5", "4", "3", "2", "1"}, minutes)

	minutes = addMinute(minutes, m("6"))
	minutes = addMinute(minutes, m("7"))
	equal(t, []string{"7", "6", "5", "4", "3"}, minutes)
}

func equal(t *testing.T, expected []string, actual []Minute) {
	require.Len(t, actual, len(expected))
	for i, v := range expected {
		assert.Equal(t, v, string(actual[i].Time))
	}
}

func m(t string) Minute {
	return Minute{
		Time: t,
	}
}
