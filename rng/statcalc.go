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
	Nul int     `json:"nul,omitempty"` // zero items count
	Val float64 `json:"val"`           // last value, x
	Cat string  `json:"cat,omitempty"` // category name
}

// Add single value to statistical calculator.
func (sc *StatCalc) put(x float64) {
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
}

// Reset statistical calculator.
func (s *StatCalc) Reset() {
	s.Cnt, s.Sum, s.Sqr, s.Min, s.Max = 0, 0, 0, 0, 0
	s.Abe, s.Avg, s.Dev, s.Nul, s.Val = 0, 0, 0, 0, 0
}

// Add values to statistical calculator.
func (sc *StatCalc) Add(values ...float64) float64 {
	for _, x := range values {
		sc.put(x)
	}
	return sc.Sum
}

// Add integers to statistical calculator.
func (sc *StatCalc) Int(items ...int) float64 {
	for _, i := range items {
		sc.put(float64(i))
	}
	return sc.Sum
}
