package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

type List = []string

var Kinds = List{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
var Suits = List{"♠", "♦", "♥", "♣"} // preferans order

type Card struct {
	Card  int
	Kind  int
	Suit  int
	Face  string
	Index int
}

// Reveal card.
func (card *Card) Reveal() {
	if c := card.Card; 1 <= c && c <= 52 {
		c--
		k, s := c/4, c%4
		card.Face = Kinds[k] + Suits[s]
		card.Kind = k + 2
		card.Suit = s + 1
	} else {
		card.Face, card.Kind, card.Suit = "", 0, 0
	}
}

type Pack = []int

// My favourite deck of cards.
type Piatnik struct {
	Croupier LCPRNG
	Cards    Pack
	Rest     int
}

// Initialize deck of cards.
func (deck *Piatnik) Init() {
	deck.Croupier.Randomize()
	deck.Cards = deck.Croupier.Deck()
	deck.Reset()
}

// New deal.
func (deck *Piatnik) Reset() {
	deck.Rest = len(deck.Cards)
}

// Draw single card from deck.
func (deck *Piatnik) Draw() (card Card) {
	if deck.Rest > 0 {
		n := deck.Croupier.Choice(deck.Rest)
		card.Index = n
		card.Card = deck.Cards[n]
		card.Reveal()
		deck.Rest--
		deck.Cards[deck.Rest], deck.Cards[n] = deck.Cards[n], deck.Cards[deck.Rest]
	} else {
		card.Index = -1 // error
	}
	return
}

type Cards = []Card

// Deal cards from deck (Fisher-Yates).
func (deck *Piatnik) Deal(cards int) (deal Cards) {
	if cards > 0 {
		deal = make(Cards, cards)
		for i := range deal {
			deal[i] = deck.Draw()
		}
	}
	return
}

// Deck of cards in use.
var Deck Piatnik

// Ready to use.
func init() {
	Deck.Init()
}
