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

func (sw *StopWatch) Eplased() float64 {
	sw.Since = time.Since(sw.Begin).Seconds()
	return sw.Since
}

func SpeedTest(n int) {
	var sw StopWatch
	sw.Start()
	var scr Screen
	var swap, diam StatCalc
	cat := [5]int{}
	for i := 0; i < n; i++ {
		d, s := scr.Play()
		swap.Int(s)
		k := 0
		for _, c := range d {
			if c.Suit == DiamondSuit {
				k++
			}
		}
		diam.Int(k)
		cat[k]++
	}
	elapsed := sw.Eplased()
	speed := float64(n) / elapsed
	fmt.Printf("%d games,  %.0f swaps,  elapsed = %.3f\",  speed = %.0f deals / s\n", n, swap.Sum, elapsed, speed)
}

func main() {
	var scr Screen
	for i := 1; i <= 10; i++ {
		fmt.Println(scr.Play())
	}
	SpeedTest(10000000)
}
