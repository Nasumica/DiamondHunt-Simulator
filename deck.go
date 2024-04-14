package main

import "fmt"

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

const DiamondSuit int = 3 // baklava

type (
	Cards = []int
	List  = []string
)

type DeckOfCards struct {
	Cards    Cards
	Kinds    List
	Suits    List
	Croupier LCPRNG
	Rest     int
}

// Initialize deck of cards.
func (deck *DeckOfCards) Init() {
	deck.Kinds = List{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	deck.Suits = List{"♠", "♦", "♥", "♣"} // preferans order
	deck.Croupier.Randomize()
	deck.Cards = make(Cards, 52)
	for c := range deck.Cards {
		deck.Cards[c] = c + 1
	}
	deck.NewDeal()
}

// Ready to new deal.
func (deck *DeckOfCards) NewDeal() {
	deck.Rest = len(deck.Cards)
}

type Card struct {
	Value int
	Kind  int
	Suit  int
	Face  string
	Index int
}

// Draw single card from deck.
func (deck *DeckOfCards) Draw() (card Card) {
	if deck.Rest > 0 {
		n := deck.Croupier.Choice(deck.Rest)
		card.Index = n
		card.Value = deck.Cards[n]
		deck.Rest--
		deck.Cards[deck.Rest], deck.Cards[n] = deck.Cards[n], deck.Cards[deck.Rest]
		n = card.Value - 1
		k, s := n/4, n%4
		card.Face = deck.Kinds[k] + deck.Suits[s]
		card.Kind = k + 2
		card.Suit = s + 1
	}
	return
}

// Deck of cards in use.
var Deck DeckOfCards

func init() {
	Deck.Init()
	fmt.Println(Deck.Draw())
	fmt.Println(Deck.Draw())
	fmt.Println(Deck.Draw())
	fmt.Println(Deck.Draw())
}
