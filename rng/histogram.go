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
	ox, oy     float64
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
	h.Mode, h.Peak = 0, 0
}

// # Define line (x1, y1)--(x2, y2)
func (h *Histogram) Linear(x1, y1, x2, y2 float64) {
	h.dy = y2 - y1
	h.dx = x2 - x1
	if h.dx == 0 || h.dy == 0 {
		h.Scale(1, 1)
	} else {
		h.k = h.dy / h.dx
		h.ox = x1
		h.oy = y1
	}
}

func (h *Histogram) Scale(sx, sy float64) {
	h.Linear(0, 0, sx, sy)
}

func (h *Histogram) Y(x float64) float64 {
	y := (x-h.ox)*h.k + h.oy
	return y
}

func (h *Histogram) X(y float64) float64 {
	x := (y-h.oy)/h.k + h.ox
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
		h.Title, df, h.Calc.Cnt, str(h.Calc.Min), str(h.Calc.Max), h.Calc.Avg, h.Calc.Dev)
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
			u := trim(fmt.Sprintf("%10.3f", x), " ")
			fmt.Printf("%v  %s", u, iif(d == h.Peak, "►", " "))
			fmt.Printf("%s%s  %v  %.0f", iif(x == 0, "┤", "│"), b, v, n)
			fmt.Println()
		}
	}
}

func HistTest(sample int) {
	var h Histogram

	par := func(a, b float64) float64 {
		const s = 100
		m, n := int(math.Round(s*a)), int(math.Round(s*b))
		return float64(h.RNG.Int(m, n)) / s
	}

	if true {
		h.Reset()
		h.Scale(1, 4)
		ɑ := par(1.5, 3.5)
		h.Title = fmt.Sprintf("Gamma distribution (ɑ = %v)", ɑ)
		for h.Calc.Cnt < sample {
			x := h.RNG.Gamma(ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 30, false)
	}
	if true {
		ɑ := par(0.4, 3.5)
		β := par(0.4, 3.5)
		if ɑ > β {
			h.Reset(math.Ceil)
		} else {
			h.Reset(math.Floor)
		}
		h.Scale(1, 25)
		h.Title = fmt.Sprintf("Beta distribution (ɑ = %v, β = %v)", ɑ, β)
		for h.Calc.Cnt < sample {
			x := h.RNG.Beta(ɑ, β)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if true {
		h.Reset()
		h.Scale(1, 10)
		ξ, ω, ɑ := 0., 1., par(-5, 5)
		h.Title = fmt.Sprintf("Skew-normal distribution (ξ = %v, ω = %v, ɑ = %v)", ξ, ω, ɑ)
		for h.Calc.Cnt < sample {
			x := h.RNG.SkewNormal(ξ, ω, ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 30, false)
	}
	if true {
		h.Reset(math.Floor)
		h.Scale(1, 5)
		ƛ := par(2, 3)
		h.Title = fmt.Sprintf("Exponential distribution (ƛ = %v)", ƛ)
		for h.Calc.Cnt < sample {
			x := h.RNG.Exponential(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
		// h.Graph(100, 100, true)
	}
	if true {
		h.Reset(math.Ceil)
		h.Scale(1, 5)
		rtp := par(0.9, 0.96)
		h.Title = fmt.Sprintf("House edgne distribution (rtp = %v%%)", 100*rtp)
		for h.Calc.Cnt < sample {
			x := h.RNG.Edge(rtp)
			h.Add(float64(x))
		}
		h.Graph(100, 30, false)
	}
	if true {
		h.Reset()
		ƛ := par(2, 4)
		h.Title = fmt.Sprintf("Poisson distribution (ƛ = %v)", ƛ)
		for h.Calc.Cnt < sample {
			x := h.RNG.Poisson(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 50, false)
	}
	if true {
		h.Reset()
		ɑ1 := par(0.05, 0.15) * 2
		ɑ2 := (5 - ɑ1) / 2
		h.Title = fmt.Sprintf("Hermite distribution (ɑ1 = %v, ɑ2 = %v)", ɑ1, ɑ2)
		for h.Calc.Cnt < sample {
			x := h.RNG.Hermite(ɑ1, ɑ2)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if true {
		h.Reset()
		μ1 := par(1.6, 3)
		μ2 := par(1.5, 2.9)
		h.Title = fmt.Sprintf("Skellam distribution (μ1 = %v, μ2 = %v)", μ1, μ2)
		for h.Calc.Cnt < sample {
			x := h.RNG.Skellam(μ1, μ2)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if true {
		h.Reset()
		p := par(0.4, 0.6)
		h.Title = fmt.Sprintf("Geometric distribution (p = %v)", p)
		for h.Calc.Cnt < sample {
			x := h.RNG.Geometric(p)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if true {
		h.Reset()
		n := 52
		p := 0.5
		h.Title = fmt.Sprintf("Binomial distribution (n = %d, p = %v)", n, p)
		for h.Calc.Cnt < sample {
			x := h.RNG.Binomial(n, p)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if true {
		h.Reset()
		h.Scale(1, 1)
		a, b := 0., 20.
		mode := par(a+1, b-1)
		h.Title = fmt.Sprintf("Triangular distribution (a = %v, b = %v, mode = %v)", a, b, mode)
		for h.Calc.Cnt < sample {
			x := h.RNG.Triangular(a, b, mode)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if false {
		h.Reset()
		h.Title = "3 dice throw sum"
		for h.Calc.Cnt < sample {
			x := h.RNG.Int(1, 6) + h.RNG.Int(1, 6) + h.RNG.Int(1, 6)
			h.Add(float64(x))
		}
		h.Graph(100, 100, false)
	}
	if false {
		h.Reset()
		h.Title = "Sic-Bo"
		for h.Calc.Cnt < sample {
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
