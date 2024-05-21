package simplex

import (
	"fmt"
	"kw-algos/fractional"
)

type Methods interface {
	DualMethod()
}

type Method struct {
	Table        *Table
	CO           []*fractional.Fraction
	isDualMethod bool
}

func New(table *Table) *Method {
	return &Method{
		Table:        table,
		isDualMethod: true,
	}
}

func (m *Method) String() string {
	var s string
	var offset = 8

	s += fmt.Sprintf(" B.V%*s%*d%*s", 2, "|", offset, 1, 2, "|")
	for i := range m.Table.Cols - 1 {
		s += fmt.Sprintf("%*sx%d%*s", offset/2, "", i+1, offset/2-2, "")
	}
	if !m.isDualMethod {
		s += " |	CO"
	}
	s += "\n"
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
		if !m.isDualMethod {
			s += fmt.Sprintf("%*s", offset-3, "|")
			if m.CO[i] == nil {
				s += fmt.Sprintf("%*s", offset, "-")
			} else {
				s += fmt.Sprintf("%*s", offset, m.CO[i])
			}
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
	if m.isDualMethod {
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
					s += fmt.Sprintf("%*s", offset, co)
				} else {
					s += fmt.Sprintf("%*s", offset, co)
				}
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
		var resolveRow, resolveColumn int

		isOptimal := true
		isZStringIsNegative := false
		isResolveRowIsNegative := false    //	Для двойственного симплекс метода
		isResolveColumnIsPositive := false //	Для стандартного симплекс метода

		//	В столбце свободных членов ищем самый минимальный отрицательный элемент
		//	если такого нет ставим 1ый признак оптимальности
		var maxValue *fractional.Fraction
		prepareMaxValue := false
		for i := range m.Table.Rows {
			if m.Table.Matrix[i][m.Table.Cols-1].LessThan(*fractional.ZeroValue) {
				if !prepareMaxValue {
					prepareMaxValue = true
					maxValue = m.Table.Matrix[i][m.Table.Cols-1]
					resolveRow = i
					isOptimal = false
				} else {
					if maxValue.GreaterThan(*m.Table.Matrix[i][m.Table.Cols-1]) {
						maxValue = m.Table.Matrix[i][m.Table.Cols-1]
						resolveRow = i
						isOptimal = false
					}
				}
			}
		}

		//	В Z-строке ищем нет ли нулей (признак того что решение не единственое)
		//	параллельно проверяем есть ли в строке отрицателные элементы
		// 	(если есть отриц. элемент и при этом 1ый признак оптимальности присутствует, нужно применить обычный симплекс метод)
		minNegativeZValueIndex := 0
		for i, z := range m.Table.Z {
			if z.LessThan(*fractional.ZeroValue) {
				isZStringIsNegative = true
				if m.Table.Z[minNegativeZValueIndex].GreaterThan(*z) {
					minNegativeZValueIndex = i
				}
			} else if _, ok := m.Table.IsContainedInBasis(i); z.Equal(*fractional.ZeroValue) && !ok {
				resolveColumn = i
				if InfinityCycles == 1 {
					linearCombinations(m.Table, infinityCopyTable)
					m.printAnswer()
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
				break
			}
		}

		//	Если есть 1ый признак оптиальности, Z-строка положительная и нет признака того что реш. не единственно - Получено оптимальное решение!
		if isOptimal && !isZStringIsNegative && InfinityCycles == -1 {
			fmt.Println(m)
			m.printAnswer()
			return nil
		}

		if !isOptimal {
			// Вычисление двойственных CO
			m.CO = make([]*fractional.Fraction, m.Table.Cols-1)
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
		} else {
			// Вычисление обычных CO
			if InfinityCycles == -1 {
				resolveColumn = minNegativeZValueIndex
			}
			m.CO = make([]*fractional.Fraction, m.Table.Rows)
			var err error
			for i := range m.Table.Rows {
				if m.Table.Matrix[i][resolveColumn].GreaterThan(*fractional.ZeroValue) {
					isResolveColumnIsPositive = true
					m.CO[i], err = m.Table.Matrix[i][m.Table.Cols-1].Divide(*m.Table.Matrix[i][resolveColumn])
					if err != nil {
						return err
					}
				}
			}
			m.isDualMethod = false
			fmt.Println(m)
		}
		fmt.Printf("\n")

		if isOptimal && !isResolveColumnIsPositive {
			return fmt.Errorf("no solutions\n")
		}
		if !isResolveRowIsNegative && InfinityCycles == -1 {
			if !isOptimal {
				return fmt.Errorf("no solutions\n")
			}
		}

		if InfinityCycles == -1 && !isOptimal {
			resolveColumn = m.findMinimumValueInCO()
		} else {
			resolveRow = m.findMinimumValueInCO()
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
}

func (m *Method) findMinimumValueInCO() int {
	var minElem *fractional.Fraction
	prepareMinElem := false
	minIndex := 0
	for i, co := range m.CO {
		if co == nil {
			continue
		} else {
			if !prepareMinElem {
				prepareMinElem = true
				minElem = co
				minIndex = i
			}
		}
		if minElem != nil && co.LessThan(*minElem) {
			minElem = co
			minIndex = i
		}
	}
	return minIndex
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
		x2 := findIndex(i, table1)
		x1 := findIndex(i, table2)
		sub := x2.Subtract(*x1)
		if sub.LessThan(*fractional.ZeroValue) {
			fmt.Printf("%s%sλ", x1, sub)
		} else {
			fmt.Printf("%s+%sλ", x1, sub)
		}
		if i != table1.Vars-1 {
			fmt.Print("; ")
		}
	}
	fmt.Print(")")
}

func convertZString(t *Table) {
	for z := range t.Z {
		if _, ok := t.IsContainedInBasis(z); ok && t.Z[z].NotEqual(*fractional.ZeroValue) {
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

func (m *Method) printAnswer() {
	fmt.Printf("\n")
	if m.Table.IsMinimizationProblem {
		fmt.Print("Zmin(")
		m.Table.ZFree = m.Table.ZFree.Reverse()
	} else {
		fmt.Print("Zmax(")
	}
	for i := range m.Table.Vars {
		if index, ok := m.Table.IsContainedInBasis(i); ok {
			fmt.Printf("%s", m.Table.Matrix[index][m.Table.Cols-1])
		} else {
			fmt.Print("0")
		}
		if i != m.Table.Vars-1 {
			fmt.Print(";")
		}
	}
	fmt.Printf(") = %s\n", m.Table.ZFree)
}
