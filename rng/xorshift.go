package rng

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync"
	"time"
)

// # XOR shift pseudo-random numbers generator
type XORshift struct {
	seed octa       // generator seed
	dog  sync.Mutex // watchdog Šarko
}

// George Marsaglia shift-register generator.
func xorshift64(x octa) octa {
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	return x
}

// # Current seed.
func (rnd *XORshift) Seed() octa {
	rnd.dog.Lock()
	defer rnd.dog.Unlock()
	return rnd.seed
}

// # Inititalize generator with seeds or system state.
func (rnd *XORshift) Randomize(seeds ...octa) (seed octa) {
	if len(seeds) == 0 {
		seed = rnd.Next()
	} else {
		for _, s := range seeds {
			seed = xorshift64(seed) ^ s
		}
	}
	rnd.dog.Lock()         // Meni je nekako logičnije
	defer rnd.dog.Unlock() // da bude obrnuto, ... :)
	rnd.seed = seed
	return
}

// # Next random value from generator.
func (rnd *XORshift) Next() octa {
	rnd.dog.Lock()         // ... da pustim kuče dok ja radim,
	defer rnd.dog.Unlock() // a da ga vežem kad rade drugi. :)
	rnd.seed = xorshift64(rnd.seed)
	for rnd.seed == 0 {
		rnd.seed = octa(time.Now().UnixNano())
		if g, e := rand.Int(rand.Reader, new(big.Int).SetBit(new(big.Int), 64, 1)); e == nil {
			rnd.seed ^= g.Uint64() // seed from computer crypto entropy generator
		}
	}
	return rnd.seed
}

// # Next random value from generator limited to range [0, n].
//
//	μ  = n / 2
//	σ² = n · (n + 2) / 12
func (rnd *XORshift) Limited(n octa) octa {
	if n != 0 {
		if n++; n == 0 { // n = 2⁶⁴
			n = rnd.Next()
		} else { // acceptance-rejection: p(reject) = 2⁶⁴ % n / 2⁶⁴
			for f, m := -n/n+1, n; n >= m; { // f = 2⁶⁴ / n
				n = rnd.Next() / f
			}
		}
	}
	return n
}

// # Random integer in range [m, n].
func (rnd *XORshift) Int(m, n int) int {
	if m > n {
		m, n = n, m
	}
	return m + int(rnd.Limited(octa(n-m)))
}

// # Random integer in range [0, n) for positive n, -1 for n = 0, else in range [n, -1].
//
//	μ  = (n - 1) / 2
//	σ² = (n² - 1) / 12
func (rnd *XORshift) Choice(n int) int {
	if n > 1 {
		return int(rnd.Limited(octa(n - 1)))
	} else if n < 0 {
		return -int(rnd.Limited(octa(-n-1))) - 1
	} else {
		return n - 1
	}
}

// # True with probability 1/2.
//
// Coin flip decision.
func (rnd *XORshift) Flip() bool {
	const mask octa = 1 << 61 // prime number high bit
	return (rnd.Next() & mask) != 0
}

// # True with probability k/n.
func (rnd *XORshift) Choose(n, k int) bool {
	return (n > 0) && (k > 0) && (n <= k || rnd.Choice(n) < k)
}

// # Knuth shuffle (Fisher-Yates).
func (rnd *XORshift) Shuffle(a *list) {
	for j, i := 0, len(*a); i > 1; (*a)[i], (*a)[j] = (*a)[j], (*a)[i] {
		j, i = rnd.Choice(i), i-1
	}
}

// # List of n integers in range [m, m + n) in random order.
func (rnd *XORshift) Fill(m, n int) (a list) {
	if n > 0 {
		a = make(list, n)
		for i := range a {
			j := rnd.Int(0, i)
			a[i], a[j] = a[j], (m + i)
		}
	}
	return
}

// # Random permutation.
func (rnd *XORshift) Permutation(n int) list {
	return rnd.Fill(0, n)
}

// # Random combination k of n elements (quickpick).
func (rnd *XORshift) Combination(n, k int) (c list) {
	for i := 0; n > 0 && k > 0; i, n = i+1, n-1 {
		if rnd.Choose(n, k) {
			c = append(c, i)
			k--
		}
	}
	return
}

// # Random sample of k elements.
func (rnd *XORshift) Sample(k int, a list) list {
	s := rnd.Combination(len(a), k)
	for i, j := range s {
		s[i] = a[j]
	}
	return s
}

// # Hypergeometric distribution random variable.
/*
	p  = draw / size
	q  = 1 - p
	μ  = succ · p
	σ² = μ · q · (size - succ) / (size - 1)
with
	prob(hits) = HypGeomDist(hits, draw, succ, size)
*/
func (rnd *XORshift) HyperGeometric(draw, succ, size int) (hits int) {
	if size >= draw && size >= succ {
		for ; draw > 0 && succ > 0; draw-- {
			if size == succ {
				if draw < succ {
					hits += draw
				} else {
					hits += succ
				}
				break
			}
			if rnd.Choose(size, succ) {
				hits++
				succ--
			}
			size--
		}
	}
	return
}

// # Negative hypergeometric distribution random variable.
/*
	p  = miss / (size - succ + 1)
	q  = 1 - p
	μ  = succ · p
	σ² = μ · q · (size + 1) / (size - succ + 2)
*/
func (rnd *XORshift) NegHyperGeometric(miss, succ, size int) (draw int) {
	if miss <= succ && miss+succ <= size {
		for miss > 0 {
			if rnd.Choose(size, succ) {
				draw++
				succ--
			} else {
				miss--
			}
			size--
		}
	}
	return
}

// # Random list index for non-empty list else -1.
func (rnd *XORshift) Index(items list) int {
	return rnd.Choice(len(items))
}

// # Random item from non-empty list else default.
func (rnd *XORshift) Item(items list, def int) int {
	if i := rnd.Index(items); i < 0 {
		return def
	} else {
		return items[i]
	}
}

// # Random value from non-empty array else default.
func (rnd *XORshift) Value(values array, def float) float {
	if n := rnd.Choice(len(values)); n < 0 {
		return def
	} else {
		return values[n]
	}
}

// # Loaded uniform random integer.
/*
From precalculated cumulative mass function table c,
calculate loaded uniform (weighted) random integer in range [0, len(c)).

The sequence must be non-negative, non-decreasing.
The last element of the sequence must be greater than 0.
*/
func (rnd *XORshift) Loaded(c list) int {
	r := len(c) - 1
	if r > 0 { // data present and not single
		n := rnd.Choice(c[r]) // last c is "probabilityDown"
		for l := 0; l < r; {  // binary search
			m := (l + r) / 2
			if n < c[m] { // c[m] = ∑ "probabilityUp" to m
				r = m
			} else {
				l = m + 1
			}
		} // at this point l = r
	}
	return r
}

// # Weighted-uniform random integer.
//
//	ProbabilityUp[i] = w[i]
//	ProbabolityDown  = ∑ w
func (rnd *XORshift) Weighted(w list) int {
	r := len(w) - 1
	if r > 0 {
		t := 0      // total mass (probabilityDown)
		c := list{} // cumulative mass table
		for _, m := range w {
			if m < 0 {
				return -1 // no negative mass
			}
			t += m
			c = append(c, t)
		}
		if t == 0 {
			return rnd.Index(c) // random photon
		} else {
			return rnd.Loaded(c)
		}
	} else {
		return r
	}
}

// # Uniform random number in range (0, 1).
//
//	μ  = 1/2
//	σ² = 1/12
func (rnd *XORshift) Random() float {
	var n octa
	for n == 0 {
		n = rnd.Next() >> 11 // trim to 53 bits mantissa
	}
	const ε float = 0x1p-53 // ε = 2⁻⁵³
	return ε * float(n)
}

// # Random angle (0, 2π).
//
//	μ = π
//	σ = π / sqrt(3) = 1.8137993642342178505940782576422
func (rnd *XORshift) Angle() float {
	const τ float = 2 * math.Pi // τ = 2π = 0x3243F6A8885A3p-47
	return τ * rnd.Random()
}

// # Uniform random number in range (0, x).
func (rnd *XORshift) Uniform(x float) float {
	if x != 0 {
		x *= rnd.Random()
	}
	return x
}

// # Uniform random number in range (a, b).
func (rnd *XORshift) Range(a, b float) float {
	return a + rnd.Uniform(b-a)
}

// # True with probability p.
func (rnd *XORshift) Bernoulli(p float) bool {
	return (p >= 1) || (p > 0 && p > rnd.Random())
}

// # Random bit: 1 with probability p, else 0.
//
//	μ  = p
//	σ² = p  - p²
func (rnd *XORshift) Bit(p float) int {
	if rnd.Bernoulli(p) {
		return 1
	} else {
		return 0
	}
}

// # Rademacher distribution random variable {-x or x}.
//
// Random sign of the given number.
//
//	μ = 0
//	σ = x
func (rnd *XORshift) Rademacher(x float) float {
	if x != 0 && rnd.Flip() {
		x = -x
	}
	return x
}

// # Binomial distribution random variable.
//
//	μ  = n · p
//	σ² = μ  · (1 - p)
func (rnd *XORshift) Binomial(n int, p float) (b int) {
	if p <= 0 || n <= 0 {
		return 0
	} else if p >= 1 {
		return n
	}
	const limit, y = 50, 9
	x, q := float(n), 1-p
	if (n > limit) && (x*p > y*q) && (x*q > y*p) { // Central Limit Theorem
		x *= p           // μ
		q *= x           // σ²
		q = math.Sqrt(q) // σ
		for b = rnd.Discrete(x, q); (b < 0) || (b > n); {
			b = rnd.Discrete(x, q)
		}
	} else {
		for ; n > 0; n-- {
			b += rnd.Bit(p)
		}
	}
	return
}

// # Exponential distribution random variable.
//
//	μ = σ = 1 / ƛ
func (rnd *XORshift) Exponential(ƛ ...float) float {
	e := -math.Log1p(-rnd.Random()) // domain = [0, 36.7368]
	if len(ƛ) > 0 {
		e /= ƛ[0]
	}
	return e
}

// # Pascal (negative binomial) distribution random variable.
func (rnd *XORshift) Pascal(r int, p float) (n float) {
	if r <= 0 || p <= 0 {
		n = math.Inf(1)
	} else if p < 1 {
		for p = -math.Log1p(-p); r > 0; r-- {
			n += math.Floor(rnd.Exponential(p))
		}
	}
	return
}

// # Geometric distribution random variable.
/*
	q = 1 - p
	μ = q / p
	σ = sqrt(q) / p
*/
func (rnd *XORshift) Geometric(p float) float {
	return rnd.Pascal(1, p)
}

// # Rayleigh distribution random variable.
/*
	μ² + σ² = ς² + ς²
where
	μ = ς · sqrt(π/2)     = ς · 1.25331413731550025120788264240552
	σ = ς · sqrt(2 - π/2) = ς · 0.65513637756203355309393588562466
*/
func (rnd *XORshift) Rayleigh(ς float) float {
	if ς != 0 {
		ς *= math.Sqrt(2 * rnd.Exponential()) // Box-Muller transform
	}
	return ς
}

// # Arcus distribution random variable (-1, 1).
//
//	μ  = 0
//	σ² = 1/2
func (rnd *XORshift) Arcus() float {
	return math.Sin(rnd.Angle())
}

// # ArcSine distribution random variable (0, 1).
//
//	μ  = 1/2
//	σ² = 1/8
func (rnd *XORshift) ArcSine() float {
	return (rnd.Arcus() + 1) / 2
}

// # Gauss distribution random variable.
//
//	μ = 0
//	σ = 1
func (rnd *XORshift) Gauss() float {
	return rnd.Rayleigh(rnd.Arcus()) // domain = [±8.57167]
}

// # Normal distribution random variable.
func (rnd *XORshift) Normal(μ, σ float) float {
	if σ != 0 {
		σ *= rnd.Gauss()
	}
	return μ + σ
}

// # Discrete normal distribution random variable.
//
// Default quantization method = round(x)
func (rnd *XORshift) Discrete(μ, σ float, quantize ...func(x float) float) int {
	x := rnd.Normal(μ, σ)
	for _, f := range quantize {
		x = f(x)
	}
	return int(math.Round(x))
}

// # Skew-normal distribution random variable.
/*
	δ = ɑ / sqrt(ɑ² + 1) = ɑ / hypot(ɑ, 1)
	μ = ξ + ω · δ · sqrt(2 / π)
	σ = ω · sqrt(1 - δ² · 2 / π)
*/
func (rnd *XORshift) SkewNormal(ξ, ω, ɑ float) (s float) {
	const limit = 1024
	if ω != 0 {
		switch { // some special cases
		case ɑ == 0:
			s = rnd.Gauss()
		case ɑ < -limit:
			s = -math.Abs(rnd.Gauss())
		case ɑ > +limit:
			s = +math.Abs(rnd.Gauss())
		default:
			u, v := rnd.Target(1)
			if u > v {
				u, v = v, u
			}
			switch {
			case ɑ == -1:
				s = u
			case ɑ == +1:
				s = v
			default:
				s = (u*(1-ɑ) + v*(1+ɑ)) / math.Sqrt(2*(1+ɑ*ɑ))
			}
		}
		s *= ω
	}
	s += ξ
	return
}

// # Log-normal distribution random variable.
func (rnd *XORshift) LogNormal(μ, σ float) float {
	return math.Exp(rnd.Normal(μ, σ))
}

// # Exponentially modified normal distribution random variable.
func (rnd *XORshift) ExpNormal(μ, σ, ƛ float) float {
	return rnd.Normal(μ, σ) + rnd.Exponential(ƛ)
}

// # Laplace distribution random variable.
//
//	σ = b · sqrt(2) = b · 1.4142135623730950488016887242097
func (rnd *XORshift) Laplace(μ, b float) float {
	if b != 0 {
		b *= rnd.Rademacher(rnd.Exponential())
	}
	return μ + b
}

// # Gumbel distribution random variable.
/*
	γ = 0.57721566490153286060651209008240243104215933593992 (Euler-Mascheroni constant)
	μ = m + β · γ
	σ = β · π / sqrt(6) = β · 1.282549830161864095544036359671
*/
func (rnd *XORshift) Gumbel(m, β float) float {
	if β != 0 {
		β *= -math.Log(rnd.Exponential())
	}
	return m + β
}

// # Suzuki distribution random variable.
/*
	w  =  ν²
	t  = exp(2·m + w)
	μ² = t · π/2
	σ² = t · (2·exp(w) - π/2)
*/
func (rnd *XORshift) Suzuki(m, ν float) float {
	return rnd.Rayleigh(rnd.LogNormal(m, ν))
}

// # Cauchy distribution random variable.
//
//	μ = σ = undefined
func (rnd *XORshift) Cauchy(x0, ɣ float) float {
	if ɣ != 0 {
		ɣ *= math.Tan(rnd.Angle()) // due to inexact π: tan(π/2) = 16331239353195392
	}
	return x0 + ɣ
}

// # Tukey distribution random variable.
func (rnd *XORshift) Tukey(ƛ float) float {
	p := rnd.Random()
	switch ƛ {
	case 0:
		return math.Log(1/p - 1) // Logistic
	case 1:
		return 2*p - 1 // Uniform (-1, 1)
	case 2:
		return p - 0.5 // Uniform (-1/2, 1/2)
	default:
		return (math.Pow(p, ƛ) - math.Pow(1-p, ƛ)) / ƛ
	}
}

// # Logistic distribution random variable.
//
//	σ = s · π / sqrt(3) = s · 1.8137993642342178505940782576422
func (rnd *XORshift) Logistic(μ, s float) float {
	if s != 0 {
		s *= math.Log(1/rnd.Random() - 1)
	}
	return μ + s
}

// # Poisson distribution random variable.
//
//	μ = σ² = ƛ
func (rnd *XORshift) Poisson(ƛ float) (n int) {
	const limit = 256
	if ƛ > 0 {
		if ƛ < limit { // Knuth method
			ƛ = math.Exp(-ƛ)
			for p := rnd.Random(); p > ƛ; n++ {
				p *= rnd.Random()
			}
		} else { // Variance stabilizing
			const adj = 0.25 // adjustment (empirical = variance)
			ƛ = rnd.Normal(math.Sqrt(ƛ-adj), 0.5)
			n = int(math.Round(ƛ * ƛ))
		}
	}
	return
}

// # Skellam distribution random variable.
//
//	μ  = μ₁ - μ₂
//	σ² = μ₁ + μ₂
func (rnd *XORshift) Skellam(μ1, μ2 float) (n int) {
	if μ1 >= 0 && μ2 >= 0 {
		n = rnd.Poisson(μ1) - rnd.Poisson(μ2)
	}
	return
}

// # Hermite distribution random variable.
//
//	μ  = ɑ₁ + 2·ɑ₂
//	σ² = ɑ₁ + 4·ɑ₂
func (rnd *XORshift) Hermite(ɑ1, ɑ2 float) (n int) {
	if ɑ1 >= 0 && ɑ2 >= 0 {
		n = rnd.Poisson(ɑ1) + 2*rnd.Poisson(ɑ2)
	}
	return
}

// # χ² distribution random variable with k degrees of freedom.
/*
Sum of k squared Gauss randoms.
	μ  = k
	σ² = 2·k
*/
func (rnd *XORshift) ChiSquared(k int) (x float) {
	const limit = 256
	if k < limit {
		if k > 1 {
			for x = 1; k > 1; k -= 2 {
				x -= rnd.Uniform(x)
			}
			x = -2 * math.Log(x)
		}
		if k > 0 {
			x += rnd.Exponential() * (rnd.Arcus() + 1)
		}
	} else { // Central Limit Theorem
		x = float(k)
		x = rnd.Normal(x, math.Sqrt(2*x))
	}
	return
}

// # χ distribution random variable with k degrees of freedom.
//
// Length of k-dimensional vector with Gauss random coordinates.
func (rnd *XORshift) Chi(k int) (x float) {
	switch { // speed up by special cases
	case k == 1:
		x = math.Abs(rnd.Gauss()) // Half-Normal distribution
	case k == 2:
		x = rnd.Rayleigh(1) // Rayleigh distribution
	case k >= 3:
		x = math.Sqrt(rnd.ChiSquared(k))
	}
	return
}

// # Gamma distribution random variable.
/*
Sum of α Exponential(β) randoms.
	μ  = ɑ / β
	σ² = ɑ / β²
*/
func (rnd *XORshift) Gamma(ɑ float, β ...float) (g float) {
	if ɑ > 0 {
		t, a := math.Modf(2 * ɑ) // trunc & frac
		if a > 0 {
			a /= 2
			// Ahrens-Dieter acceptance-rejection method
			for u, r, l, f := a+math.E, 1/a, a-1, false; !f; {
				if rnd.Uniform(u) < math.E {
					g = math.Pow(rnd.Random(), r)
					f = rnd.Bernoulli(math.Exp(-g))
				} else {
					g = 1 + rnd.Exponential()
					f = rnd.Bernoulli(math.Pow(g, l))
				}
			}
		}
		g += rnd.ChiSquared(int(t)) / 2 // add integer part
		if len(β) > 0 {
			g /= β[0]
		}
	}
	return
}

// # Beta distribution random variable.
/*
	s = ɑ + β
	μ = ɑ / s
	σ = sqrt(ɑ · β / (s + 1)) / s
*/
func (rnd *XORshift) Beta(ɑ, β float) (b float) {
	if ɑ > 0 && β > 0 {
		switch { // some special cases
		case ɑ == 1 && β == 1:
			b = rnd.Random() // uniform
		case ɑ == 1:
			b = 1 - math.Pow(1-rnd.Random(), 1/β) // maximum of β uniform randoms
		case β == 1:
			b = math.Pow(rnd.Random(), 1/ɑ) // minimum of α uniform randoms
		case ɑ == 0.5 && β == 0.5:
			b = rnd.ArcSine()
		default:
			b = rnd.Gamma(ɑ)
			if b != 0 {
				b /= b + rnd.Gamma(β)
			}
		}
	}
	return
}

// # Beta-prime distribution random variable.
/*
	μ = ɑ / (β - 1)
	σ = sqrt(ɑ · (ɑ + β - 1) / (β - 2)) / (β - 1)
*/
func (rnd *XORshift) BetaPrime(ɑ, β float) (b float) {
	b = rnd.Beta(ɑ, β)
	if b != 0 && b != 1 {
		b /= 1 - b // Gamma(α) / Gamma(β)
	}
	return
}

// # Beta-binomial distribution random variable.
/*
	s  = ɑ + β
	p  = ɑ / s
	q  = 1 - p
	μ  = n · p
	σ² = μ · q · (s + n) / (s + 1)
*/
func (rnd *XORshift) BetaBinomial(n int, ɑ, β float) (b int) {
	if n > 0 && ɑ >= 0 && β >= 0 {
		switch { // some special cases
		case ɑ == 0:
			b = 0
		case β == 0:
			b = n
		case ɑ == 1 && β == 1:
			b = rnd.Int(0, n) // uniform
		case n == 1:
			b = rnd.Bit(ɑ / (ɑ + β))
		default:
			b = rnd.Binomial(n, rnd.Beta(ɑ, β))
		}
	}
	return
}

// # Pólya distribution random variable.
/*
	μ  = n · p
	σ² = μ · (1 - p) · (ɑ · n + 1) / (ɑ + 1)
*/
func (rnd *XORshift) Polya(n int, p, ɑ float) int {
	if n > 0 && p > 0 && ɑ > 0 {
		return rnd.BetaBinomial(n, p/ɑ, (1-p)/ɑ)
	} else {
		return 0
	}
}

// # Erlang distribution random variable.
//
//	μ  = k / ƛ
//	σ² = k / ƛ²
func (rnd *XORshift) Erlang(k int, ƛ float) float {
	return rnd.Gamma(float(k), ƛ)
}

// # Inverse-Gamma distribution random variable.
func (rnd *XORshift) InvGamma(ɑ, β float) float {
	if ɑ > 0 && β > 0 {
		return β / rnd.Gamma(ɑ)
	} else {
		return 0
	}
}

// # Student's t-distribution random variable with ν degrees of freedom.
/*
Normal distribution with
	μ  = 0
	σ² = ν / χ²(ν) = v / (v - 2) for v > 2
For ν -> ∞, σ -> 1
	StudentsT(∞) = Normal(0, 1) = Gauss()
*/
func (rnd *XORshift) StudentsT(ν float) (t float) {
	if ν > 0 {
		t = rnd.Gauss()
		if !math.IsInf(ν, 1) {
			ν /= 2
			t *= math.Sqrt(ν / rnd.Gamma(ν))
		}
	}
	return
}

// # Snedecor's F-ratio distribution random variable.
//
//	f = (χ²(d₁) / d₁) / (χ²(d₂) / d₂)
func (rnd *XORshift) SnedecorsF(d1, d2 float) (f float) {
	if d1 > 0 && d2 > 0 {
		f = rnd.BetaPrime(d1/2, d2/2) * d2 / d1
	}
	return
}

// # Fisher Z-distribution random variable.
func (rnd *XORshift) FisherZ(d1, d2 float) (f float) {
	if f = rnd.SnedecorsF(d1, d2); f > 0 {
		f = math.Log(f) / 2
	}
	return
}

// # Dirichlet distribution random array which sum is equal to 1.
/*
Weighted random cuts.
	s  = Σ ɑ
	μᵢ = ɑᵢ / s
	σᵢ = sqrt(ɑᵢ · (s - ɑᵢ) / (s + 1)) / s
*/
func (rnd *XORshift) Dirichlet(ɑ ...float) (d array) {
	if n := len(ɑ); n > 0 {
		d = make(array, n)
		if n == 1 {
			d[0] = 1
		} else {
			var s float
			for i, a := range ɑ {
				d[i] = rnd.Gamma(a)
				s += d[i]
			}
			if s == 0 {
				for i := range d {
					d[i] = rnd.Exponential()
					s += d[i]
				}
			}
			for i := range d {
				d[i] /= s
			}
		}
	}
	return
}

// # Nakagami distribution random variable.
/*
	μ² = Ω · (Γ(m + 1/2) / Γ(m))² / m
	σ² = Ω - μ²
*/
func (rnd *XORshift) Nakagami(m, Ω float) float {
	if m < 0.5 || Ω <= 0 {
		return 0
	}
	return math.Sqrt(rnd.Gamma(m) * Ω / m)
}

// # Maxwell–Boltzmann distribution random variable (3 degrees of freedom).
/*
Random speed of particle (m/s) in 3D space where
	M = molar mass (g/mol)
	T = temperature (°C)
For example, use
	Maxwellian(32, 25)
for
	M (oxygen molecule O₂) = 16 · 2 = 32 g/mol
	T (room temperature) = 25°C
*/
func (rnd *XORshift) Maxwellian(M, T float) (v float) {
	const (
		O = -273.15       // Absolute zero (°C)
		c = 299792458     // Speed of light (m/s)
		N = 6.02214076e23 // Avogadro constant (mol⁻¹)
		k = 1.380649e-23  // Boltzmann constant (J/K)
		R = N * k * 1000  // Ideal gas constant (scaled kg/g)
	)
	T -= O // temperature in Kelvin
	switch {
	case M < 0 || T < 0: // unknown (yet)
		v = math.NaN()
	case M == 0: // photon
		v = c
	case T == 0: // absolute zero
		v = 0
	default: // Brownian motion
		v = math.Min(math.Sqrt(rnd.ChiSquared(3)*R*T/M), c)
	}
	return
}

// # Inverse Gausian distribution random variable.
//
//	σ² = μ³ / ƛ
func (rnd *XORshift) Wald(μ, ƛ float) (w float) {
	if μ > 0 && ƛ > 0 {
		ƛ *= 2
		w = μ * rnd.ChiSquared(1)
		w = μ * (w - math.Sqrt(w*(2*ƛ+w)))
		w = μ + w/ƛ
		if rnd.Uniform(μ+w) > μ {
			w = μ * μ / w
		}
	}
	return
}

// # Pareto distribution random variable.
/*
	x  = xm / (ɑ - 1)
	μ  = x · ɑ
	σ² = x · μ / (ɑ - 2)
*/
func (rnd *XORshift) Pareto(xm, ɑ float) (p float) {
	if xm > 0 && ɑ > 0 {
		p = xm * math.Exp(rnd.Exponential(ɑ))
	}
	return
}

// # Lomax distribution random variable.
/*
	μ  = ƛ  / (ɑ - 1)
	σ² = μ² / (ɑ - 2)
*/
func (rnd *XORshift) Lomax(ɑ, ƛ float) float {
	return rnd.Pareto(ƛ, ɑ) - ƛ
}

// # Weibull distribution random variable.
/*
	μ  = ƛ  · Γ(1 + 1/k)
	σ² = ƛ² · Γ(1 + 2/k) - μ²
*/
func (rnd *XORshift) Weibull(ƛ, k float) (w float) {
	if ƛ > 0 && k > 0 {
		w = ƛ * math.Pow(rnd.Exponential(), 1/k)
	}
	return
}

// # Yule–Simon distribution random variable.
/*
	μ  = ρ  / (ρ - 1)
	σ² = μ² / (ρ - 2)
*/
func (rnd *XORshift) Yule(ρ float) float {
	if ρ > 0 {
		return rnd.Geometric(math.Exp(-rnd.Exponential(ρ))) + 1
	} else {
		return 0
	}
}

// # Logarithmic-uniform random variable.
func (rnd *XORshift) Logarithmic(a, b float) (l float) {
	if a > b {
		a, b = b, a
	}
	if a > 0 {
		if a == b {
			l = a
		} else {
			l = math.Exp(rnd.Range(math.Log(a), math.Log(b)))
			l = math.Min(b, math.Max(a, l))
		}
	}
	return
}

// # Benford law random integer in range [m, n].
/*
For m = 1, set
	a, b = LogGG(n + 1, true)
Then
	μ  = n - a
	σ² = a - a² + 2·b
*/
func (rnd *XORshift) Benford(m, n int) (b int) {
	if m > n {
		m, n = n, m
	}
	if m > 0 {
		if m == n {
			b = m
		} else {
			b = int(math.Exp(rnd.Range(math.Log(float(m)), math.Log(float(n)+1))))
			b = rnd.Censor(m, b, n)
		}
	}
	return
}

// # Irwin-Hall distribution random variable.
/*
Sum of n uniform randoms.
	μ  = n / 2
	σ² = n / 12
*/
func (rnd *XORshift) IrwinHall(n int) (x float) {
	const limit = 64
	if n > limit { // Central Limit Theorem
		x = float(n)
		x = rnd.Normal(x/2, math.Sqrt(x/12))
	} else {
		for ; n > 0; n-- {
			x += rnd.Random()
		}
	}
	return
}

// # Bates distribution random variable.
/*
	μ  = (a + b) / 2
	σ² = (b - a)² / (12·n)
*/
func (rnd *XORshift) Bates(n int, a, b float) float {
	if n > 0 {
		b = (b - a) * rnd.IrwinHall(n) / float(n)
		return a + b
	} else {
		return 0
	}
}

// # Triangulat distribution random variable.
/*
	c = mode
	μ = (a + b + c) / 3
	σ = sqrt((a·(a - b) + b·(b - c) + c·(c - a)) / 2) / 3
*/
func (rnd *XORshift) Triangular(a, b, mode float) (t float) {
	if a > b {
		a, b = b, a
	}
	if a <= mode && mode <= b {
		w, d := b-a, mode-a
		if x := rnd.Uniform(w); x < d {
			t = a + math.Sqrt(x*d)
		} else {
			t = b - math.Sqrt((w-x)*(b-mode))
		}
	}
	return
}

// # Sort list with random pivot.
func (rnd *XORshift) Sort(x *list) {
	const treshold = 16 // algorithm selection treshold

	type part struct {
		l, r int
	}
	q := part{0, len(*x) - 1}
	queue := []part{q} // partition queue

	var l, r, p int // left, right and pivot

	qsort := func() { // Quick Sort by Sir C. A. R. Hoare (1960)
		l, r = q.l, q.r
		p = (*x)[rnd.Int(l, r)] // random pivot
		for l <= r {
			for (*x)[l] < p {
				l++
			}
			for p < (*x)[r] {
				r--
			}
			if l <= r {
				(*x)[l], (*x)[r] = (*x)[r], (*x)[l]
				l++
				r--
			}
		}
		if q.l < r {
			queue = append(queue, part{q.l, r})
		}
		if l < q.r {
			queue = append(queue, part{l, q.r})
		}
	}

	isort := func() { // Insertion Sort (better for short arrays)
		for r = q.l + 1; r <= q.r; r++ {
			p = (*x)[r]
			for l = r; l > q.l && (*x)[l-1] > p; l-- {
				(*x)[l] = (*x)[l-1]
			}
			(*x)[l] = p
		}
	}

	for len(queue) > 0 { // main loop
		q, queue = queue[0], queue[1:]
		if (q.r - q.l) > treshold {
			qsort()
		} else {
			isort()
		}
	}
}

// # Weighted-uniform random variation k of n elements.
/*
Calculated by race simulation standing list.
	w = tuning
	n = len(tuning)
	k = podium
*/
func (rnd *XORshift) Race(podium int, tuning list) (stand list) {
	cars := len(tuning) // number of cars

	if podium = rnd.Censor(0, podium, cars); podium == 0 { // race canceled
		return
	}

	stand = make(list, podium) // standing list
	var place int              // current place
	var finish int             // cars count
	var pos int                // positive total tunings
	var neg int                // negative total tunings
	var head, body, tail list  // cars list

	// Gentlemen, start your engines!

	for c, t := range tuning {
		switch {
		case t > 0: // good tuning
			head, pos = append(head, c), pos+t
		case t < 0: // bad tuning
			tail, neg = append(tail, c), neg-t
		default: // no tuning
			body = append(body, c)
		}
	}

	// ▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀

	race := func(car list, tune, dir int) {
		for l := len(car); l > 0 && finish < podium; l-- {
			i := 0
			if l > 1 {
				if tune == 0 { // uniform
					i = rnd.Choice(l)
				} else { // weighted
					n, t := rnd.Choice(tune), 0
					for i = -1; n >= 0; n -= t {
						i++
						t = tuning[car[i]] * dir
					}
					tune -= t
				}
			}
			if place < podium { // chequered flag
				stand[place] = car[i]
				finish++
			}
			place += dir             // next place on podium
			copy(car[i:], car[i+1:]) // remove car from track
		}
	}

	race(head, pos, 1)  // convoy head (favorites)
	race(body, 0, 1)    // convoy body (uniform)
	place = cars - 1    // backwards
	race(tail, neg, -1) // convoy tail (wrecks)

	// ▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀

	return
}

// # Weighted-uniform random permutation.
func (rnd *XORshift) Convoy(tuning list) list {
	return rnd.Race(len(tuning), tuning)
}

// # Random forest.
//
// Random string of nested parenthesis.
//
// TAOCP Vol 4a, p 453, Algorithm W.
func (rnd *XORshift) Forest(n int) (f string) {
	for p, q := n, n; q > 0; {
		if rnd.Choose((q+p)*(q-p+1), (q-p)*(q+1)) {
			f += ")"
			q--
		} else {
			f += "("
			p--
		}
	}
	return
}

// # Cut deck of cards near middle.
//
//	μ  = len(deck) / 2
//	σ² = len(deck) / 4
func (rnd *XORshift) CutDeck(deck list) (l, r list) {
	if n := len(deck); n > 0 {
		n = rnd.Binomial(n, 0.5)
		l, r = append(l, deck[:n]...), append(r, deck[n:]...)
	}
	return
}

// # Interleave cards from left and right hand.
//
// Gilbert-Shannon-Reeds model.
func (rnd *XORshift) DoveTail(l, r list) (d list) {
	i, j := len(l), len(r)
	n := i + j
	d = make(list, n)
	for n > 0 {
		if rnd.Choose(n, i) {
			n--
			i--
			d[n] = l[i]
		} else {
			n--
			j--
			d[n] = r[j]
		}
	}
	return
}

// # Riffle shuffle deck of cards.
func (rnd *XORshift) RiffleShuffle(deck *list) {
	if n := len(*deck); n > 1 {
		// by Bayer & Diaconis (n = 8 for standard deck)
		for n = int(math.Log2(float(n)) * 3 / 2); n > 0; n-- {
			(*deck) = rnd.DoveTail(rnd.CutDeck(*deck))
		}
	}
}

// # List of n random integers which sum is equal to s.
//
//	μ  = s / n
//	σ² = μ · (1 - 1 / n)
//
// Deli špil od s karata na n približno jednakih delova.
//
// Metoda određuje "koliko dinara će sakupiti svako dete",
// kada kum na "Kume, izgoreti kesa!" baci s dinara, a ispred crkve se nalazi n dece.
func (rnd *XORshift) Scatter(s, n int) (d list) {
	if n > 0 {
		d = make(list, n)
		if s != 0 {
			const l = 2 * 53 * math.Ln2
			if n > 1 && math.Abs(float(s)) < float(n-1)*l { // Bernoulli method
				for ; s > 0; s-- {
					d[rnd.Choice(n)]++
				}
				for ; s < 0; s++ {
					d[rnd.Choice(n)]--
				}
			} else { // Central Limit Theorem
				for n > 1 {
					t, c := float(s), float(n) // total & count
					n--
					d[n] = rnd.Discrete(t/c, math.Sqrt(math.Abs(t)*(c-1))/c)
					s -= d[n]
				}
				d[0] = s
			}
		}
	}
	return
}

// # Random point on circle of radius r.
func (rnd *XORshift) Circle(r float) (x, y float) {
	if r != 0 {
		y, x = math.Sincos(rnd.Angle())
		x, y = r*x, r*y
	}
	return
}

// # Uniform random point in disc of radius r.
func (rnd *XORshift) Disc(r float) (x, y float) {
	if r != 0 {
		x, y = rnd.Circle(r * math.Sqrt(rnd.Random()))
	}
	return
}

// # Shooting on target bullet position.
func (rnd *XORshift) Target(dispersion float) (float, float) {
	return rnd.Circle(rnd.Rayleigh(dispersion))
}

// # Bi-normal distribution random pair.
func (rnd *XORshift) BiNormal(μ1, μ2, σ1, σ2 float) (float, float) {
	n1, n2 := rnd.Target(1)
	return μ1 + n1*σ1, μ2 + n2*σ2
}

// # Beckmann distribution random variable.
//
// Diagonal length of a rectangle with normal random sides.
func (rnd *XORshift) Beckmann(μ1, μ2, σ1, σ2 float) float {
	return math.Hypot(rnd.BiNormal(μ1, μ2, σ1, σ2))
}

// # Rice distribution random variable.
func (rnd *XORshift) Rice(ν, σ float) float {
	if ν == 0 {
		return rnd.Rayleigh(σ)
	} else {
		x, y := rnd.Circle(ν)
		return rnd.Beckmann(x, y, σ, σ)
	}
}

// # Color pixel random dither (with γ correction).
func (rnd *XORshift) Dither(r, g, b byte, γ ...float) bool {
	const ( // CIELAB white-point
		x = 0.212671232040624
		y = 0.715159645674898
		z = 1 - (x + y)
		w = 255
	)
	var p float = 1
	if r < w || g < w || b < w {
		p = (x*float(r) + y*float(g) + z*float(b)) / w
		if 0 < p && len(γ) > 0 {
			c := γ[0] // correction
			if c < 0 {
				p = math.Pow(p, 1/(1-c))
			} else if c > 0 {
				p = math.Pow(p, 1+c)
			}
		}
	}
	return rnd.Bernoulli(p)
}

// # House edge random variable for given return-to-player.
//
//	μ = 1 - rtp
//	σ = rtp
func (rnd *XORshift) Edge(rtp float) float {
	if rtp > 0 {
		return 1 - rtp*rnd.Exponential()
	} else {
		return 1
	}
}

// # 3 dice roll.
//
// Returns random dice roll (111-666), virtue (1-56) and frequency (1, 3, 6).
//
// Ludus Clericalis, TAOCP 4b, pp 493-494.
func (rnd *XORshift) SicBo() (dice list, virtue, freq int) {
	d := dice
	for roll := rnd.Choice(216); len(d) < 3; roll /= 6 {
		d = append(d, roll%6+1)
	}
	dice = append(dice, d...) // random variation
	rnd.Sort(&d)              // random combination
	virtue = 56 - ((6-d[0])*(7-d[0])*(8-d[0])/6 + (6-d[1])*(7-d[1])/2 + (6-d[2])/1)
	switch {
	case d[0] == d[2]: // triplet
		freq = 1
	case d[0] != d[1] && d[1] != d[2]: // singles
		freq = 6
	default: // double + single
		freq = 3
	}
	return
}

// # Slot reels stop positions and grid.
func (rnd *XORshift) Slot(reels grid, height ...int) (stop list, grid grid) {
	l := len(height)
	for i, r := range reels {
		s := rnd.Index(r)
		stop = append(stop, s)
		if s >= 0 {
			r = append(r[s:], r[:s]...)
			if l > 0 {
				if h := height[i%l]; h < len(r) {
					r = r[:h]
				}
			}
		}
		grid = append(grid, r)
	}
	return
}

// # Balls mixer.
func (rnd *XORshift) Mixer(balls int) list {
	return rnd.Fill(1, balls)
}

// # Standard deck of 52 cards.
func (rnd *XORshift) Deck() list {
	return rnd.Mixer(52)
}

// # Tombola mixer.
func (rnd *XORshift) Tombola() list {
	return rnd.Mixer(90)
}

// # Bingo mixer.
func (rnd *XORshift) Bingo() list {
	return rnd.Mixer(75)
}

// # Keno mixer.
func (rnd *XORshift) Keno() list {
	return rnd.Mixer(80)
}

// # Lucky 6 mixer.
func (rnd *XORshift) Lucky6() list {
	return rnd.Mixer(49)
}

// # Censor value to range [min, nax].
//
// Used to censor other distributions.
func (rnd *XORshift) Censor(min, value, max int) int {
	if min > max {
		min, max = max, min
	}
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}
