package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

const ( // preferans order
	NoSuit      = iota
	SpadeSuit   // ♠, pik,  лист
	DiamondSuit // ♦, karo, баклава
	HeartSuit   // ♥, srce, срце
	ClubSuit    // ♣, tref, детелина
)

type Screen struct {
	Hand []Card
	Diam []Card
}

// Base game deal.
func (s *Screen) Deal() {
	s.Hand = Dealer.NewDeal(4) // 4 cards in hand from new deck
	s.Diam = Dealer.Deal(0)    // no cards in diamond yet
}

// Reveal new card in diamond.
func (s *Screen) Next() (card Card) {
	card = Dealer.Draw()          // draw single card from rest of the deck
	s.Diam = append(s.Diam, card) // add card to diamond
	return
}

func (s *Screen) Swap(n int) {
	if l := len(s.Diam); l > 0 {
		h := &s.Hand[n]
		d := &s.Diam[l-1]
		*h, *d = *d, *h
	}
}
