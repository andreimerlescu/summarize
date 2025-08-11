package main

import (
	"github.com/andreimerlescu/goenv/env"
)

// addFromEnv takes a pointer to a slice of strings and a new ENV os.LookupEnv name to return the figtree ToList on the Flesh that sends the list into simplify before being returned
func addFromEnv(e string, l *[]string) {
	for _, entry := range env.List(e, []string{}) {
		*l = append(*l, entry)
	}
	*l = simplify(*l)
}
