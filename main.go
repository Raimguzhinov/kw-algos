package main

import (
	"fmt"
	"io"
	"kw-algos/matrix"
	"os"
)

func main() {
	//path := os.Args[1]
	//_ = path
	f, err := os.Open("1.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var r io.Reader
	r = f
	m, err := matrix.New(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", m.ToCanonicalForm())
	basis, err := m.ToBasis()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", basis)
}
