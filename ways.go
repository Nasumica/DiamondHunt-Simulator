package main

import "fmt"

type symbol = int

type WaysItemNew struct {
	Symbol  symbol  `json:"s"`           // symbol
	Order   int     `json:"o,omitempty"` // order: 1 =  L2R; -1 = R2L, 0 = both
	Line    float64 `json:"l,omitempty"` // single line win amount
	Factors []int   `json:"f,omitempty"` // consecutive occurrences
	Width   int     `json:"m,omitempty"` // line width
	Wilds   int     `json:"e,omitempty"` // excluded wilds only lines
	Ways    int     `json:"w,omitempty"` // number of ways = Π factors - wilds
	Payout  float64 `json:"p,omitempty"` // payout = line * ways
}

func Ways(grid [][]symbol, wild symbol) (result []WaysItemNew) {
	width := len(grid)

	symbols := map[symbol]int{} // list of symbols
	for _, reel := range grid {
		for _, symbol := range reel {
			if symbol != wild {
				symbols[symbol]++
			}
		}
	}

	var wilds []WaysItemNew // wilds ways

	scan := func(symbol symbol) (win WaysItemNew) {
		win = WaysItemNew{Symbol: symbol, Ways: 1}
		for _, reel := range grid {
			count := 0
			for _, s := range reel {
				if s == symbol || s == wild {
					count++
				}
			}
			if count == 0 {
				break
			}
			win.Ways *= count
			win.Factors = append(win.Factors, count)
			if symbol == wild {
				wilds = append(wilds, win)
			}
		}
		if win.Width = len(win.Factors); win.Width == 0 {
			win.Ways = 0
		}
		if symbol != wild {
			if win.Width > 0 && win.Width <= len(wilds) {
				win.Wilds = wilds[win.Width-1].Ways
				win.Ways -= win.Wilds
			}
		}
		if win.Ways > 0 && win.Width <= width {
			result = append(result, win)
		}
		return
	}

	// scan symbols
	scan(wild)
	for symbol := range symbols {
		scan(symbol)
	}

	return
}

func WaysTest() {
	grid := [][]symbol{
		{0, 0, 0, 0, 18, 12, 14, 3}, {-1, -1, -1, -1, 18, 1, 4, 4}, {-2, -2, -2, -2, 18, 13, 13, 11},
		{-3, -3, -3, -3, 18, 11, 9, 14}, {-4, -4, -4, -4, 18, 9, 14, 14}, {-5, -5, -5, -5, 18, 13, 11, 14},

		// {"A", "W"},
		// {"B", "W"},
		// {"C", "W"},
		// {"W", "W", "A"},
	}
	n := 0
	f := n < len(grid)
	for f {
		f = false
		for _, g := range grid {
			if n < len(g) {
				f = true
				fmt.Printf("%4d", g[n])
			}
		}
		fmt.Println()
		n++
	}
	items := Ways(grid, 18)
	for _, w := range items {
		fmt.Printf("%2d x %2d  %d\n", w.Ways, w.Symbol, w.Width)
	}
}
