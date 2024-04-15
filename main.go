package main

import (
	"fmt"
)

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

const (
	million         = 1000 * 1000        // million ^ 1
	billion         = million * million  // million ^ 2
	trillion        = million * billion  // million ^ 3
	quadrillion     = million * trillion // million ^ 4
	milliard        = 1000 * million
	billiard        = 1000 * billion
	trilliard       = 1000 * trillion
	quadrilliard    = 1000 * quadrillion
	usa_billion     = 1000 * million
	usa_trillion    = 1000 * usa_billion
	usa_quadrillion = 1000 * usa_trillion
)

var Dealer Deck

func init() {
	Dealer.Init()
}

func main() {
	h := Dealer.Deal(7)
	fmt.Println(h)
	fmt.Println(Dealer.Likelihood(h))
}
