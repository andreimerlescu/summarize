package main

import (
	"github.com/andreimerlescu/figtree/v2"
	"os"
	"strconv"
)

// envVal takes a name for os.LookupEnv with a fallback to return a string
func envVal(name, fallback string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	return v
}

// envIs takes a name for os.LookupEnv with a fallback of false to return a bool
func envIs(name string) bool {
	v, ok := os.LookupEnv(name)
	if !ok {
		return false
	}
	vb, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return vb
}

// envInt takes a name for os.Lookup with a fallback value to return an int
func envInt(name string, fallback int) int {
	v, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

// addFromEnv takes a pointer to a slice of strings and a new ENV os.LookupEnv name to return the figtree ToList on the Flesh that sends the list into simplify before being returned
func addFromEnv(e string, l *[]string) {
	v, ok := os.LookupEnv(e)
	if ok {
		flesh := figtree.NewFlesh(v)
		maybeAdd := flesh.ToList()
		for _, entry := range maybeAdd {
			*l = append(*l, entry)
		}
	}
	*l = simplify(*l)
}
