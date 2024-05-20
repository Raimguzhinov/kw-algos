package simplex

import (
	"bufio"
	"fmt"
	"io"
	"kw-algos/fractional"
	"math"
	"strconv"
	"strings"
)

type Comparison int

const (
	EqualTo Comparison = iota
	LessThanOrEqualTo
	GreaterThanOrEqualTo
)

func (c *Comparison) String() string {
	return [...]string{"=", "<=", ">="}[*c]
}

type Table struct {
	Rows, Cols, Vars      int
	Matrix                [][]*fractional.Fraction
	Z                     []*fractional.Fraction
	IsMinimizationProblem bool
	BasisVars             []int
	ZFree                 *fractional.Fraction
	comparisons           []Comparison
}

func Scan(r io.Reader) (*Table, error) {
	reader := bufio.NewReader(r)
	var rows, cols, vars int
	comparisons := make([]Comparison, 0)

	line, _ := reader.ReadString('\n')
	parts := strings.Fields(line)
	rows, _ = strconv.Atoi(parts[0])
	cols, _ = strconv.Atoi(parts[1])

	vars = cols
	cols++

	matrix := make([][]*fractional.Fraction, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]*fractional.Fraction, cols, cols*2)
		line, _ := reader.ReadString('\n')
		parts := strings.Fields(line)
		for j := 0; j < cols; j++ {
			if j == vars {
				// Считываем математический знак и результат для текущего уравнения
				sign := parts[j]
				comparison, err := parseComparison(sign)
				if err != nil {
					return nil, err
				}
				comparisons = append(comparisons, comparison)
				result, err := strconv.ParseInt(parts[j+1], 10, 64)
				if err != nil {
					return nil, err
				}
				matrix[i][cols-1], err = fractional.New(result, 1)
				if err != nil {
					return nil, err
				}
				break
			}
			// Запись свободных переменных
			value, err := strconv.ParseInt(parts[j], 10, 64)
			if err != nil {
				return nil, err
			}
			matrix[i][j], err = fractional.New(value, 1)
			if err != nil {
				return nil, err
			}
		}

	}
	// Чтение строки для максимизации Z
	line, _ = reader.ReadString('\n')
	parts = strings.Fields(line)
	Z := make([]*fractional.Fraction, vars)
	for j := 0; j < vars; j++ {
		value, err := strconv.ParseInt(parts[j], 10, 64)
		if err != nil {
			return nil, err
		}
		Z[j], err = fractional.New(value, 1)
		if err != nil {
			return nil, err
		}
	}
	z, err := strconv.ParseInt(parts[vars], 10, 64)
	if err != nil {
		return nil, err
	}
	ZFree, err := fractional.New(z, 1)
	if err != nil {
		return nil, err
	}

	maxMinSign := parts[vars+1]
	var isMinimization bool
	if maxMinSign == "min" {
		isMinimization = true
	} else {
		isMinimization = false
	}

	return &Table{
		rows,
		cols,
		vars,
		matrix,
		Z,
		isMinimization,
		make([]int, rows),
		ZFree,
		comparisons,
	}, nil
}

func parseComparison(sign string) (Comparison, error) {
	switch sign {
	case "<=":
		return LessThanOrEqualTo, nil
	case ">=":
		return GreaterThanOrEqualTo, nil
	case "=":
		return EqualTo, nil
	default:
		return -1, fmt.Errorf("invalid comparison: %s", sign)
	}
}

func (t *Table) String() string {
	var s string
	for i := 0; i < t.Rows; i++ {
		for j := 0; j < t.Cols; j++ {
			s += fmt.Sprintf("%*s ", 8, t.Matrix[i][j])
		}
		s += "\n"
	}
	return s
}

func (t *Table) ToCanonicalForm() *Table {
	var newColsCnt, beforeNormalization int
	basis := make(map[int][]*fractional.Fraction)
	var lasts []*fractional.Fraction

	for i, comparison := range t.comparisons {
		var b *fractional.Fraction

		switch comparison {
		case LessThanOrEqualTo:
			b = fractional.OneValue
			t.Z = append(t.Z, fractional.ZeroValue)
		case GreaterThanOrEqualTo:
			b = fractional.RevOneValue
			t.Z = append(t.Z, fractional.ZeroValue)
		default:
			beforeNormalization++
		}
		newColsCnt++
		last := t.Matrix[i][t.Vars]
		lasts = append(lasts, last)
		basis[i] = make([]*fractional.Fraction, t.Rows)
		for j := 0; j < t.Rows; j++ {
			basis[i][j] = fractional.ZeroValue
		}
		basis[i][i] = b
	}
	if newColsCnt > 0 {
		t.Cols += newColsCnt
		for i := 0; i < t.Rows; i++ {
			head := t.Matrix[i][:t.Vars]
			tail := append(t.Matrix[i][t.Vars+1:], basis[i]...)
			tail = append(tail, lasts[i])
			t.Matrix[i] = append(head, tail...)
			t.comparisons[i] = EqualTo
		}
	}
	t.Matrix = normalizeMatrix(t.Matrix)
	t.Cols -= beforeNormalization
	if t.IsMinimizationProblem {
		for i := 0; i < t.Cols-1; i++ {
			t.Z[i] = t.Z[i].Reverse()
		}
	}
	return t
}

func normalizeMatrix(matrix [][]*fractional.Fraction) [][]*fractional.Fraction {
	var result [][]*fractional.Fraction
	if len(matrix) == 0 {
		return result
	}
	nilColumns := make(map[int]bool)

	for _, row := range matrix {
		for i, value := range row {
			if value == nil {
				nilColumns[i] = true
			}
		}
	}
	for _, row := range matrix {
		var newRow []*fractional.Fraction
		for i, value := range row {
			if _, exists := nilColumns[i]; !exists {
				newRow = append(newRow, value)
			}
		}
		result = append(result, newRow)
	}
	return result
}

func (t *Table) ToBasis() (*Table, error) {
	var columOfResolver int
	for i := 0; i < t.Rows; i++ {
		fmt.Println(t)

		t.swapMatrixRows(i, columOfResolver)
		needToCheck := false
		newMatrix := t.CopyMatrix()

		currentColumOfResolver := columOfResolver
		if t.Matrix[i][currentColumOfResolver].Equal(*fractional.ZeroValue) {
			needToCheck = true
			for j := currentColumOfResolver + 1; j < t.Cols-1; j++ {
				t.swapMatrixRows(i, j)

				if !t.Matrix[i][j].Equal(*fractional.ZeroValue) {
					newMatrix = t.Matrix
					needToCheck = false
					columOfResolver = j + 1
					currentColumOfResolver = j

					break
				}
			}
			if needToCheck && t.Matrix[i][t.Cols-1].Equal(*fractional.ZeroValue) {
				rank, err := t.checkRank()
				if err != nil {
					return nil, err
				}
				if i == t.Rows-1 && rank == 1 {
					return t, nil
				}
				continue
			} else if currentColumOfResolver == columOfResolver {
				_, err := t.checkRank()
				if err != nil {
					return nil, err
				}
				return t, err
			}
		} else {
			columOfResolver++
		}

		if t.Matrix[i][currentColumOfResolver].NotEqual(*fractional.RevOneValue) {
			divider := t.Matrix[i][currentColumOfResolver]
			for j := 0; j < t.Cols; j++ {
				var err error
				newMatrix[i][j], err = t.Matrix[i][j].Divide(*divider)
				if err != nil {
					return nil, err
				}
			}
		}

		t.BasisVars[i] = currentColumOfResolver
		if err := t.methodRectangle(newMatrix, i, currentColumOfResolver); err != nil {
			return nil, err
		}
		t.Matrix = newMatrix

		if i == t.Rows-1 || needToCheck {
			_, err := t.checkRank()
			if err != nil {
				return nil, err
			}
			return t, err
		}
	}
	return t, nil
}

func (t *Table) swapMatrixRows(startRow, startColumn int) {
	maxValue := math.Abs(t.Matrix[startRow][startColumn].Float64())
	maxIndex := startRow
	for j := startRow; j < t.Rows; j++ {
		if maxValue < math.Abs(t.Matrix[j][startColumn].Float64()) {
			maxValue = math.Abs(t.Matrix[j][startColumn].Float64())
			maxIndex = j
		}
	}
	if maxIndex != startRow {
		copyRow := make([]*fractional.Fraction, t.Cols)
		copy(copyRow, t.Matrix[startRow])
		t.Matrix[startRow] = t.Matrix[maxIndex]
		t.Matrix[maxIndex] = copyRow
	}
}

func (t *Table) checkRank() (int, error) {
	var rank, extendedRank int

	for _, rows := range t.Matrix {
		counter := 0

		for j := 0; j < t.Cols-1; j++ {
			if !rows[j].Equal(*fractional.ZeroValue) {
				counter++
			}
		}
		if counter != 0 {
			rank++
			extendedRank++
		} else if !rows[t.Cols-1].Equal(*fractional.ZeroValue) {
			extendedRank++
		}
	}

	if rank != extendedRank {
		return -1, fmt.Errorf("no solution")
	} else if rank < t.Cols-1 {
		return 1, nil
	}
	return 0, nil
}

func (t *Table) methodRectangle(newMatrix [][]*fractional.Fraction, resolveRow, resolveColumn int) error {
	for i := 0; i < t.Rows; i++ {
		if i == resolveRow {
			continue
		}
		for j := resolveColumn; j < t.Cols; j++ {
			if j == resolveColumn && i != resolveRow {
				newMatrix[i][j] = fractional.ZeroValue
			} else {
				subexpression, err := t.Matrix[i][resolveColumn].Multiply(*t.Matrix[resolveRow][j]).Divide(*t.Matrix[resolveRow][resolveColumn])
				if err != nil {
					return err
				}
				newMatrix[i][j] = t.Matrix[i][j].Subtract(*subexpression)
			}
		}
	}
	return nil
}

func (t *Table) ToNegativeRightSide() *Table {
	for _, rows := range t.Matrix {
		if rows[t.Cols-1 : t.Cols][0].Numerator() > 0 {
			for j, cols := range rows {
				rows[j] = cols.Reverse()
			}
		}
	}
	return t
}

func (t *Table) IsContainedInBasis(index int) bool {
	for _, basisVar := range t.BasisVars {
		if index == basisVar {
			return true
		}
	}
	return false
}

func (t *Table) CopyMatrix() [][]*fractional.Fraction {
	newMatrix := make([][]*fractional.Fraction, t.Rows)
	for r := range newMatrix {
		newMatrix[r] = make([]*fractional.Fraction, t.Cols)
		copy(newMatrix[r], t.Matrix[r])
	}
	return newMatrix
}

func (t *Table) CopyZ() []*fractional.Fraction {
	newZ := make([]*fractional.Fraction, t.Cols-1)
	copy(newZ, t.Z)
	return newZ
}

func (t *Table) CopyZFree() *fractional.Fraction {
	newZFree, _ := fractional.New(
		t.ZFree.Numerator(),
		t.ZFree.Denominator(),
	)
	return newZFree
}

func (t *Table) CopyBasisVars() []int {
	newBasisVars := make([]int, len(t.BasisVars))
	copy(newBasisVars, t.BasisVars)
	return newBasisVars
}
