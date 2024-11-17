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
	Hazard  bool
	Sturm   bool
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
	NoRisk
	RiskOne
	NewRisk
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
		n := i + l

		swap := !d.IsDiam
		if !swap && h.IsRoyal && !d.IsRoyal {
			swap = n >= 4
			if !swap {
				switch Strategy {
				case SwapCourt:
					swap = true
				case RiskOne:
					if scr.Kenta && n >= 3 {
						swap = true
					}
				case NewRisk:
					if scr.Kenta {
						m := 4 - i
						r := scr.RHand
						swap = m-r <= 1
						if false && !swap {
							if l == 1 && i == 1 && scr.Diam[0].Kind == 11 {
								// if i == 1 && scr.Diam[0].Load > 1 {
								swap = h.Kind == 12
								// swap = h.Load > 1
								if swap {
									scr.Sturm = true
								}
							}
						}
					}
				}
				if swap {
					scr.Hazard = swap
					scr.Force++
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

// var flip bool

// Play one hand.
func (scr *Screen) Play(bet float64) HuntResponse {
	Dealer.Reset()
	// if flip = !flip; flip {
	// Dealer.AddCheats("Q♦", "Q♠", "Q♥", "Q♣", "J♦", "2♦")
	// }
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
	Hazard   int
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
	cat_win      = "2♦ + 4♦"
	cat_free     = "(3) + 3♦"
)

const (
	win_handy    = 50000
	win_straight = 2850
	win_four     = 800
	win_royal    = 1
	win_free     = 1
	win_2        = 0.5
	win_3        = 0
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
	case 2:
		resp.Win = win_2
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
				if scr.Hazard {
					resp.Hazard++
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
	resp.Total = resp.Win + resp.JackPot //+ resp.Free

	if scr.Sturm {
		AddCat("sturm", resp.Total+resp.Free)

		scr.Sturm = false
	}

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
			scr.Sturm = false
			play++
			ans := scr.Play(chip)
			opens[ans.Open]++
			// chart[ans.Open][ans.Count]++
			chart[ans.Open][ans.Close]++
			jp := ans.JackPot
			if scr.Force > 0 {
				AddCat("force", 0)
			}
			if ans.Win > 0 {
				AddCat(cat_win, ans.Win)
			}
			if ans.Total > 0 {
				win.Add(ans.Total)
				AddCat("total", ans.Total)
			}
			if ans.Free > 0 {
				AddCat(cat_free, 0)
			}
			AddCat(ans.Cat, ans.Win)
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
			if ans.Hazard > 0 {
				AddCat("hazard", 0)
			}
			run += int(ans.Free)
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
	fmt.Printf("\n%d tickets,  %d free games,  %d max free\n", play.Cnt, int(play.Sum)-play.Cnt, int(play.Max)-1)
	fmt.Print("strategy: ")
	switch Strategy {
	case SwapCourt:
		fmt.Print("swap low diamond with strongest court card")
	case NoRisk:
		fmt.Print("no risk")
	case RiskOne:
		fmt.Print("risk one diamond lost")
	case NewRisk:
		fmt.Print("optimal")
	default:
		fmt.Print("no swap diamond")
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf("%-30s  %10.2f\n", "2♦", win_2)
	fmt.Printf("%-30s  %10d\n", "3♦", win_free)
	fmt.Printf("%-30s  %10d\n", "4♦", win_4)
	fmt.Printf("%-30s  %10d\n", cat_handy, win_handy)
	fmt.Printf("%-30s  %10d\n", cat_straight, win_straight)
	fmt.Printf("%-30s  %10d\n", cat_four, win_four)
	fmt.Printf("%-30s  %10d\n", cat_royal, win_royal)
	fmt.Println()
	fmt.Println("category                         count              sum     probability         rtp             rate")
	spisak := []string{"-", "0♦", "1♦", "2♦", "3♦", "4♦",
		cat_handy, cat_straight, cat_four, cat_royal, "-",
		"total",
		"", cat_win, cat_court, cat_free, "", "", "sturm", "force", "waste", "hazard",
		"", "JDQ"}
	counter := play.Sum
	for _, d := range spisak {
		if d == "-" {
			fmt.Print("----------------------------------------------------------------------------------------------------")
		}
		if d == "total" {
			counter = play.Sum
		}
		s, e := CatStat[d]
		if e {
			prob := float64(s.Cnt) / counter
			rtp := s.Sum / bet.Sum
			fmt.Printf("%-26s  %10d  ", d, s.Cnt)
			if s.Sum > 0 {
				fmt.Printf("%15.2f", s.Sum)
			} else {
				fmt.Printf("%15s", "")
			}
			fmt.Printf("  %13.9f%%  ", 100*prob)
			if rtp > 0 {
				fmt.Printf("%9.5f%%", 100*rtp)
			} else {
				fmt.Printf("%10s", "")
			}
			fmt.Printf("  %15.2f", 1/prob)
		}
		fmt.Println()
		if d == "force" {
			counter = float64(s.Cnt)
		}
	}
	// free := CatStat[cat_free].Sum
	// twin := win.Sum - free
	// tbet := bet.Sum - free
	// rtp := twin / tbet
	fmt.Println()
	// fmt.Printf("\nrtp = (%.0f - %.0f) / (%.0f - %.0f) = %.0f / %.0f =  %.2f%%\n", win.Sum, free, bet.Sum, free, twin, tbet, 100*rtp)
	fmt.Println()
}
