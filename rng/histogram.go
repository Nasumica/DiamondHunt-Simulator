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

func (h *Histogram) Graph(width int, limit int, cumul bool, nozero ...bool) {
	iif := func(c bool, t, f string) string {
		if c {
			return t
		} else {
			return f
		}
	}
	trim := func(s, r string) string {
		l := len(s) - 1
		for ; s[l] == '0'; l-- {
			s = s[:l] + r + s[l+1:]
		}
		if s[l] == '.' {
			s = s[:l] + r + s[l+1:]
		}
		return s
	}
	str := func(a any) string {
		return trim(fmt.Sprintf("%.2f", a), "")
	}
	nz := false
	if len(nozero) > 0 {
		nz = nozero[0]
	}
	fmt.Println()
	df := iif(cumul, "CDF", "PDF")
	fmt.Printf("%s %s:   n = %d   [%v, %v]   μ = %.5f  σ = %.5f\n",
		df, h.Title, h.Calc.Cnt, str(h.Calc.Min), str(h.Calc.Max), h.Calc.Avg, h.Calc.Dev)
	p, t := float64(h.Peak), float64(h.Calc.Cnt)
	if cumul {
		p = t
	}
	s := p / float64(width)
	lo, hi := h.Min, h.Max
	if !nz {
		lo = h.RNG.Censor(h.Min, h.Mode-limit, h.Max)
		hi = h.RNG.Censor(h.Min, h.Mode+limit, h.Max)
	}
	c := 0.
	for i, d := range h.Data {
		if i < lo {
			c += float64(d)
		}
	}
	for i := lo; i <= hi; i++ {
		d := h.Data[i]
		if d != 0 || !nz {
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
			v := trim(fmt.Sprintf("%.2f", 100*y), "") + "%"
			x := h.X(float64(i))
			u := trim(fmt.Sprintf("%9.2f", x), " ")
			fmt.Printf("%v  %s", u, iif(d == h.Peak, "►", " "))
			fmt.Printf("%s%s  %v  %.0f", iif(x == 0, "┤", "│"), b, v, n)
			fmt.Println()
		}
	}
}

func HistTest(n int) {
	var h Histogram
	if true {
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
	if true {
		h.Reset()
		h.Scale(1, 1)
		ƛ := float64(h.RNG.Int(200, 400)) / 100
		h.Title = fmt.Sprintf("Poisson distribution (ƛ = %v)", ƛ)
		for h.Calc.Cnt < n {
			x := h.RNG.Poisson(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 50, false)
	}
	if true {
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
	if true {
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
	if true {
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
	if true {
		h.Reset()
		h.Scale(1, 1)
		p := float64(h.RNG.Int(55, 65)) / 100
		h.Title = fmt.Sprintf("Geometric distribution (p = %v)", p)
		for h.Calc.Cnt < n {
			x := h.RNG.Geometric(p)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if false {
		h.Reset()
		h.Scale(1, 1)
		h.Title = "Sic-Bo"
		for h.Calc.Cnt < n {
			d, _, _ := h.RNG.SicBo()
			h.RNG.Sort(&d)
			x := 0
			for _, i := range d {
				x = x*10 + i
			}
			h.Add(float64(x))
		}
		h.Graph(100, 100, false, true)
	}
}

func init() {
	// HistTest(1 * 1000 * 1000)
}
