package rng

import (
	"fmt"
	"math"
	"strings"
)

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

type Histogram struct {
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
		b := strings.Repeat("─", w)
		y := n / t
		if !cumul {
			y *= h.k
		}
		x := h.X(float64(i))
		fmt.Printf("%9.2f  ", x)
		if d == h.Peak {
			fmt.Print("►")
		} else {
			fmt.Print(" ")
		}
		fmt.Printf("│%s  %.2f%%  %.0f", b, y*100, n)
		fmt.Println()
	}
	fmt.Printf("μ = %.5f  σ = %.5f\n", h.Calc.Avg, h.Calc.Dev)
}

func HistTest() {
	var h Histogram
	h.Reset()
	h.Scale(1, 4)
	for h.Calc.Cnt < 1000000 {
		x := h.RNG.Gamma(3.24)
		h.Add(float64(x))
	}
	h.Graph(100, 100, false)
}

func init() {
	// HistTest()
}
