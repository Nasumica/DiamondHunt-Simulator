package main

import (
	"DHSimulator/rng"
	"fmt"
	"time"
)

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

type StopWatch struct {
	Begin time.Time
	Since float64
}

func (sw *StopWatch) Start() {
	sw.Begin = time.Now()
}

func (sw *StopWatch) Eplased(n int) (float64, float64) {
	sw.Since = time.Since(sw.Begin).Seconds()
	return sw.Since, float64(n) / sw.Since
}

func CalcProb(w int) (prob []float64) {
	if w >= 0 {
		prob = make([]float64, w+1)
		prob[w] = 1
		for h := range prob {
			p := rng.HypGeomDist(h, w, 13, 52)
			r := p
			for d := range prob {
				q, n := r, h+d
				if n < w {
					q = p * rng.NegHypGeomDist(d, h+1, 13-h, 52-w)
				} else if d < w {
					q = p * rng.HypGeomDist(d, w, 13-h, 52-w)
				}
				r -= q
				if n < w {
					prob[n] += q
					prob[w] -= q
				} else {
					n = w
				}
				fmt.Printf("%d   %d  %d   %12.9f%%\n", n, h, d, 100*q)
			}
		}
	}
	return
}

func compress(n int) (p float64) {
	for h := 0; h <= n; h++ {
		d := n - h
		p += rng.Binomial(13, h) * rng.Binomial(39, 4-h) * rng.Binomial(47-n, 13-n) * rng.Binomial(n, d) / rng.Binomial(48, 13-h)
	}
	p /= 270725
	return
}

func ShowDiamHuntProb() {
	prob := CalcProb(4)
	for i, p := range prob {
		p = compress(i)
		fmt.Printf("%d   %12.5f%%  %10.2f\n", i, 100*p, 1/p)
	}
	fmt.Println()
}

func main() {
	// ShowDiamHuntProb()
	var sw StopWatch
	sw.Start()
	fmt.Println()
	million := 1000 * 1000
	iter := 1000 * million

	// Strategy = SwapCourt
	// Strategy = NoStrategy
	// Strategy = RiskOne
	// Strategy = NoRisk
	Strategy = NewRisk
	DiamondHunt(iter)

	elapsed, speed := sw.Eplased(iter)
	fmt.Printf("%d games,  elapsed = %.3f\",  speed = %.0f games / s\n", iter, elapsed, speed)
}
