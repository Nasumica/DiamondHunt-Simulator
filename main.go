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
	var swap, diam StatCalc
	cat := [5]int{}
	opens := [5]int{}
	chart := [5][5]int{}
	for i := 0; i < n; i++ {
		r := scr.Play(1)
		k, s := r.Count, r.Swaps
		swap.Int(s)
		opens[scr.Open]++
		chart[scr.Open][k]++
		diam.Int(k)
		cat[k]++
	}
	fmt.Printf("random simulation %d deals\n", n)
	for h, o := range opens {
		fmt.Println()
		hp := float64(o) / float64(n)
		fmt.Printf("%d      %9.5f%%              %10d\n", h, 100*hp, o)
		for d, c := range chart[h] {
			dp := float64(c) / float64(o)
			tp := float64(c) / float64(n)
			// fmt.Printf("%23s", "")
			fmt.Printf("    %d  %9.5f%%  %9.5f%%  %10d", d, 100*tp, 100*dp, c)
			fmt.Println()
		}
	}
	fmt.Println()
	elapsed, speed := sw.Eplased(n)
	fmt.Printf("%d games,  %.0f swaps,  elapsed = %.3f\",  speed = %.0f deals / s\n", n, swap.Sum, elapsed, speed)
}

func main() {
	// var scr Screen
	// for i := 1; i <= 10; i++ {
	// 	fmt.Println(scr.Play())
	// }
	// Chart(35000000)
	// fmt.Println()
	// SpeedTest(10 * 100000)
	// fmt.Println()
}
