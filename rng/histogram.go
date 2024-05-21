package rng

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Histogram struct {
	Title      string      // title
	Data       map[int]int // data
	Min, Max   int         // range
	Mode, Peak int         // mode
	RNG        LCPRNG      // entropy
	Calc       StatCalc    // calculator
	ox, oy     float64     // origin
	dx, dy     float64     // delta
	k          float64     // slope
	f          []func(x float64) float64
	E          struct{ μ, σ float64 }
	watch      time.Time
}

func (h *Histogram) Reset(f ...func(x float64) float64) {
	h.RNG.Randomize()
	h.Calc.Reset()
	h.Data = map[int]int{}
	h.f = f
	h.Scale(1, 1)
	h.Mode, h.Peak = 0, 0
	h.E.μ, h.E.σ = 0, 0
	h.watch = time.Now()
}

func (h *Histogram) Eplased() (float64, float64) {
	s := time.Since(h.watch).Seconds()
	return s, float64(h.Calc.Cnt) / s
}

// # Define line (x1, y1)--(x2, y2)
func (h *Histogram) Linear(x1, y1, x2, y2 float64) {
	if x1 == x2 || y1 == y2 {
		h.Scale(1, 1)
	} else {
		h.dx = x2 - x1
		h.dy = y2 - y1
		h.k = h.dy / h.dx
		h.ox = x1
		h.oy = y1
	}
}

func (h *Histogram) Scale(dx, dy float64) {
	h.Linear(0, 0, dx, dy)
}

func (h *Histogram) Y(x float64) float64 {
	x -= h.ox
	y := x * h.dy / h.dx
	y += h.oy
	return y
}

func (h *Histogram) X(y float64) float64 {
	y -= h.oy
	x := y * h.dx / h.dy
	x += h.ox
	return x
}

// # Quantize value
func (h *Histogram) N(x float64) int {
	y := h.Y(x)
	if len(h.f) == 0 {
		y = math.Round(y)
	} else {
		for _, f := range h.f {
			y = f(y)
		}
	}
	return int(y)
}

func (h *Histogram) Add(x float64) {
	h.Calc.Add(x)
	n := h.N(x)
	c := h.Data[n] + 1
	h.Data[n] = c
	if h.Calc.Cnt == 1 {
		h.Min, h.Max = n, n
	} else if h.Min > n {
		h.Min = n
	} else if h.Max < n {
		h.Max = n
	}
	if h.Peak < c {
		h.Mode = n
		h.Peak = c
	}
}

func (h *Histogram) Graph(width int, limit int, flags ...bool) {
	flag := func(i int) bool {
		return i < len(flags) && flags[i]
	}
	iif := func(c bool, t, f any) any {
		if c {
			return t
		} else {
			return f
		}
	}
	trim := func(s string, repl ...string) string {
		r := ""
		if len(repl) > 0 {
			r = repl[0]
		}
		l := len(s) - 1
		for ; s[l] == '0'; l-- {
			s = s[:l] + r + s[l+1:]
		}
		if s[l] == '.' {
			s = s[:l] + r + s[l+1:]
		}
		return s
	}
	str := func(a float64) string {
		if math.IsInf(a, 1) {
			return "∞"
		} else if math.IsNaN(a) {
			return "undefined"
		}
		return trim(fmt.Sprintf("%.2f", a))
	}
	elapsed, speed := h.Eplased()
	cumul, nozero, pmf := flag(0), flag(1), h.Calc.Cnt == h.Calc.Int
	df := iif(cumul, "CDF", iif(pmf, "PMF", "PDF"))
	fmt.Println()
	fmt.Printf("%s %s:  [%v, %v]  μ = %v  σ = %v",
		h.Title, df, str(h.Calc.Min), str(h.Calc.Max),
		str(h.Calc.Avg), str(h.Calc.Dev))
	if h.E.μ != 0 || h.E.σ != 0 {
		fmt.Printf("  (expected = %v ± %v)", str(h.E.μ), str(h.E.σ))
	}
	freq := int64(math.Round(speed))
	fmt.Println()
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
	c, bottom := 0., h.Peak
	for i, d := range h.Data {
		if i < lo {
			c += float64(d)
		}
		if bottom > d {
			bottom = d
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
			b := strings.Repeat("╶", w)
			y := n / t
			if !cumul {
				y *= h.k
			}
			v := ""
			if pmf {
				v = str(100*y) + "%"
			} else {
				if y >= 1 {
					v = trim(fmt.Sprintf("%.3f", y))
				} else {
					v = trim(fmt.Sprintf("%.5f", y))
				}
			}
			x := h.X(float64(i))
			u := trim(fmt.Sprintf("%10.3f", x), " ")
			mark := " "
			if d == h.Peak {
				mark = "►"
			} else if d > 0 && d == bottom && !pmf {
				mark = "◂"
			}
			fmt.Printf("%v  %s", u, mark)
			fmt.Printf("%s%s  %v  %.0f", iif(x == math.Floor(x), "┼", "│"), b, v, n)
			fmt.Println()
		}
	}
	if true {
		fmt.Printf("%d randoms  t = %.3f\"  f = %d.%.6d M / s", h.Calc.Cnt, elapsed, freq/1000000, freq%1000000)
		fmt.Println()
	}
}

func (h *Histogram) StressTest(sample int) *Histogram {
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
		ɑ := par(0.5, 4)
		β := par(0.9, 1.1, 4)
		h.Title = fmt.Sprintf("Gamma distribution (ɑ = %v, β = %v)", ɑ, β)
		h.Reset()
		h.E.μ, h.E.σ = ɑ/β, math.Sqrt(ɑ)/β
		h.Scale(1, 5)
		for h.Calc.Cnt < sample {
			x := h.RNG.Gamma(ɑ, β)
			h.Add(float64(x))
		}
		h.Graph(100, 25)
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
		s := ɑ + β
		h.E.μ = ɑ / s
		h.E.σ = math.Sqrt(ɑ*β/(s+1)) / s
		h.Scale(1, 25)
		for h.Calc.Cnt < sample {
			x := h.RNG.Beta(ɑ, β)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		ξ, ω, ɑ := 0., 1., par(-5, 5, 10)
		h.Title = fmt.Sprintf("Skew-normal distribution (ξ = %v, ω = %v, ɑ = %v)", ξ, ω, ɑ)
		h.Reset()
		δ := ɑ / math.Hypot(ɑ, 1)
		h.E.μ = ξ + ω*δ*math.Sqrt(2/math.Pi)
		h.E.σ = ω * math.Sqrt(1-δ*δ*2/math.Pi)
		h.Scale(1, 10)
		for h.Calc.Cnt < sample {
			x := h.RNG.SkewNormal(ξ, ω, ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 20)
	}
	if true {
		ν := par(0.9, 9, 10)
		if h.RNG.Choose(100, 5) {
			ν = math.Inf(1)
		}
		h.Title = fmt.Sprintf("Student's t distribution (ν = %v)", ν)
		h.Reset()
		if math.IsInf(ν, 1) {
			h.E.σ = 1
		} else if ν > 2 {
			h.E.σ = math.Sqrt(ν / (ν - 2))
		} else if ν > 1 {
			h.E.σ = math.Inf(1)
		} else {
			h.E.σ = math.NaN()
		}
		h.Scale(1, 5)
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
		h.E.μ = 1 / ƛ
		h.E.σ = h.E.μ
		h.Scale(1, 5)
		for h.Calc.Cnt < sample {
			x := h.RNG.Exponential(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 40)
		// h.Graph(100, 100, true)
	}
	if true {
		rtp := par(0.9, 0.96)
		h.Title = fmt.Sprintf("House edgne distribution (rtp = %v%%)", 100*rtp)
		h.Reset(math.Ceil)
		h.E.μ, h.E.σ = 1-rtp, rtp
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
		h.E.μ = ƛ
		h.E.σ = math.Sqrt(ƛ)
		for h.Calc.Cnt < sample {
			x := h.RNG.Poisson(ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 50)
	}
	if true {
		ɑ1 := par(0.05, 0.15) * 2
		ɑ2 := math.Round((par(4, 6, 1)-ɑ1)/2*100) / 100
		h.Title = fmt.Sprintf("Hermite distribution (ɑ1 = %v, ɑ2 = %v)", ɑ1, ɑ2)
		h.Reset()
		h.E.μ = ɑ1 + 2*ɑ2
		h.E.σ = math.Sqrt(ɑ1 + 4*ɑ2)
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
		h.E.μ = μ1 - μ2
		h.E.σ = math.Sqrt(μ1 + μ2)
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
		q := 1 - p
		h.E.μ, h.E.σ = q/p, math.Sqrt(q)/p
		for h.Calc.Cnt < sample {
			x := h.RNG.Geometric(p)
			h.Add(float64(x))
		}
		h.Graph(100, 30)
	}
	if true {
		n, p := h.RNG.Int(30, 70), par(0.1, 0.9, 10)
		h.Title = fmt.Sprintf("Binomial distribution (n = %d, p = %v)", n, p)
		h.Reset()
		h.E.μ = float64(n) * p
		h.E.σ = math.Sqrt(h.E.μ * (1 - p))
		for h.Calc.Cnt < sample {
			x := h.RNG.Binomial(n, p)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		k := h.RNG.Int(2, 6)
		h.Title = fmt.Sprintf("χ² distribution (%d degrees of freedom)", k)
		h.Reset()
		h.Scale(1, 2)
		h.E.μ = float64(k)
		h.E.σ = math.Sqrt(2 * h.E.μ)
		for h.Calc.Cnt < sample {
			x := h.RNG.ChiSquared(k)
			h.Add(float64(x))
		}
		h.Graph(100, 30)
	}
	if true {
		a, b := 0., 20.
		c := par(a+1, b-1, 1)
		h.Title = fmt.Sprintf("Triangular distribution (a = %v, b = %v, mode = %v)", a, b, c)
		h.Reset()
		h.E.μ = (a + b + c) / 3
		h.E.σ = math.Sqrt((a*(a-b)+b*(b-c)+c*(c-a))/2) / 3
		h.Scale(1, 1)
		for h.Calc.Cnt < sample {
			x := h.RNG.Triangular(a, b, c)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if false {
		xm, ɑ := 1., par(1, 3)
		h.Title = fmt.Sprintf("Pareto distribution (xm = %v, ɑ = %v)", xm, ɑ)
		h.Reset(math.Floor)
		h.Scale(1, 20)
		for h.Calc.Cnt < sample {
			x := h.RNG.Pareto(xm, ɑ)
			h.Add(float64(x))
		}
		h.Graph(100, 30)
	}
	if false {
		n, a, b := h.RNG.Int(2, 10), 0., 1.
		h.Title = fmt.Sprintf("Bates distribution (n = %v, a = %v, b = %v)", n, a, b)
		h.Reset()
		h.Scale(1, 20)
		for h.Calc.Cnt < sample {
			x := h.RNG.Bates(n, a, b)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		m := 1
		n := h.RNG.Int(5, 20)
		h.Title = fmt.Sprintf("Benford law distribution (m = %v, n = %v)", m, n)
		h.Reset()
		a, b := LogGG(n+1, true)
		h.E.μ = float64(n) - a
		h.E.σ = math.Sqrt(a - a*a + 2*b)
		for h.Calc.Cnt < sample {
			x := h.RNG.Benford(m, n)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		μ, σ := par(1, 5, 1), 2.
		ƛ := μ * μ * μ / σ / σ
		h.Title = fmt.Sprintf("Inverse Gaussian (Wald) distribution (μ = %v, ƛ = %v)", μ, ƛ)
		h.Reset()
		h.E.μ, h.E.σ = μ, σ
		h.Scale(1, 4)
		for h.Calc.Cnt < sample {
			x := h.RNG.Wald(μ, ƛ)
			h.Add(float64(x))
		}
		h.Graph(100, 20)
	}
	if true {
		n := 10
		ɑ := par(0.5, 2.5, 2)
		β := par(0.5, 2.5, 2)
		h.Title = fmt.Sprintf("Beta-binomial distribution (n = %v, ɑ = %v, β = %v)", n, ɑ, β)
		h.Reset()
		s := ɑ + β
		p := ɑ / s
		q := 1 - p
		h.E.μ = float64(n) * p
		h.E.σ = math.Sqrt(h.E.μ * q * (s + float64(n)) / (s + 1))
		for h.Calc.Cnt < sample {
			x := h.RNG.BetaBinomial(n, ɑ, β)
			h.Add(float64(x))
		}
		h.Graph(100, 100)
	}
	if true {
		x0, ɣ := 0., par(0.5, 2, 10)
		h.Title = fmt.Sprintf("Cauchy distribution (x0 = %v, ɣ = %v)", x0, ɣ)
		h.Reset()
		h.Scale(5, 1)
		h.E.μ, h.E.σ = math.NaN(), math.NaN()
		for h.Calc.Cnt < sample {
			x := h.RNG.Cauchy(x0, ɣ)
			h.Add(float64(x))
		}
		h.Graph(100, 20)
	}
	if true {
		ν, σ := par(1, 2, 10), par(0.5, 1.5, 10)
		h.Title = fmt.Sprintf("Rice distribution (ν = %v, σ = %v)", ν, σ)
		h.Reset()
		h.Scale(1, 4)
		for h.Calc.Cnt < sample {
			x := h.RNG.Rice(ν, σ)
			h.Add(float64(x))
		}
		h.Graph(100, 20)
	}
	if true {
		balls, draw, max := 80, 20, 10
		for play := 1; play <= max; play++ {
			h.Title = fmt.Sprintf("KENO distribution (play = %v, draw = %v, balls = %v)", play, draw, balls)
			h.Reset()
			p := float64(draw) / float64(balls)
			q := 1 - p
			h.E.μ = float64(play) * p
			h.E.σ = math.Sqrt(h.E.μ * q * float64(balls-play) / float64(balls-1))
			for h.Calc.Cnt < sample {
				x := h.RNG.HyperGeometric(draw, play, balls)
				h.Add(float64(x))
			}
			h.Graph(100, play+1)
		}
	}
	if true {
		wins, balls := 6, 49
		for play := wins; play <= wins+4; play++ {
			fail := balls - play
			h.Title = fmt.Sprintf("Lucky 6 distribution (wins = %v, play = %v, balls= %v)", wins, play, balls)
			h.Reset()
			p := float64(wins) / float64(play+1)
			q := 1 - p
			h.E.μ = float64(fail) * p
			h.E.σ = math.Sqrt(h.E.μ * q * float64(balls+1) / float64(play+2))
			h.E.μ += float64(wins)
			for h.Calc.Cnt < sample {
				x := h.RNG.NegHyperGeometric(wins, fail, balls) + wins
				h.Add(float64(x))
			}
			h.Graph(100, 100)
		}
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
	if true {
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
		h.Graph(60, 0, false, true)
	}
	return h
}

func Slicke() {
	var h Histogram
	h.Reset()
	h.Title = "Slicke"

	iter := 100000
	album := 728
	fale := 200
	imam := album - fale
	kesica := 6
	kupujem := 10

	s := make([]int, album)
	t := make([]int, album)
	{
		z := h.RNG.Combination(len(t), imam)
		for _, j := range z {
			t[j] = 1
		}
	}

	for h.Calc.Cnt < iter {
		copy(s, t)
		k := 0
		for i := 0; i < kupujem; i++ {
			x := h.RNG.Combination(album, kesica)
			for _, n := range x {
				if s[n] == 0 {
					k++
				}
				s[n]++
			}
		}
		h.Add(float64(k))
	}
	h.Graph(100, 100, true)

	n := kesica * kupujem
	p := h.Calc.Avg / float64(n)
	h.Reset()
	for h.Calc.Cnt < iter {
		x := -1
		for x < 0 {
			x = h.RNG.Binomial(n, p)
		}
		h.Add(float64(x))
	}
	h.Graph(100, 100)
}

func AlgP(n int) {
	// good
	a := make([]byte, 2*n+2)
	a[1] = ')'
	for k := 1; k <= n; k++ {
		a[2*k] = '('
		a[2*k+1] = ')'
	}
	i := 0
	for {
		m := 2 * n
		k := m
		for {
			i++
			fmt.Println(i, string(a[2:]))
			a[m] = ')'
			m--
			if a[m] == '(' {
				break
			}
			a[m] = '('
		}

		for a[m] == '(' {
			a[m] = ')'
			a[k] = '('
			m -= 1
			k -= 2
		}
		if m == 1 {
			break
		}
		a[m] = '('
	}
}

func init() {
	new(Histogram).StressTest(1000000)
	// Slicke()
	// AlgP(2)
}
