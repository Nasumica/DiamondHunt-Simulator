package main

import (
	"fmt"
	"time"
)

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

var Kinds = [...]string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
var Suits = [...]string{"♠", "♦", "♥", "♣"} // preferans order

type Card struct {
	Face  string
	Kind  int
	Suit  int
	Card  int
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

// Swap with other card.
func (card *Card) Swap(other *Card) {
	*card, *other = *other, *card
}

// Deck of cards.
type Deck struct {
	Croupier LCPRNG
	Cards    []int
	Rest     int
}

// Initialize deck of cards.
func (deck *Deck) Init() {
	deck.Croupier.Randomize()
	deck.Cards = deck.Croupier.Deck()
	deck.Reset()
}

// New deal.
func (deck *Deck) Reset() {
	deck.Rest = len(deck.Cards)
}

// Draw single card from deck.
func (deck *Deck) Draw() (card Card) {
	if deck.Rest > 0 {
		n := deck.Croupier.Choice(deck.Rest)
		card.Index, card.Card = n, deck.Cards[n]
		card.Reveal()
		deck.Rest--
		deck.Cards[deck.Rest], deck.Cards[n] = deck.Cards[n], deck.Cards[deck.Rest]
	} else {
		card.Index = -1 // error
	}
	return
}

// Deal cards from deck (Fisher-Yates).
func (deck *Deck) Deal(cards int) (deal []Card) {
	if cards > 0 {
		deal = make([]Card, cards)
		for i := range deal {
			deal[i] = deck.Draw()
		}
	}
	return
}

// Deal cards from whole deck.
func (deck *Deck) NewDeal(cards int) []Card {
	deck.Reset()
	return deck.Deal(cards)
}

func Likelihood(cards []Card) ([]int, int) {
	var p BitPoker
	p.Classic()
	var h uint64
	for _, c := range cards {
		h |= 1 << (c.Card - 1)
	}
	return p.Likelihood(h)
}

func SpeedTest(n int) {
	start := time.Now()
	var deck Deck
	deck.Init()
	for i := 0; i < n; i++ {
		deck.NewDeal(4)
		deck.Deal(4)
	}
	elapsed := time.Since(start).Seconds()
	speed := float64(n) / elapsed
	fmt.Printf("elapsed = %.3f\",  speed = %.0f deals / s\n", elapsed, speed)
}

// Croupier with deck of cards.
var Dealer Deck

func init() {
	Dealer.Init()
}
