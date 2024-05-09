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

func (h *Histogram) Graph(width int, limit int, flags ...bool) {
	flag := func(i int) bool {
		return i < len(flags) && flags[i]
	}
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
	cumul, nozero := flag(0), flag(1)
	df := iif(cumul, "CDF", "PDF")
	fmt.Println()
	fmt.Printf("%s %s:   n = %d   [%v, %v]   μ = %.5f  σ = %.5f\n",
		h.Title, df, h.Calc.Cnt, str(h.Calc.Min), str(h.Calc.Max), h.Calc.Avg, h.Calc.Dev)
	p, t := float64(h.Peak), float64(h.Calc.Cnt)
	if cumul {
		p = t
	}
	s := float64(width) / p
	lo, hi := h.Min, h.Max
	if !nozero {
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
		if d != 0 || !nozero {
			n := float64(d)
			if cumul {
				c += n
				n = c
			}
			w := int(math.Round(n * s))
			b := strings.Repeat("-", w)
			y := n / t
			if !cumul {
				y *= h.k
			}
			v := trim(fmt.Sprintf("%.2f", 100*y), "") + "%"
			x := h.X(float64(i))
			u := trim(fmt.Sprintf("%10.3f", x), " ")
			fmt.Printf("%v  %s", u, iif(d == h.Peak, "►", " "))
			fmt.Printf("%s%s  %v  %.0f", iif(x == math.Floor(x), "┤", "│"), b, v, n)
			fmt.Println()
		}
	}
}

func HistTest(sample int) {
	var h Histogram
	h.Reset()

	par := func(a, b float64, q ...float64) float64 {
		s := 100.
		if len(q) > 0 {
			s = q[0]
		}
		m, n := int(math.Round(s*a)), int(math.Round(s*b))
		return float64(h.RNG.Int(m, n)) / s
	}

	if true {
		ɑ := par(0.5, 6)
		h.Title = fmt.Sprintf("Gamma distribution (ɑ = %v)", ɑ)
		h.Reset()
		h.Scale(1, 4)
		for h.Calc.Cnt < sample {
			x := h.RNG.Gamma(ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 30)
	}
	if true {
		ɑ := par(0.4, 3.5, 4)
		β := par(0.4, 3.5, 4)
		h.Title = fmt.Sprintf("Beta distribution (ɑ = %v, β = %v)", ɑ, β)
		if ɑ > β {
			h.Reset(math.Ceil)
		} else if ɑ < β {
			h.Reset(math.Floor)
		} else {
			h.Reset()
		}
		h.Scale(1, 25)
		for h.Calc.Cnt < sample {
			x := h.RNG.Beta(ɑ, β)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		ξ, ω, ɑ := 0., 1., par(-5, 5, 2)
		h.Title = fmt.Sprintf("Skew-normal distribution (ξ = %v, ω = %v, ɑ = %v)", ξ, ω, ɑ)
		h.Reset()
		h.Scale(1, 10)
		for h.Calc.Cnt < sample {
			x := h.RNG.SkewNormal(ξ, ω, ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 20)
	}
	if true {
		ν := par(1, 10)
		if h.RNG.Bernoulli(0.01) {
			ν = math.Inf(1)
		}
		h.Title = fmt.Sprintf("Student's t distribution (ν = %v)", ν)
		h.Reset()
		h.Scale(1, 4)
		for h.Calc.Cnt < sample {
			x := h.RNG.StudentsT(ν)
			h.Add(float64(x))
		}
		h.Graph(100, 20)
	}
	if true {
		ƛ := par(2, 3)
		h.Title = fmt.Sprintf("Exponential distribution (ƛ = %v)", ƛ)
		h.Reset(math.Floor)
		h.Scale(1, 5)
		for h.Calc.Cnt < sample {
			x := h.RNG.Exponential(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
		// h.Graph(100, 100, true)
	}
	if true {
		rtp := par(0.9, 0.96)
		h.Title = fmt.Sprintf("House edgne distribution (rtp = %v%%)", 100*rtp)
		h.Reset(math.Ceil)
		h.Scale(1, 5)
		for h.Calc.Cnt < sample {
			x := h.RNG.Edge(rtp)
			h.Add(float64(x))
		}
		h.Graph(100, 30)
	}
	if true {
		ƛ := par(2, 4)
		h.Title = fmt.Sprintf("Poisson distribution (ƛ = %v)", ƛ)
		h.Reset()
		for h.Calc.Cnt < sample {
			x := h.RNG.Poisson(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 50)
	}
	if true {
		ɑ1 := par(0.05, 0.15) * 2
		ɑ2 := (5 - ɑ1) / 2
		h.Title = fmt.Sprintf("Hermite distribution (ɑ1 = %v, ɑ2 = %v)", ɑ1, ɑ2)
		h.Reset()
		for h.Calc.Cnt < sample {
			x := h.RNG.Hermite(ɑ1, ɑ2)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		μ1, μ2 := par(1.6, 3), par(1.5, 2.9)
		h.Title = fmt.Sprintf("Skellam distribution (μ1 = %v, μ2 = %v)", μ1, μ2)
		h.Reset()
		for h.Calc.Cnt < sample {
			x := h.RNG.Skellam(μ1, μ2)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		p := par(0.4, 0.6)
		h.Title = fmt.Sprintf("Geometric distribution (p = %v)", p)
		h.Reset()
		for h.Calc.Cnt < sample {
			x := h.RNG.Geometric(p)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		n, p := h.RNG.Int(30, 70), par(0.1, 0.9, 10)
		h.Title = fmt.Sprintf("Binomial distribution (n = %d, p = %v)", n, p)
		h.Reset()
		for h.Calc.Cnt < sample {
			x := h.RNG.Binomial(n, p)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		a, b := 0., 20.
		mode := par(a+1, b-1, 1)
		h.Title = fmt.Sprintf("Triangular distribution (a = %v, b = %v, mode = %v)", a, b, mode)
		h.Reset()
		h.Scale(1, 1)
		for h.Calc.Cnt < sample {
			x := h.RNG.Triangular(a, b, mode)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if false {
		h.Title = "3 dice throw sum"
		h.Reset()
		for h.Calc.Cnt < sample {
			x := h.RNG.Int(1, 6) + h.RNG.Int(1, 6) + h.RNG.Int(1, 6)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if false {
		h.Title = "Sic-Bo"
		h.Reset()
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
