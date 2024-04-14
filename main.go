package main

import (
	"DHSimulator/rng"
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

var Deck rng.Deck

func init() {
	Deck.Init()
}

func main() {
	fmt.Println(Deck.Cards)
}
