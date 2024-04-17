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

// Count royals.
func Royals(cards []Card) int {
	n := 0
	for _, c := range cards {
		if c.Suit == DiamondSuit && c.Kind >= 11 {
			n++
		}
	}
	return n
}

// Game screen
type Screen struct {
	Hand  []Card
	Diam  []Card
	Open  int
	Swaps int
}

// Sort hand.
func (scr *Screen) Sort() {
	SortCards(&scr.Hand)
}

// Base game deal.
func (scr *Screen) Deal() {
	scr.Hand = Dealer.NewDeal(4)  // 4 cards in hand from new deck
	scr.Diam = Dealer.Deal(0)     // no cards in diamond yet
	scr.Swaps = 0                 // reset counter
	scr.Open = Diamonds(scr.Hand) // for stat
}

// Reveal new card in diamond.
func (scr *Screen) Next() (card Card) {
	card = Dealer.Draw() // draw single card from rest of the deck
	card.Index = len(scr.Diam)
	scr.Diam = append(scr.Diam, card) // add card to diamond
	return
}

// Hunt for diamond.
func (scr *Screen) Hunt() (more bool) {
	l := scr.Next().Index
	h, d := &scr.Hand[0], &scr.Diam[l]

	more = d.Suit == DiamondSuit
	if !more {
		if more = h.Suit == DiamondSuit && h.Load > d.Load; more { // swap
			*h, *d = *d, *h
			(*h).Index, (*d).Index = (*d).Index, (*h).Index
			scr.Swaps++
			scr.Sort() // best strategy sort
		}
	}

	l++
	more = more && l < 4

	return
}

// Play one hand.
func (scr *Screen) Play() HuntResponse {
	scr.Deal()
	scr.Sort() // best strategy sort
	for next := true; next; {
		next = scr.Hunt()
	}
	return scr.Eval()
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
}

// Evaluate hand.
func (scr *Screen) Eval() (resp HuntResponse) {
	resp.Swaps = scr.Swaps
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
	const straight int = 0xbcde
	resp.Straight = resp.Value == straight

	resp.Cat = fmt.Sprintf("%d♦", resp.Count)

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
			ans := scr.Play()
			if ans.Total > 0 {
				ans.Total *= chip
				win.Add(ans.Total)
			}
			if ans.Free > 0 {
				run += ans.Free
				fg += ans.Free
			}
			AddCat(ans.Cat, ans.Total)
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
