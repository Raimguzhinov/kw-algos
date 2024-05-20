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

	t, err := simplex.Scan(r)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", t.ToCanonicalForm())
	fmt.Println("Jordaan Gauss:")
	basis, err := t.ToBasis()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", t)

	fmt.Println("Dual Simplex method:")
	simplexTable := simplex.New(basis)
	result, err := simplexTable.DualMethod()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
