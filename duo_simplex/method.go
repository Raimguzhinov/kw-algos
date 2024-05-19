package duo_simplex

import (
	"fmt"
	"kw-algos/fractional"
)

func SimplexMethod(t *Table) (*Table, error) {
	for i := range t.Z {
		reverse, _ := fractional.New(-1, 1)
		t.Z[i] = t.Z[i].Multiply(*reverse)
	}

	maxValue := fractional.ZeroValue
	var resolveRow int
	isOptimal := false
	for i := range t.Rows {
		if t.Matrix[i][t.Cols-1 : t.Cols][0].LessThan(
			*fractional.ZeroValue,
		) && t.Matrix[i][t.Cols-1 : t.Cols][0].LessThan(
			*maxValue) {
			maxValue = t.Matrix[i][t.Cols-1 : t.Cols][0]
			resolveRow = i
		} else {
			isOptimal = true
		}
	}
	if isOptimal {
		panic("TODO!!!")
	}

	CO := make([]*fractional.Fraction, t.Cols-1)
	isNegative := false
	var err error
	for j := range t.Cols - 1 {
		if t.Matrix[resolveRow][j].LessThan(*fractional.ZeroValue) {
			isNegative = true
			CO[j], err = t.Z[j].Divide(*t.Matrix[resolveRow][j])
			if err != nil {
				return nil, err
			}
		}
	}
	if !isNegative {
		panic("NO SOLUTION!!!")
	}

	for _, co := range CO {

	}

	fmt.Println(maxValue)
	return t, nil
}
