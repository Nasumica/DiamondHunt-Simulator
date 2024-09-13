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
				q := r
				if d+h < w {
					q = p * rng.NegHypGeomeDist(d, h+1, 13-h, 52-w)
				} else if d < w {
					q = p * rng.HypGeomDist(d, w, 13-h, 52-w)
				}
				r -= q
				if n := h + d; n < w {
					prob[n] += q
					prob[w] -= q
				}
			}
		}
	}
	return
}

func ShowDiamHuntProb() {
	prob := CalcProb(4)
	for i, p := range prob {
		fmt.Printf("%d   %6.2f%%  %10.2f\n", i, 100*p, 1/p)
	}
	fmt.Println()
}

func main() {
	ShowDiamHuntProb()
	var sw StopWatch
	sw.Start()
	fmt.Println()
	million := 1000 * 1000
	iter := 10 * 100 * million
	DiamondHunt(iter)
	elapsed, speed := sw.Eplased(iter)
	fmt.Printf("%d games,  elapsed = %.3f\",  speed = %.0f games / s\n", iter, elapsed, speed)
}
