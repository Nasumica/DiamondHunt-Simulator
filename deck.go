package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

const ( // preferans order
	NoSuit      = iota
	SpadeSuit   // ♠ pik
	DiamondSuit // ♦ karo
	HeartSuit   // ♥ srce
	ClubSuit    // ♣ tref
)

type (
	Cards = []int
	List  = []string
)

// My favourite deck of cards.
type Piatnik struct {
	Cards    Cards
	Kinds    List
	Suits    List
	Croupier LCPRNG
	Rest     int
}

// Initialize deck of cards.
func (deck *Piatnik) Init() {
	deck.Kinds = List{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	deck.Suits = List{"♠", "♦", "♥", "♣"} // preferans order
	deck.Croupier.Randomize()
	deck.Cards = make(Cards, 52)
	for c := range deck.Cards {
		deck.Cards[c] = c + 1
	}
	deck.Reset()
}

// Ready to new deal.
func (deck *Piatnik) Reset() {
	deck.Rest = len(deck.Cards)
}

type Card struct {
	Card  int
	Kind  int
	Suit  int
	Face  string
	Index int
}

// Draw single card from deck.
func (deck *Piatnik) Draw() (card Card) {
	if deck.Rest > 0 {
		c := deck.Croupier.Choice(deck.Rest)
		card.Index = c
		card.Card = deck.Cards[c]
		deck.Rest--
		deck.Cards[deck.Rest], deck.Cards[c] = deck.Cards[c], deck.Cards[deck.Rest]
		c = card.Card - 1
		k, s := c/4, c%4
		card.Face = deck.Kinds[k] + deck.Suits[s]
		card.Kind = k + 2
		card.Suit = s + 1
	} else {
		card.Card = -1 // error
	}
	return
}

type Hold = []Card

// Deal cards from deck (Fisher-Yates).
func (deck *Piatnik) Deal(cards int) (hold Hold) {
	if cards > 0 {
		hold = make(Hold, cards)
		for i := range hold {
			hold[i] = deck.Draw()
		}
	}
	return
}

// Deck of cards in use.
var Deck Piatnik

func init() {
	Deck.Init()
}
