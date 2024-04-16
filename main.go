package main

import "fmt"

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

func main() {
	var scr Screen
	for i := 1; i <= 10; i++ {
		fmt.Println(scr.Play())
	}
	SpeedTest(1000000)
}
