package main

import (
	"fmt"
	"io"
	"kw-algos/simplex"
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

	m, err := simplex.Scan(r)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", m.ToCanonicalForm())
	fmt.Println("Jordan Gauss:")
	table, err := m.ToBasis()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", m)

	fmt.Println("Dual Simplex method:")
	simplexTable := simplex.New(table)

	if err := simplexTable.DualMethod(); err != nil {
		panic(err)
	}
}
