package main

import "fmt"

type symbol = string

type WaysItem struct {
	Symbol  symbol  `json:"s"`           // symbol
	Order   int     `json:"o,omitempty"` // order: 1 =  L2R; -1 = R2L, 0 = both
	Line    float64 `json:"l,omitempty"` // single line win amount
	Factors []int   `json:"f,omitempty"` // consecutive occurrences
	Width   int     `json:"m,omitempty"` // line width
	Wilds   int     `json:"e,omitempty"` // excluded wilds only lines
	Ways    int     `json:"w,omitempty"` // number of ways = Π factors - wilds
	Payout  float64 `json:"p,omitempty"` // payout = line * ways
}

func Ways(grid [][]symbol, wild symbol) (result []WaysItem) {
	width := len(grid)

	symbols := map[symbol]int{} // list of symbols
	for _, reel := range grid {
		for _, symbol := range reel {
			if symbol != wild {
				symbols[symbol]++
			}
		}
	}

	var wilds []WaysItem // wilds ways

	scan := func(symbol symbol) (win WaysItem) {
		win = WaysItem{Symbol: symbol, Ways: 1}
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
		{"W", "W"},
		{"A", "W"},
		{"A", "W"},
		{"A"},
		// {"W", "W", "A"},
	}
	n := 0
	f := n < len(grid)
	for f {
		f = false
		for _, g := range grid {
			if n < len(g) {
				f = true
				fmt.Printf("%-3s", g[n])
			}
		}
		fmt.Println()
		n++
	}
	items := Ways(grid, "W")
	for _, w := range items {
		fmt.Printf("%d x %s %d\n", w.Ways, w.Symbol, w.Width)
	}
}

func init() {
	WaysTest()
}
