package rng

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

import (
	"math"
)

// # Statistical calculator
type StatCalc struct {
	Cnt int     `json:"cnt"`           // sample length, n
	Sum float64 `json:"sum"`           // sum of items, Σ x
	Min float64 `json:"min"`           // minimum
	Max float64 `json:"max"`           // maximum
	Avg float64 `json:"avg"`           // average, μ
	Dev float64 `json:"dev,omitempty"` // standard deviation, σ
	Abe float64 `json:"abe,omitempty"` // aberation, n · variance
	Sqr float64 `json:"sqr,omitempty"` // sum of squares, Σ x²
	Nul int     `json:"nul,omitempty"` // zeroes count
	Int int     `json:"int,omitempty"` // integers count
	Val float64 `json:"val"`           // last value, x
	Gcd uint64  `json:"gcd"`           // greatest common divisor (integers only)
	Cat string  `json:"cat,omitempty"` // category name
}

// # Add single value to statistical calculator.
func (sc *StatCalc) Put(x float64) {
	if sc.Cnt == 0 {
		sc.Reset()
		sc.Min = x
		sc.Max = x
	} else if sc.Min > x {
		sc.Min = x
	} else if sc.Max < x {
		sc.Max = x
	}
	sc.Cnt++
	if x == 0 {
		sc.Nul++
	} else {
		sc.Sum += x
		sc.Sqr += x * x
	}
	if sc.Min == sc.Max {
		sc.Avg = sc.Min
	} else {
		n := float64(sc.Cnt)
		sc.Val = sc.Avg
		sc.Avg = sc.Sum / n
		sc.Abe += (x - sc.Avg) * (x - sc.Val)
		sc.Dev = math.Sqrt(sc.Abe / n)
		// sc.Dev = math.Sqrt(math.Abs(n*sc.Sqr-sc.Sum*sc.Sum)) / n
	}
	sc.Val = x
	if x == math.Floor(x) {
		sc.Int++
		if sc.Gcd != 1 && x != 0 {
			if x = math.Abs(x); x <= math.MaxUint64 {
				for n := uint64(x); n != 0; {
					sc.Gcd, n = n, sc.Gcd%n
				}
			} else {
				sc.Gcd = 1
			}
		}
	}
}

// # Reset statistical calculator.
func (sc *StatCalc) Reset() {
	sc.Cnt, sc.Sum, sc.Sqr, sc.Min, sc.Max = 0, 0, 0, 0, 0
	sc.Abe, sc.Avg, sc.Dev, sc.Nul, sc.Val = 0, 0, 0, 0, 0
	sc.Int, sc.Gcd = 0, 0
}

// # Add values to statistical calculator.
func (sc *StatCalc) Add(values ...float64) float64 {
	for _, x := range values {
		sc.Put(x)
	}
	return sc.Sum
}
