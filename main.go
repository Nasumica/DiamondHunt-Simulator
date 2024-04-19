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

func main() {
	var sw StopWatch
	sw.Start()
	fmt.Println()
	iter := 109750000
	DiamondHunt(iter)
	elapsed, speed := sw.Eplased(iter)
	fmt.Printf("%d games,  elapsed = %.3f\",  speed = %.0f games / s\n", iter, elapsed, speed)
}
