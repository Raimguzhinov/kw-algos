package fractional

import (
	"errors"
	"fmt"
	"math"
)

type Fraction struct {
	numerator   int64
	denominator int64
}

type integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

var (
	ErrDivideByZero    = errors.New("denominator cannot be zero")
	ErrZeroDenominator = errors.New("denominator cannot be zero")

	ZeroValue = &Fraction{
		numerator: 0, denominator: 1,
	}
	RevOneValue = &Fraction{
		numerator: -1, denominator: 1,
	}
	OneValue = &Fraction{
		numerator: 1, denominator: 1,
	}
	MinValue = &Fraction{
		numerator: math.MinInt64, denominator: 1,
	}
	MaxValue = &Fraction{
		numerator: math.MaxInt64, denominator: 1,
	}
)

func New[T, K integer](numerator T, denominator K) (*Fraction, error) {
	if denominator == 0 {
		return ZeroValue, ErrZeroDenominator
	}
	if numerator == 0 {
		return ZeroValue, nil
	}

	n := int64(numerator)
	d := int64(denominator)
	if d < 0 {
		d *= -1
		n *= -1
	}
	gcf := gcd(abs(n), d)

	return &Fraction{
		numerator:   n / gcf,
		denominator: d / gcf,
	}, nil
}

func (f1 *Fraction) Add(f2 Fraction) *Fraction {
	m := lcm(f1.denominator, f2.denominator)
	sum := &Fraction{
		numerator:   f1.numerator*(m/f1.denominator) + f2.numerator*(m/f2.denominator),
		denominator: m,
	}
	return sum.shrink()
}

func (f1 *Fraction) Divide(f2 Fraction) (*Fraction, error) {
	f, err := New(f1.numerator*f2.denominator, f1.denominator*f2.numerator)
	if err != nil {
		err = ErrDivideByZero
	}
	return f.shrink(), err
}

func (f1 *Fraction) Equal(f2 Fraction) bool {
	return f1.numerator == f2.numerator && f1.denominator == f2.denominator
}

func (f1 *Fraction) NotEqual(f2 Fraction) bool {
	return !f1.Equal(f2)
}

func (f1 *Fraction) Multiply(f2 Fraction) *Fraction {
	f, _ := New(f1.numerator*f2.numerator, f1.denominator*f2.denominator)
	return f.shrink()
}

func (f1 *Fraction) Subtract(f2 Fraction) *Fraction {
	f2.numerator *= -1
	return f1.Add(f2).shrink()
}

func (f *Fraction) Float64() float64 {
	f = f.shrink()
	return float64(f.numerator) / float64(f.denominator)
}

func (f *Fraction) String() string {
	if f.denominator == 1 {
		return fmt.Sprintf("%d", f.numerator)
	}
	f = f.shrink()
	return fmt.Sprintf("%d/%d", f.numerator, f.denominator)
}

func (f *Fraction) shrink() *Fraction {
	gcf := gcd(abs(f.numerator), f.denominator)
	return &Fraction{f.numerator / gcf, f.denominator / gcf}
}

func (f1 *Fraction) Denominator() int64 {
	return f1.denominator
}

func (f1 *Fraction) Numerator() int64 {
	return f1.numerator
}

func (f1 *Fraction) LessThan(f2 Fraction) bool {
	return (f1.numerator * f2.denominator) < (f2.Numerator() * f1.denominator)
}

func (f1 *Fraction) GreaterThan(f2 Fraction) bool {
	return (f1.numerator * f2.denominator) > (f2.Numerator() * f1.denominator)
}

func (f *Fraction) Reverse() *Fraction {
	return f.Multiply(*RevOneValue).shrink()
}

func (f *Fraction) Abs() *Fraction {
	if f.numerator < 0 {
		return &Fraction{
			numerator:   -f.numerator,
			denominator: f.denominator,
		}
	}
	return f.shrink()
}

func abs[T integer](n T) T {
	if n < 0 {
		return -n
	}
	return n
}

func gcd(n1, n2 int64) int64 {
	for n2 != 0 {
		n1, n2 = n2, n1%n2
	}
	return n1
}

func lcm(n1, n2 int64) int64 {
	if n1 > n2 {
		n1, n2 = n2, n1
	}
	return n1 * (n2 / gcd(n1, n2))
}
