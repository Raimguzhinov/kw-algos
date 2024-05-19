package matrix

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

type Matrix struct {
	rows, cols, vars      int
	matrix                [][]*fractional.Fraction
	comparisons           []Comparison
	Z                     []*fractional.Fraction
	IsMinimizationProblem bool
}

func New(r io.Reader) (*Matrix, error) {
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
	}

	maxMinSign := parts[vars]
	var isMinimization bool
	if maxMinSign == "min" {
		isMinimization = true
	} else {
		isMinimization = false
	}

	return &Matrix{
		rows,
		cols,
		vars,
		matrix,
		comparisons,
		Z,
		isMinimization,
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

func (m *Matrix) String() string {
	var s string
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			s += fmt.Sprintf("%*s ", 6, m.matrix[i][j])
		}
		s += "\n"
	}
	s += "Z = "
	for i, z := range m.Z {
		if i == 0 {
			s += fmt.Sprintf("%s", z)
		} else {
			s += fmt.Sprintf("%*s", 7, z)
		}
	}
	s += "\n"
	return s
}

func (m *Matrix) ToCanonicalForm() *Matrix {
	var newColsCnt, beforeNormalization int
	basis := make(map[int][]*fractional.Fraction)
	var lasts []*fractional.Fraction

	for i, comparison := range m.comparisons {
		var b *fractional.Fraction

		switch comparison {
		case LessThanOrEqualTo:
			b, _ = fractional.New(1, 1)
			m.Z = append(m.Z, fractional.ZeroValue)
		case GreaterThanOrEqualTo:
			b, _ = fractional.New(-1, 1)
			m.Z = append(m.Z, fractional.ZeroValue)
		default:
			beforeNormalization++
		}
		newColsCnt++
		last := m.matrix[i][m.vars]
		lasts = append(lasts, last)
		basis[i] = make([]*fractional.Fraction, m.rows)
		for j := 0; j < m.rows; j++ {
			basis[i][j] = fractional.ZeroValue
		}
		basis[i][i] = b
	}
	if newColsCnt > 0 {
		m.cols += newColsCnt
		for i := 0; i < m.rows; i++ {
			head := m.matrix[i][:m.vars]
			tail := append(m.matrix[i][m.vars+1:], basis[i]...)
			tail = append(tail, lasts[i])
			m.matrix[i] = append(head, tail...)
			m.comparisons[i] = EqualTo
		}
	}
	m.matrix = normalizeMatrix(m.matrix)
	m.cols -= beforeNormalization
	if m.IsMinimizationProblem {
		for i := 0; i < m.cols-1; i++ {
			reverse, _ := fractional.New(-1, 1)
			m.Z[i] = m.Z[i].Multiply(*reverse)
		}
	}
	return m
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

func (m *Matrix) ToBasis() (*Matrix, error) {
	var columOfResolver int
	for i := 0; i < m.rows; i++ {

		m.swapMatrixRows(i, columOfResolver)
		needToCheck := false
		//newMatrix := m.matrix
		newMatrix := make([][]*fractional.Fraction, m.rows)
		for r := range newMatrix {
			newMatrix[r] = make([]*fractional.Fraction, m.cols)
			copy(newMatrix[r], m.matrix[r])
		}

		currentColumOfResolver := columOfResolver
		if m.matrix[i][currentColumOfResolver].Equal(*fractional.ZeroValue) {
			needToCheck = true
			for j := currentColumOfResolver + 1; j < m.cols-1; j++ {
				m.swapMatrixRows(i, j)

				if !m.matrix[i][j].Equal(*fractional.ZeroValue) {
					newMatrix = m.matrix
					needToCheck = false
					columOfResolver = j + 1
					currentColumOfResolver = j

					break
				}
			}
			if needToCheck && m.matrix[i][m.cols-1].Equal(*fractional.ZeroValue) {
				rank, err := m.checkRank()
				if err != nil {
					return nil, err
				}
				if i == m.rows-1 && rank == 1 {
					return m, nil
				}
				continue
			} else if currentColumOfResolver == columOfResolver {
				_, err := m.checkRank()
				if err != nil {
					return nil, err
				}
				return m, err
			}
		} else {
			columOfResolver++
		}

		reverse, _ := fractional.New(-1, 1)
		if !m.matrix[i][currentColumOfResolver].Equal(*reverse) {
			divider := m.matrix[i][currentColumOfResolver]
			for j := 0; j < m.cols; j++ {
				var err error
				newMatrix[i][j], err = m.matrix[i][j].Divide(*divider)
				if err != nil {
					return nil, err
				}
			}
		}

		if err := m.methodRectangle(newMatrix, i, currentColumOfResolver); err != nil {
			return nil, err
		}
		m.matrix = newMatrix

		if i == m.rows-1 || needToCheck {
			_, err := m.checkRank()
			if err != nil {
				return nil, err
			}
			return m, err
		}
	}
	return m, nil
}

func (m *Matrix) swapMatrixRows(startRow, startColumn int) {
	maxValue := math.Abs(m.matrix[startRow][startColumn].Float64())
	maxIndex := startRow
	for j := startRow; j < m.rows; j++ {
		if maxValue < math.Abs(m.matrix[j][startColumn].Float64()) {
			maxValue = math.Abs(m.matrix[j][startColumn].Float64())
			maxIndex = j
		}
	}
	if maxIndex != startRow {
		copyRow := make([]*fractional.Fraction, m.cols)
		copy(copyRow, m.matrix[startRow])
		m.matrix[startRow] = m.matrix[maxIndex]
		m.matrix[maxIndex] = copyRow
	}
}

func (m *Matrix) checkRank() (int, error) {
	var rank, extendedRank int

	for _, rows := range m.matrix {
		counter := 0

		for j := 0; j < m.cols-1; j++ {
			if !rows[j].Equal(*fractional.ZeroValue) {
				counter++
			}
		}
		if counter != 0 {
			rank++
			extendedRank++
		} else if !rows[m.cols-1].Equal(*fractional.ZeroValue) {
			extendedRank++
		}
	}

	if rank != extendedRank {
		return -1, fmt.Errorf("no solution")
	} else if rank < m.cols-1 {
		return 1, nil
	}
	return 0, nil
}

func (m *Matrix) methodRectangle(newMatrix [][]*fractional.Fraction, resolverRow, resolverColumn int) error {
	for i := 0; i < m.rows; i++ {
		if i == resolverRow {
			continue
		}
		for j := resolverColumn; j < m.cols; j++ {
			if j == resolverColumn && i != resolverRow {
				newMatrix[i][j] = fractional.ZeroValue
			} else {
				subexpression, err := m.matrix[i][resolverColumn].Multiply(*m.matrix[resolverRow][j]).Divide(*m.matrix[resolverRow][resolverColumn])
				if err != nil {
					return err
				}
				newMatrix[i][j] = m.matrix[i][j].Subtract(*subexpression)
			}
		}
	}
	return nil
}

func (m *Matrix) ToNegativeRightSide() *Matrix {
	for _, rows := range m.matrix {
		if rows[m.cols-1 : m.cols][0].Numerator() > 0 {
			for j, cols := range rows {
				reverse, _ := fractional.New(-1, 1)
				rows[j] = cols.Multiply(*reverse)
			}
		}
	}
	return m
}
