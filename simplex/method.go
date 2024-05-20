package simplex

import (
	"fmt"
	"kw-algos/fractional"
)

type Methods interface {
	DualMethod()
}

type Method struct {
	Table *Table
	CO    []*fractional.Fraction
}

func New(table *Table) *Method {
	return &Method{Table: table}
}

func (m *Method) String() string {
	var s string
	var offset = 8

	for i := 0; i < m.Table.Rows; i++ {
		s += fmt.Sprintf(" x%d%*s", m.Table.BasisVars[i]+1, 3, "|")
		s += fmt.Sprintf("%*s", offset, m.Table.Matrix[i][m.Table.Cols-1])
		s += fmt.Sprintf("%*s", 2, "|")
		for j := 0; j < m.Table.Cols-1; j++ {
			offset := offset
			if j == 0 {
				offset = 5
			}
			s += fmt.Sprintf("%*s", offset, m.Table.Matrix[i][j])
		}
		s += "\n"
	}
	s += fmt.Sprintf("  Z%*s", 3, "|")
	s += fmt.Sprintf("%*s", offset, m.Table.ZFree)
	s += fmt.Sprintf("%*s", 2, "|")
	for j, z := range m.Table.Z {
		offset := offset
		if j == 0 {
			offset = 5
		}
		s += fmt.Sprintf("%*s", offset, z)
	}
	s += fmt.Sprintf("\n CO%*s%*s%*s", 3, "|", offset, "", 2, "|")
	for i, co := range m.CO {
		offset := offset
		if co == nil {
			if i == 0 {
				offset = 5
			}
			s += fmt.Sprintf("%*s", offset, "-")
		} else {
			if i == 0 {
				offset = 5
			} else {
				s += fmt.Sprintf("%*s", offset, co)
			}
		}
	}
	return s
}

func (m *Method) DualMethod() error {
	convertZString(m.Table)

	InfinityCycles := -1
	var infinityCopyTable *Table

	for {
		maxValue := fractional.MaxValue
		var resolveRow, resolveColumn int
		isOptimal := true
		isZStringIsNegative := false
		//isNotBasisValueEqualZero := false

		for i := range m.Table.Rows {
			if m.Table.Matrix[i][m.Table.Cols-1].LessThan(
				*fractional.ZeroValue,
			) && maxValue.GreaterThan(*m.Table.Matrix[i][m.Table.Cols-1]) {
				maxValue = m.Table.Matrix[i][m.Table.Cols-1]
				resolveRow = i
				isOptimal = false
			}
		}
		for i, z := range m.Table.Z {
			if z.LessThan(*fractional.ZeroValue) {
				isZStringIsNegative = true
			} else if z.Equal(*fractional.ZeroValue) && !m.Table.IsContainedInBasis(i) {
				resolveColumn = i
				if InfinityCycles == 1 {
					linearCombinations(m.Table, infinityCopyTable)
					return nil
				}
				InfinityCycles++
				if InfinityCycles == 0 {
					fmt.Println("solution is optimal, but not the only one")
				}
				infinityCopyTable = &Table{
					Z:         m.Table.CopyZ(),
					ZFree:     m.Table.CopyZFree(),
					Matrix:    m.Table.CopyMatrix(),
					BasisVars: m.Table.CopyBasisVars(),
				}
			}
		}
		if !isZStringIsNegative && InfinityCycles == -1 {
			fmt.Println(m)
			fmt.Printf("\n")
			return nil
		}

		m.CO = make([]*fractional.Fraction, m.Table.Cols-1)
		isResolveRowIsNegative := false
		for j := range m.Table.Cols - 1 {
			if m.Table.Matrix[resolveRow][j].LessThan(*fractional.ZeroValue) {
				isResolveRowIsNegative = true
				divide, err := m.Table.Z[j].Divide(*m.Table.Matrix[resolveRow][j])
				if err != nil {
					return err
				}
				m.CO[j] = divide.Abs()
			}
		}

		fmt.Println(m)
		fmt.Printf("\n")

		if !isResolveRowIsNegative && InfinityCycles == -1 {
			if isOptimal {
				fmt.Println(m)
				fmt.Printf("\n")
				panic("VIHOD 2!!!")
			} else {
				panic("NO SOLUTION!!!")
			}
		} else {
			maxValue := fractional.MaxValue
			for i := range m.Table.Rows {
				if maxValue.GreaterThan(*m.Table.Matrix[i][m.Table.Cols-1]) {
					maxValue = m.Table.Matrix[i][m.Table.Cols-1]
					resolveRow = i
				}
			}
		}

		if InfinityCycles == -1 {
			minElem := fractional.MaxValue
			for i, co := range m.CO {
				if co == nil {
					continue
				}
				if co.LessThan(*minElem) {
					minElem = co
					resolveColumn = i
				}
			}
		}

		newTable := &Table{
			Z:      m.Table.CopyZ(),
			ZFree:  m.Table.CopyZFree(),
			Matrix: m.Table.CopyMatrix(),
		}

		resolver := m.Table.Matrix[resolveRow][resolveColumn]
		for j := range m.Table.Cols {
			var err error
			newTable.Matrix[resolveRow][j], err = m.Table.Matrix[resolveRow][j].Divide(*resolver)
			if err != nil {
				return err
			}
		}

		if err := m.methodRectangle(newTable, resolveRow, resolveColumn); err != nil {
			return err
		}

		m.Table.Matrix = newTable.Matrix
		m.Table.Z = newTable.Z
		m.Table.ZFree = newTable.ZFree
		m.Table.BasisVars[resolveRow] = resolveColumn
	}
	return nil
}

func linearCombinations(table1 *Table, table2 *Table) {
	findIndex := func(varIndex int, table *Table) *fractional.Fraction {
		for i, basisVar := range table.BasisVars {
			if varIndex == basisVar {
				return table.Matrix[i][table1.Cols-1]
			}
		}
		return fractional.ZeroValue
	}
	fmt.Printf("x^(*) = (")
	for i := range table1.Vars {
		x1 := findIndex(i, table1)
		x2 := findIndex(i, table2)
		sub := x2.Subtract(*x1)
		if sub.LessThan(*fractional.ZeroValue) {
			fmt.Printf("%s + %sλ", x1, sub.Reverse())
		} else {
			fmt.Printf("%s - %sλ", x1, sub)
		}
		if i != table1.Vars-1 {
			fmt.Print("; ")
		}
	}
	fmt.Print(")")
}

func convertZString(t *Table) {
	for z := range t.Z {
		if t.IsContainedInBasis(z) && t.Z[z].NotEqual(*fractional.ZeroValue) {
			var row int
			for i := range t.Rows {
				if t.Matrix[i][z].NotEqual(*fractional.ZeroValue) {
					row = i
					break
				}
			}
			for j := range t.Cols - 1 {
				if j != z {
					added := t.Matrix[row][j].Multiply(*fractional.RevOneValue).Multiply(*t.Z[z])
					t.Z[j] = t.Z[j].Add(*added)
				}
			}
			t.ZFree = t.ZFree.Add(*t.Matrix[row][t.Cols-1].Multiply(*t.Z[z]))
			t.Z[z] = t.Z[z].Multiply(*fractional.ZeroValue)
		}
	}
	for i := range t.Z {
		t.Z[i] = t.Z[i].Reverse()
	}
}

func (m *Method) methodRectangle(newTable *Table, resolveRow, resolveColumn int) error {
	var err error

	matrixFormula := func(i, j int) (*fractional.Fraction, error) {
		subexpression, err := m.Table.Matrix[i][resolveColumn].Multiply(
			*m.Table.Matrix[resolveRow][j]).Divide(
			*m.Table.Matrix[resolveRow][resolveColumn])
		if err != nil {
			return nil, err
		}
		return m.Table.Matrix[i][j].Subtract(*subexpression), nil
	}
	zFormula := func(zIndex int) (*fractional.Fraction, error) {
		subexpression, err := m.Table.Z[resolveColumn].Multiply(
			*m.Table.Matrix[resolveRow][zIndex]).Divide(
			*m.Table.Matrix[resolveRow][resolveColumn])
		if err != nil {
			return nil, err
		}
		return m.Table.Z[zIndex].Subtract(*subexpression), nil
	}
	zFreeFormula := func() (*fractional.Fraction, error) {
		subexpression, err := m.Table.Z[resolveColumn].Multiply(
			*m.Table.Matrix[resolveRow][m.Table.Cols-1]).Divide(
			*m.Table.Matrix[resolveRow][resolveColumn])
		if err != nil {
			return nil, err
		}
		return m.Table.ZFree.Subtract(*subexpression), nil
	}

	for i := 0; i < m.Table.Rows; i++ {
		if i == resolveRow {
			continue
		}
		newTable.Matrix[i][m.Table.Cols-1], err = matrixFormula(i, m.Table.Cols-1)
		if err != nil {
			return err
		}
		for j := 0; j < m.Table.Cols-1; j++ {
			if j == resolveColumn && i != resolveRow {
				newTable.Matrix[i][j] = fractional.ZeroValue
			} else {
				newTable.Matrix[i][j], err = matrixFormula(i, j)
				if err != nil {
					return err
				}
			}
		}
	}
	for i := range m.Table.Z {
		newTable.Z[i], err = zFormula(i)
		if err != nil {
			return err
		}
	}

	newTable.ZFree, err = zFreeFormula()
	if err != nil {
		return err
	}

	return nil
}
