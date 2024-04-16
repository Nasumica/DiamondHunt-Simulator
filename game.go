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
	Hand   []Card
	Diam   []Card
	Swaped int
}

func (s *Screen) Sort() {
	SortCards(&s.Hand)
}

// Base game deal.
func (s *Screen) Deal() {
	s.Hand = Dealer.NewDeal(4) // 4 cards in hand from new deck
	s.Diam = Dealer.Deal(0)    // no cards in diamond yet
	s.Sort()                   // best strategy sort
	s.Swaped = 0               // reset counter
}

// Reveal new card in diamond.
func (s *Screen) Next() (card Card) {
	card = Dealer.Draw()          // draw single card from rest of the deck
	s.Diam = append(s.Diam, card) // add card to diamond
	return
}

func (s *Screen) Swap(n int) {
	h, d := &s.Hand[n], &s.Diam[len(s.Diam)-1]
	*h, *d = *d, *h
}

func (s *Screen) Hunt() {
	d := s.Next()
	h := s.Hand[0]
	if h.Suit == DiamondSuit && h.Load > d.Load {
		s.Swap(0)
		s.Sort() // can be improved
		s.Swaped++
	}
}

func (s *Screen) Play() ([]Card, int) {
	s.Deal()
	s.Hunt()
	s.Hunt()
	s.Hunt()
	s.Hunt()
	return s.Diam, s.Swaped
}
