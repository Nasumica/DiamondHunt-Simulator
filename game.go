package main

import (
	"fmt"
)

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

const million = 1000 * 1000

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
	Hand  []Card
	Diam  []Card
	Swap  []int
	Open  int
	Swaps int
}

// Swap pick best strategy.
func (scr *Screen) Best() {
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
	scr.Swap = []int{}
	for _, c := range s {
		scr.Swap = append(scr.Swap, c.Index)
	}
}

// Base game deal.
func (scr *Screen) Deal() {
	scr.Hand = Dealer.NewDeal(4) // 4 cards in hand from new deck
	scr.Diam = Dealer.Deal(0)    // no cards in diamond yet
	scr.Best()                   // swap strategy
	scr.Swaps = 0                // reset counter
	scr.Open = len(scr.Swap)
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
	i := scr.Draw()
	d := &scr.Diam[i]

	more = d.Suit == DiamondSuit
	if !more {
		more = len(scr.Swap) > 0
		if more { // swap
			var j int
			j, scr.Swap = scr.Swap[0], scr.Swap[1:]
			h := &scr.Hand[j]
			h.Index, d.Index = i, j
			(*h), (*d) = (*d), (*h)
			scr.Swaps++
		}
	}

	i++
	more = more && i < 4

	return
}

// Play one hand.
func (scr *Screen) Play(bet float64) HuntResponse {
	scr.Deal()
	scr.Best() // best strategy sort
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
	Free     int // number of free spins
	Swaps    int
	Open     int
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
	const straight int = 0xbcde // JQKA
	resp.Straight = resp.Value == straight

	resp.Cat = fmt.Sprintf("%d♦", resp.Count)

	switch resp.Count {
	case 3:
		resp.Free = 1
	case 4:
		resp.Win = 4 * bet
		switch resp.Royals {
		case 0:
		case 4:
			if resp.Straight {
				resp.JackPot = 6000 * bet
				resp.Name = "straight"
			} else {
				resp.JackPot = 800 * bet
				resp.Name = "royals"
			}
		default:
			resp.Free = 1
			resp.Name = "court"
		}
	}

	resp.Total = resp.Win + resp.JackPot

	return
}

var (
	CatStat = map[string]StatCalc{}
	CntStat = [5]StatCalc{}
)

func AddCat(cat string, x float64) {
	c := CatStat[cat]
	c.Cat = cat
	c.Add(x)
	CatStat[cat] = c
}

func DiamondHunt(iter int, chips ...float64) {
	var scr Screen
	var bet, win StatCalc
	bet.Cat, win.Cat = "bet", "win"

	for cnt := 1; cnt <= iter; cnt++ {
		chip := WSOGMM.Value(1, &chips)
		bet.Add(chip)

		fg, play := 0, 0
		for run := 1; run > 0; run-- {
			play++
			ans := scr.Play(chip)
			if ans.Total > 0 {
				win.Add(ans.Total)
			}
			if ans.Free > 0 {
				run += ans.Free
				fg += ans.Free
			}
			AddCat(ans.Cat, ans.Win)
			if ans.JackPot > 0 {
				AddCat(ans.Name, ans.JackPot)
			}
			CntStat[ans.Count].Add(ans.Total)
		}
		AddCat("play", float64(play))
		if fg > 0 {
			AddCat("free", float64(fg))
		}
	}

	rtp := win.Sum / bet.Sum
	fmt.Printf("rtp = %.2f%%\n", 100*rtp)
	play := CatStat["play"]
	for d, s := range CntStat {
		prob := float64(s.Cnt) / play.Sum
		fmt.Printf("%d  %9.5f%%\n", d, 100*prob)
	}
}

func init() {
	DiamondHunt(10 * million)
}
