package main

import (
	"DHSimulator/rng"
	"fmt"
)

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

// const million = 1000 * 1000

const ( // preferans order
	NoSuit      = iota
	SpadeSuit   // ♠, pik,  лист
	DiamondSuit // ♦, karo, баклава
	HeartSuit   // ♥, srce, срце
	ClubSuit    // ♣, tref, детелина
)

// Count diamonds.
func Diamonds(cards []Card) int {
	n := 0
	for _, c := range cards {
		if c.Suit == DiamondSuit {
			n++
		}
	}
	return n
}

// Game screen
type Screen struct {
	Hand    []Card
	Diam    []Card
	Best    []int
	Open    int
	Swaps   int
	Flow    string
	Verbose bool
}

func (scr *Screen) History(s string) {
	scr.Flow += s
}

// Swap pick best strategy.
func (scr *Screen) Strategy() {
	s := []Card{}
	for _, c := range scr.Hand {
		if c.Suit == DiamondSuit { // insertion sort
			i := len(s)
			s = append(s, c)
			for j := i - 1; (i > 0) && (s[j].Load < c.Load); j-- {
				s[i], i = s[j], j
			}
			s[i] = c
		}
	}
	scr.Best = []int{}
	for _, c := range s {
		scr.Best = append(scr.Best, c.Index)
	}
}

// Base game deal.
func (scr *Screen) Deal() {
	scr.Hand = Dealer.NewDeal(4) // 4 cards in hand from new deck
	scr.Diam = Dealer.Null()     // no cards in diamond yet
	scr.Strategy()               // swap strategy
	scr.Swaps = 0                // reset counter
	scr.Flow = ""
	scr.Open = len(scr.Best)
}

// Draw card in diamond.
func (scr *Screen) Draw() int {
	card := Dealer.Draw()             // draw single card from rest of the deck
	card.Index = len(scr.Diam)        // hand card index
	scr.Diam = append(scr.Diam, card) // add card to diamond
	return card.Index
}

// Hunt for diamond.
func (scr *Screen) Hunt() (more bool) {
	const (
		reuse = true
		diam  = true
	)

	i := scr.Draw()   // diamond card index
	d := &scr.Diam[i] // card from diamond

	if scr.Verbose {
		scr.History("[" + Hand(&scr.Hand) + "][" + Hand(&scr.Diam))
	}

	swap := false
	if l := len(scr.Best); l > 0 { // test
		j := scr.Best[0]  // get swap index
		h := &scr.Hand[j] // card from hand

		swap = !d.IsDiam()

		if diam && !swap {
			if reuse {
				swap = h.Load > d.Load
			} else {
				swap = h.IsRoyal() && !d.IsRoyal()
			}
		}

		if swap { // swap
			h.Index, d.Index = i, j // preserve index
			(*h), (*d) = (*d), (*h) // swap cards
			scr.Best = scr.Best[1:] // remove from list
			scr.Swaps++
			if reuse && h.IsDiam() {
				scr.Strategy() // recalc
			}
			if scr.Verbose {
				scr.History(" = " + d.Face)
			}
		}
	}
	if scr.Verbose {
		scr.History("] > ")
	}

	i++
	more = d.IsDiam() && i < 4

	return
}

// Play one hand.
func (scr *Screen) Play(bet float64) HuntResponse {
	scr.Deal()
	for next := true; next; {
		next = scr.Hunt()
	}
	return scr.Eval(bet)
}

type HuntResponse struct {
	Hand     []Card  // closing hand (diamonds only)
	Value    int     // hand value
	Count    int     // number of ♦
	Royals   int     // court cards
	Straight bool    // is straight?
	Cat      string  // category
	Win      float64 // win amount
	JackPot  float64
	Name     string
	Total    float64
	Free     float64 // number of free spins
	Swaps    int
	Open     int
	FLow     string
}

// Evaluate hand.
func (scr *Screen) Eval(bet float64) (resp HuntResponse) {
	resp.Swaps = scr.Swaps
	resp.Open = scr.Open
	for _, c := range scr.Diam {
		if c.Suit == DiamondSuit {
			resp.Count++
			if c.Kind >= 11 {
				resp.Royals++
			}
			resp.Hand = append(resp.Hand, c)
			resp.Value = (resp.Value << 4) + c.Kind // hex
		}
	}
	if scr.Verbose {
		scr.History("[" + Hand(&scr.Hand) + "]" + "[" + Hand(&resp.Hand) + "]")
	}

	const straight int = 0xbcde // JQKA
	resp.Straight = resp.Value == straight

	cat := fmt.Sprintf("%d", resp.Count)
	resp.Cat = cat + "♦"

	switch resp.Count {
	case 3:
		resp.Free = 1
	case 4:
		resp.Win = 4

		switch resp.Royals {
		case 0:
		case 4:
			if resp.Straight {
				resp.JackPot = 6000
				resp.Name = "straight"
			} else {
				resp.JackPot = 800
				resp.Name = "four"
			}
		default:
			resp.Free = 1
			resp.Name = "royal"
		}
	}

	if scr.Verbose {
		if resp.Name != "" {
			scr.History(" " + resp.Name)
		} else {
			scr.History(" " + cat)
		}
		resp.FLow = scr.Flow
	}

	resp.Win *= bet
	resp.JackPot *= bet
	resp.Free *= bet
	resp.Total = resp.Win + resp.JackPot + resp.Free

	return
}

var (
	CatStat = map[string]rng.StatCalc{}
	CntStat = [5]rng.StatCalc{}
)

func AddCat(cat string, x float64) {
	c := CatStat[cat]
	c.Cat = cat
	c.Add(x)
	CatStat[cat] = c
}

func DiamondHunt(iter int, chips ...float64) {
	var scr Screen
	scr.Verbose = true
	var bet, win rng.StatCalc
	bet.Cat, win.Cat = "bet", "win"

	opens := [5]int{}
	chart := [5][5]int{}

	for cnt := 1; cnt <= iter; cnt++ {
		chip := rng.WSOGMM.Value(&chips, 1)
		bet.Add(chip)

		play := 0
		for run := 1; run > 0; run-- {
			play++
			ans := scr.Play(chip)
			opens[ans.Open]++
			chart[ans.Open][ans.Count]++
			jp := ans.JackPot + ans.Free
			if ans.Total > 0 {
				win.Add(ans.Total)
			}
			if ans.Free > 0 {
				AddCat("free", float64(ans.Free))
			}
			if ans.Count == 3 {
				AddCat(ans.Cat, ans.Free)
			} else {
				AddCat(ans.Cat, ans.Win)
			}
			if ans.Name != "" {
				AddCat(ans.Name, jp)
			}
			CntStat[ans.Count].Add(ans.Total)
			if ans.Royals == 4 {
				AddCat("court", ans.JackPot)
			}
		}

		AddCat("play", float64(play))
	}

	play := CatStat["play"]

	for h, o := range opens {
		fmt.Printf("%-4d  %-4s  %10d", h, "", o)
		prob := float64(o) / play.Sum
		fmt.Printf("  %9.5f%%", 100*prob)
		if prob > 0 {
			rate := 1 / prob
			fmt.Printf("  %27.2f", rate)
		}
		fmt.Println()
		for d, c := range chart[h] {
			fmt.Printf("%-4s  %-4d  %10d", "", d, c)
			prob := float64(c) / play.Sum
			fmt.Printf("  %9.5f%%", 100*prob)
			if prob > 0 {
				rate := 1 / prob
				fmt.Printf("  %27.2f", rate)
			}
			fmt.Println()
		}
	}

	fmt.Println()
	spisak := []string{"0♦", "1♦", "2♦", "3♦", "4♦", "straight", "four", "royal", "court", "free"}
	for _, d := range spisak {
		// for d, s := range CatStat {
		s := CatStat[d]
		prob := float64(s.Cnt) / play.Sum
		rtp := s.Sum / bet.Sum
		fmt.Printf("%-10s  %10d  %9.5f%%  %9.5f%%  %15.2f\n", d, s.Cnt, 100*prob, 100*rtp, 1/prob)
	}
	rtp := win.Sum / bet.Sum
	fmt.Printf("rtp = %.2f%%\n", 100*rtp)
	fmt.Println()
}
