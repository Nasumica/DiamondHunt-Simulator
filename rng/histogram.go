package rng

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

import (
	"fmt"
	"math"
	"strings"
)

type Histogram struct {
	Title      string
	Calc       StatCalc
	Data       map[int]int
	Min, Max   int
	Mode, Peak int
	x1, x2     float64
	y1, y2     float64
	dx, dy, k  float
	f          []func(x float64) float64
	RNG        LCPRNG
}

func (h *Histogram) Reset(f ...func(x float64) float64) {
	h.Calc.Reset()
	h.RNG.Randomize()
	h.Data = map[int]int{}
	h.f = f
	h.Scale(1, 1)
	h.Peak = 0
}

func (h *Histogram) Map(x1, x2, y1, y2 float64) {
	h.x1, h.x2 = x1, x2
	h.y1, h.y2 = y1, y2
	h.dy = h.y2 - h.y1
	h.dx = h.x2 - h.x1
	h.k = h.dy / h.dx
}

func (h *Histogram) Scale(p, q float64) {
	h.Map(0, p, 0, q)
}

func (h *Histogram) Y(x float64) float64 {
	y := (x-h.x1)*h.k + h.y1
	return y
}

func (h *Histogram) X(y float64) float64 {
	x := (y-h.y1)/h.k + h.x1
	return x
}

func (h *Histogram) N(x float64) int {
	y := h.Y(x)
	for _, f := range h.f {
		y = f(y)
	}
	return int(math.Round(y))
}

func (h *Histogram) Add(x float64) {
	h.Calc.Add(x)
	n := h.N(x)
	m := h.Data[n] + 1
	h.Data[n] = m
	if h.Calc.Cnt == 1 {
		h.Min, h.Max = n, n
	} else if h.Min > n {
		h.Min = n
	} else if h.Max < n {
		h.Max = n
	}
	if h.Peak < m {
		h.Mode = n
		h.Peak = m
	}
}

func (h *Histogram) Graph(width int, limit int, cumul bool) {
	iif := func(c bool, t, f string) string {
		if c {
			return t
		} else {
			return f
		}
	}
	fmt.Println()
	df := iif(cumul, "PDF", "CDF")
	fmt.Printf("%s %s:   n = %d   [%.2f, %.2f]   μ = %.5f  σ = %.5f\n",
		df, h.Title, h.Calc.Cnt, h.Calc.Min, h.Calc.Max, h.Calc.Avg, h.Calc.Dev)
	p, t := float64(h.Peak), float64(h.Calc.Cnt)
	if cumul {
		p = t
	}
	s := p / float64(width)
	lo := h.RNG.Censor(h.Min, h.Mode-limit, h.Max)
	hi := h.RNG.Censor(h.Min, h.Mode+limit, h.Max)
	c := 0.
	for i, d := range h.Data {
		if i < lo {
			c += float64(d)
		}
	}
	for i := lo; i <= hi; i++ {
		d := h.Data[i]
		n := float64(d)
		if cumul {
			c += n
			n = c
		}
		w := int(math.Round(n / s))
		b := strings.Repeat("-", w)
		y := n / t
		if !cumul {
			y *= h.k
		}
		x := h.X(float64(i))
		fmt.Printf("%9.2f  %s", x, iif(d == h.Peak, "►", " "))
		fmt.Printf("%s%s  %.2f%%  %.0f", iif(x == 0, "┤", "│"), b, y*100, n)
		fmt.Println()
	}
}

func HistTest(n int) {
	var h Histogram
	{
		h.Reset()
		h.Scale(1, 4)
		ɑ := 3.25
		h.Title = fmt.Sprintf("Gamma distribution (ɑ = %v)", ɑ)
		for h.Calc.Cnt < n {
			x := h.RNG.Gamma(ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 50, false)
	}
	{
		h.Reset()
		h.Scale(1, 1)
		ƛ := 3.61
		h.Title = fmt.Sprintf("Poisson distribution (ƛ = %v)", ƛ)
		for h.Calc.Cnt < n {
			x := h.RNG.Poisson(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 50, false)
	}
	{
		h.Reset()
		h.Scale(1, 10)
		ξ, ω, ɑ := 0., 1., 4.
		h.Title = fmt.Sprintf("Skew-normal distribution (ξ = %v, ω = %v, ɑ = %v)", ξ, ω, ɑ)
		for h.Calc.Cnt < n {
			x := h.RNG.SkewNormal(ξ, ω, ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 50, false)
	}
	{
		h.Reset(math.Floor)
		h.Scale(1, 5)
		ƛ := 2.71
		h.Title = fmt.Sprintf("Exponential distribution (ƛ = %v)", ƛ)
		for h.Calc.Cnt < n {
			x := h.RNG.Exponential(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
		h.Graph(100, 100, true)
	}
	{
		h.Reset()
		h.Scale(1, 1)
		ɑ1, ɑ2 := 0.2, 1.9
		h.Title = fmt.Sprintf("Hermite distribution (ɑ1 = %v, ɑ2  = %v)", ɑ1, ɑ2)
		for h.Calc.Cnt < n {
			x := h.RNG.Hermite(ɑ1, ɑ2)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
}

func init() {
	//	HistTest(1 * 1000 * 1000)
}
