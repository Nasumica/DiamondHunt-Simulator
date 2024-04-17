package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

type Card struct {
	Face  string
	Kind  int
	Suit  int
	Card  int
	Index int
	Load  int
}

// Reveal card.
func (card *Card) Reveal() {
	var Kinds = [...]string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	var Suits = [...]string{"♠", "♦", "♥", "♣"} // preferans order
	if c := card.Card; 1 <= c && c <= 52 {
		c--
		k, s := c/4, c%4
		card.Face = Kinds[k] + Suits[s]
		card.Kind = k + 2
		card.Suit = s + 1
		card.Load = card.Card
		if card.Suit == 2 { // karo
			l := card.Kind
			if l >= 11 {
				l = 25 - l
			}
			card.Load = l + 52
		}
	} else {
		card.Face, card.Kind, card.Suit = "", 0, 0
	}
}

func SortCards(c *[]Card) {
	n, l, r := len(*c), 0, 0
	for r = 1; r < n; r++ { // insertion sort
		p := (*c)[r]
		for l = r; l > 0 && (*c)[l-1].Load < p.Load; l-- {
			(*c)[l] = (*c)[l-1]
		}
		(*c)[l] = p
	}
}

var CardVirtues []Card

func InitVirtues() {
	for i := 0; i <= 52; i++ {
		c := Card{Card: i}
		c.Reveal()
		CardVirtues = append(CardVirtues, c)
	}
}

// Deck of cards.
type Deck struct {
	Cards    []int
	Rest     int
	Croupier LCPRNG
}

// Initialize deck of cards.
func (deck *Deck) Init() {
	deck.Croupier.Randomize()
	deck.Cards = deck.Croupier.Deck()
	deck.Reset()
}

// Reset.
func (deck *Deck) Reset() {
	deck.Rest = len(deck.Cards)
}

// Draw single card from deck.
func (deck *Deck) Draw() (card Card) {
	if deck.Rest > 0 {
		n := deck.Croupier.Choice(deck.Rest)
		c := deck.Cards[n]
		card = CardVirtues[c]
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
			c := deck.Draw()
			if c.Index >= 0 {
				c.Index = i
			}
			deal[i] = c
		}
	}
	return
}

// Deal cards from whole deck.
func (deck *Deck) NewDeal(cards int) []Card {
	deck.Reset()
	return deck.Deal(cards)
}

// Croupier with deck of cards.
var Dealer Deck

func init() {
	InitVirtues()
	Dealer.Init()
}
