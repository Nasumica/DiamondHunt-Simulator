package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

const ( // preferans order
	NoSuit      = iota
	SpadeSuit   // ♠, pik,  лист
	DiamondSuit // ♦, karo, баклава
	HeartSuit   // ♥, srce, срце
	ClubSuit    // ♣, tref, детелина
)

type Hand struct {
	Hold []Card
	Diam []Card
}

func (h *Hand) Deal() {
	h.Hold = Dealer.NewDeal(4)
	h.Diam = Dealer.Deal(4)
}
