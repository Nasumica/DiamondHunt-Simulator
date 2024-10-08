package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

import (
	"DHSimulator/rng"
	"fmt"
)

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
	Hand    Cards
	Diam    Cards
	Best    []int
	Open    int
	Swaps   int
	Flow    string
	Verbose bool
	Wait    int
	Rest    int
	Deck    int
	RHand   int
	RDiam   int
	Kenta   bool
	Force   int
	Count   int
}

func (scr *Screen) History(s string) {
	if scr.Verbose {
		scr.Flow += s
	}
}

// Swap pick best strategy.
func (scr *Screen) Strategy() {
	s := Cards{}
	for _, c := range scr.Hand {
		if c.IsDiam { // insertion sort
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
	scr.Hand = Dealer.Deal(4) // 4 cards in hand from new deck
	scr.Diam = Dealer.Null()  // no cards in diamond yet
	scr.Strategy()            // swap strategy
	scr.Swaps = 0             // reset counter
	scr.Flow = ""
	scr.Wait = 5
	scr.Deck = 52 - len(scr.Hand)
	scr.Rest = 13 - len(scr.Best)
	scr.Open = len(scr.Best)
	scr.RHand = 0
	scr.RDiam = 0
	scr.Kenta = true
	scr.Force = 0
	scr.Count = 0
	for _, c := range scr.Hand {
		if c.IsRoyal {
			scr.RHand++
		}
	}
}

// Draw card in diamond.
func (scr *Screen) Draw() int {
	card := Dealer.Draw()             // draw single card from rest of the deck
	card.Index = len(scr.Diam)        // hand card index
	scr.Diam = append(scr.Diam, card) // add card to diamond
	scr.Deck--
	if card.IsDiam {
		scr.Rest--
	}
	return card.Index
}

const (
	NoStrategy = iota
	SwapCourt
	RiskOne
)

var Strategy = SwapCourt

// Hunt for diamond.
func (scr *Screen) Hunt() (more bool) {
	const (
		reuse = false
	)

	i := scr.Draw()   // diamond card index
	d := &scr.Diam[i] // card from diamond
	if d.IsRoyal {
		scr.RDiam++
	}
	if d.IsDiam {
		scr.Count++
	}

	if scr.Verbose {
		scr.History("[" + scr.Hand.Faces() + "][" + scr.Diam.Faces())
	}

	if l := len(scr.Best); l > 0 { // test
		j := scr.Best[0]  // get swap index
		h := &scr.Hand[j] // card from hand

		swap := !d.IsDiam

		if !swap {
			switch Strategy {
			case SwapCourt:
				swap = h.IsRoyal && !d.IsRoyal
				scr.Force++
			case RiskOne:
				if h.IsRoyal && !d.IsRoyal {
					n := i + l
					swap = n >= 4
					if !swap && scr.Kenta && n >= 3 {
						scr.Force++
						swap = true
					}
				}
			}

		}

		if swap { // swap
			if h.IsRoyal {
				scr.RHand--
				scr.RDiam++
			}
			if d.IsRoyal {
				scr.RDiam--
			}
			h.Index, d.Index = i, j // preserve index
			(*h), (*d) = (*d), (*h) // swap cards
			scr.Best = scr.Best[1:] // remove from list
			scr.Swaps++
			if reuse && h.IsDiam {
				scr.Strategy() // recalc
			}
			if scr.Verbose {
				scr.History(" = " + d.Face)
			}
		}

		if d.Load == scr.Wait {
			scr.Wait--
		} else {
			scr.Wait = 0
		}
	}

	scr.Kenta = scr.Kenta && d.IsRoyal

	if scr.Verbose {
		scr.History("] > ")
	}

	i++
	more = d.IsDiam && i < 4

	return
}

// Play one hand.
func (scr *Screen) Play(bet float64) HuntResponse {
	Dealer.Reset()
	// Dealer.Hide("J♦", "Q♦", "T♦", "9♦", "8♦", "7♦", "6♦", "5♦", "4♦", "3♦", "2♦", "A♦")
	// Dealer.AddCheats("K♦")
	// Dealer.Release()
	// Dealer.AddCheats("Q♦", "J♦", "2♦")
	// Dealer.AddCheats("Q♦", "J♦", "2♦", "3♦", "K♦", "7♦", "5♦", "6♦")
	scr.Deal()
	for next := true; next; {
		next = scr.Hunt()
	}
	return scr.Eval(bet)
}

type HuntResponse struct {
	Final    Cards // closing hand (diamonds only)
	Value    int   // hand value
	Code     int   // hand code
	Count    int   // number of ♦
	Diams    int
	Waste    int
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
	Close    int
}

const (
	cat_handy    = "(0) ROYAL STRAIGHT NO SWAP"
	cat_straight = "(1) ROYAL STRAIGHT"
	cat_four     = "(2) ROYAL FOUR"
	cat_royal    = "(3) ROYAL CARD"
	cat_court    = "(0) + (1) + (2)"
	cat_free     = "(3) + 3♦"
)

const (
	win_handy    = 50000
	win_straight = 6000
	win_four     = 850
	win_royal    = 1
	win_free     = 1
	win_4        = 4
)

// Evaluate hand.
func (scr *Screen) Eval(bet float64) (resp HuntResponse) {
	resp.Swaps = scr.Swaps
	resp.Open = scr.Open
	resp.Code = 1
	resp.Close = scr.Count
	// scr.Diam = Make("J♦", "Q♦", "K♦", "A♦")
	resp.Diams = 0
	for i := len(Dealer.Cards); i > Dealer.Rest; {
		i--
		j := Dealer.Cards[i]
		c := CardVirtues[j]
		if c.IsDiam {
			resp.Diams++
		}
	}
	for _, c := range scr.Diam {
		if c.IsDiam {
			resp.Count++
			if c.IsRoyal {
				resp.Royals++
			}
			resp.Final = append(resp.Final, c)
			resp.Value = resp.Value*10 + c.Load
			resp.Code *= c.Code
		}
	}
	if scr.Verbose {
		scr.History("[" + scr.Hand.Faces() + "]" + "[" + resp.Final.Faces() + "]")
	}

	const straight_code int = 5432 // JQKA
	resp.Straight = resp.Value == straight_code

	cat := fmt.Sprintf("%d", resp.Count)
	resp.Cat = cat + "♦"

	switch resp.Count {
	case 3:
		resp.Free = win_free
	case 4:
		resp.Win = win_4
		switch resp.Royals {
		case 0:
		case 4:
			if resp.Straight {
				if scr.Swaps == 0 {
					resp.JackPot = win_handy
					resp.Name = cat_handy
				} else {
					resp.JackPot = win_straight
					resp.Name = cat_straight
				}
			} else {
				resp.JackPot = win_four
				resp.Name = cat_four
			}
		default:
			resp.Free = win_royal
			resp.Name = cat_royal
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

	if resp.Diams >= 4 && resp.Count < 4 {
		resp.Waste++
		// fmt.Println(resp.FLow)
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
	scr.Verbose = false
	var bet, win rng.StatCalc
	bet.Cat, win.Cat = "bet", "win"

	opens := [5]int{}
	chart := [5][5]int{}

	for cnt := 1; cnt <= iter; cnt++ {
		chip := rng.WSOGMM.Value(chips, 1)
		bet.Add(chip)

		play := 0
		for run := 1; run > 0; run-- {
			play++
			ans := scr.Play(chip)
			opens[ans.Open]++
			// chart[ans.Open][ans.Count]++
			chart[ans.Open][ans.Close]++
			jp := ans.JackPot + ans.Free
			if scr.Force > 0 {
				AddCat("force", float64(scr.Force))
			}
			if ans.Total > 0 {
				win.Add(ans.Total)
				AddCat("total", ans.Total)
			}
			if ans.Free > 0 {
				AddCat(cat_free, float64(ans.Free))
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
				AddCat(cat_court, ans.JackPot)
			}
			if ans.Waste > 0 {
				AddCat("waste", 0)
			}
		}

		AddCat("play", float64(play))
	}

	play := CatStat["play"]

	/*
		for h, o := range opens {
			fmt.Printf("%16s%-4d  %-4s  %10d", "", h, "", o)
			prob := float64(o) / play.Sum
			fmt.Printf("  %13.9f%%", 100*prob)
			if prob > 0 {
				rate := 1 / prob
				fmt.Printf("  %27.2f", rate)
			}
			fmt.Println()
			for d, c := range chart[h] {
				fmt.Printf("%16s%-4s  %-4d  %10d", "", "", d, c)
				prob := float64(c) / play.Sum
				fmt.Printf("  %13.9f%%", 100*prob)
				if prob > 0 {
					rate := 1 / prob
					fmt.Printf("  %27.2f", rate)
					// fmt.Printf("  %12.9f%%", 100*prob)
				}
				fmt.Println()
			}
		}
	*/
	fmt.Println()
	fmt.Printf("%-30s  %10d\n", "3♦", win_free)
	fmt.Printf("%-30s  %10d\n", "4♦", win_4)
	fmt.Printf("%-30s  %10d\n", cat_handy, win_handy)
	fmt.Printf("%-30s  %10d\n", cat_straight, win_straight)
	fmt.Printf("%-30s  %10d\n", cat_four, win_four)
	fmt.Printf("%-30s  %10d\n", cat_royal, win_royal)
	fmt.Println()
	spisak := []string{"0♦", "1♦", "2♦", "3♦", "4♦",
		cat_handy, cat_straight, cat_four, cat_royal,
		"total", "", cat_court, cat_free, "", "waste", "force"}
	for _, d := range spisak {
		s, e := CatStat[d]
		if e {
			prob := float64(s.Cnt) / play.Sum
			rtp := s.Sum / bet.Sum
			fmt.Printf("%-26s  %10d  %13.9f%%  %9.5f%%  %15.2f", d, s.Cnt, 100*prob, 100*rtp, 1/prob)
		}
		fmt.Println()
	}
	free := CatStat[cat_free].Sum
	rtp := (win.Sum - free) / (bet.Sum - free)
	fmt.Println()
	if Strategy == SwapCourt {
		fmt.Println("strategy: swap low diamond with strongest court card")
	} else if Strategy == RiskOne {
		fmt.Println("strategy: risk one diamond lost")
	} else {
		fmt.Println("no swap diamond")
	}
	fmt.Printf("\nrtp = (%.0f - %.0f) / (%.0f - %.0f) =  %.2f%%\n", win.Sum, free, bet.Sum, free, 100*rtp)
	fmt.Println()
}
