package main

import "fmt"

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
	Hand   []Card
	Diam   []Card
	Open   int
	Swaped int
}

// Sort hand.
func (scr *Screen) Sort() {
	SortCards(&scr.Hand)
}

// Base game deal.
func (scr *Screen) Deal() {
	scr.Hand = Dealer.NewDeal(4)  // 4 cards in hand from new deck
	scr.Diam = Dealer.Deal(0)     // no cards in diamond yet
	scr.Swaped = 0                // reset counter
	scr.Open = Diamonds(scr.Hand) // for stat
}

// Reveal new card in diamond.
func (scr *Screen) Next() (card Card) {
	card = Dealer.Draw()              // draw single card from rest of the deck
	scr.Diam = append(scr.Diam, card) // add card to diamond
	return
}

// Swap cards (if possible).
func (scr *Screen) Swap() (f bool) {
	l := len(scr.Diam)
	h, d := &scr.Hand[0], &scr.Diam[l-1]
	f = d.Suit == DiamondSuit
	if !f {
		if f = h.Suit == DiamondSuit && h.Load > d.Load; f { // swap
			*h, *d = *d, *h
			scr.Swaped++
		}
	}
	f = f && l < 4
	return
}

// Hunt for diamond.
func (scr *Screen) Hunt() bool {
	scr.Next()
	scr.Sort() // best strategy sort
	return scr.Swap()
}

// Play one hand.
func (scr *Screen) Play() Response {
	scr.Deal()
	for f := true; f; {
		f = scr.Hunt()
	}
	return scr.Eval()
}

type Response struct {
	Hand     []Card  // closing hand (diamonds only)
	Value    int     // hand value
	Count    int     // number of ♦
	Royals   int     // court cards
	Straight bool    // is straight?
	Cat      string  // category
	Win      float64 // win amount
	Free     int     // number of free spins
	Swaped   int
}

// Evaluate hand.
func (scr *Screen) Eval() (resp Response) {
	resp.Swaped = scr.Swaped
	for _, c := range scr.Diam {
		c.Reveal()
		if c.Suit == DiamondSuit {
			resp.Count++
			resp.Value *= 16
			if c.Kind >= 11 {
				resp.Royals++
			}
			resp.Hand = append(resp.Hand, c)
			resp.Value += c.Kind
		}
	}
	const straight int = 0xbcde
	resp.Straight = resp.Value == straight

	resp.Cat = fmt.Sprintf("%d♦", resp.Count)

	switch resp.Count {
	case 3:
		resp.Free = 1
	case 4:
		switch resp.Royals {
		case 0:
			resp.Win = 4
		case 4:
			if resp.Straight {
				resp.Win = 6000
				resp.Cat = "straight"
			} else {
				resp.Win = 800
				resp.Cat = "royals"
			}
		default:
			resp.Win = 4
			resp.Free = 1
			resp.Cat = "court"
		}
	}

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

func DiamondHunt(iter int) {
	var scr Screen
	var bet, win StatCalc
	bet.Cat, win.Cat = "bet", "win"

	for cnt := 1; cnt <= iter; cnt++ {
		bet.Add(1)
		fg := 0
		play := 0
		for run := 1; run > 0; run-- {
			play++
			ans := scr.Play()
			// if ans.Free > 0 {
			// 	AddCat("free", float64(ans.Free))
			// 	ans.Win += float64(ans.Free)
			// }
			if ans.Win > 0 {
				win.Add(ans.Win)
			}
			if ans.Free > 0 {
				run += ans.Free
				fg += ans.Free
			}
			AddCat(ans.Cat, ans.Win)
			CntStat[ans.Count].Add(ans.Win)
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
