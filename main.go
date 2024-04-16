package main

import (
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

func SpeedTest(n int) {
	var sw StopWatch
	sw.Start()
	var scr Screen
	var swap, diam, win StatCalc
	cat := [5]int{}
	chart := [5][5]int{}
	for i := 0; i < n; i++ {
		d, s := scr.Play()
		swap.Int(s)
		k := 0
		for _, c := range d {
			if c.Suit == DiamondSuit {
				k++
			}
		}
		chart[scr.Open][k]++
		if k < 3 {
			win.Int(0)
		} else {
			win.Int(100)
		}
		diam.Int(k)
		cat[k]++
	}
	elapsed, speed := sw.Eplased(n)
	fmt.Printf("%d games,  %.0f swaps,  elapsed = %.3f\",  speed = %.0f deals / s\n", n, swap.Sum, elapsed, speed)
}

func Chart(runs float64) {
	axis := [5]float64{}
	chart := [5][5]float64{}
	for h := 0; h <= 4; h++ {
		hprob := HypGeomDist(h, 4, 13, 52)
		axis[h] = hprob
		for d := 0; d <= 4; d++ {
			dprob := HypGeomDist(d, 4, 13-h, 52-4)
			m := h + d
			if m > 4 {
				m = 4
			}
			chart[h][m] += dprob
			// fmt.Printf("%d  %d  %8.5f  %8.5f\n", h, d, 100*hprob, 100*dprob)
		}

	}
	for h := 0; h <= 4; h++ {
		fmt.Println()
		a := axis[h]
		x := runs * a
		for d := 0; d <= 4; d++ {
			c := chart[h][d]
			t := a * c
			y := runs * t
			fmt.Printf("%d  %d  %9.5f%%  %9.5f%%  %9.5f%%  %10.0f  %10.0f\n", h, d, 100*a, 100*t, 100*c, x, y)
		}
	}
}

func main() {
	var scr Screen
	for i := 1; i <= 10; i++ {
		fmt.Println(scr.Play())
	}
	// Chart(35000000)
	SpeedTest(1 * 1000000)
}
