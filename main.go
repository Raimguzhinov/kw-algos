package main

import (
	"flag"
	"fmt"
	"io"
	"kw-algos/simplex"
	"os"
)

func main() {
	var path string
	flag.StringVar(&path, "p", "test.txt", "(path to file) <filename>.txt")
	flag.Parse()

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
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
		fmt.Println(err)
	}
}
