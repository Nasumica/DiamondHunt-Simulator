package main

import "DHSimulator/rng"

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

type Card struct {
	Face  string
	Kind  int
	Suit  int
	Card  int
	Index int
	Load  int
	Code  int
}

type Cards []Card

// Reveal card.
func (card *Card) Reveal() {
	kinds := [...]string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	suits := [...]string{"♠", "♦", "♥", "♣"} // preferans order
	if c := card.Card; 1 <= c && c <= 52 {
		c--
		k, s := c/4, c%4
		card.Face = kinds[k] + suits[s]
		card.Kind = k + 2
		card.Suit = s + 1
		card.Load = 0
		if card.Suit == 2 { // karo
			if card.Kind < 11 {
				card.Load = 1
			} else {
				card.Load = 16 - card.Kind
			}
		}
		card.Code = []int{1, 2, 3, 5, 7, 11}[card.Load]
	} else {
		card.Face = "★"
	}
}

func (card *Card) IsDiam() bool {
	return (card.Load > 0)
}

func (card *Card) IsRoyal() bool {
	return (card.Load > 1)
}

func (c *Cards) Sort() {
	n := len(*c)
	var l, r int
	for r = 1; r < n; r++ { // insertion sort
		p := (*c)[r]
		for l = r; l > 0 && (*c)[l-1].Load < p.Load; l-- {
			(*c)[l] = (*c)[l-1]
		}
		(*c)[l] = p
	}
}

func (c *Cards) Faces() string {
	f, d := "", ""
	for _, c := range *c {
		f += d + c.Face
		d = " "
	}
	return f
}

func (c *Cards) Code() int {
	n := 1
	for _, c := range *c {
		n *= c.Code
	}
	return n
}

func (c *Cards) Value() int {
	n := 0
	for _, c := range *c {
		n = 10*n + c.Load
	}
	return n
}

var CardVirtues Cards
var CardMap map[string]int

func InitVirtues() {
	CardMap = map[string]int{}
	for i := 0; i <= 52; i++ {
		c := Card{Card: i}
		c.Reveal()
		c.Index = i
		CardVirtues = append(CardVirtues, c)
		CardMap[c.Face] = i
	}
}

// Make hand from faces.
func Make(faces ...string) (hand Cards) {
	for _, f := range faces {
		if m, e := CardMap[f]; e {
			hand = append(hand, CardVirtues[m])
		}
	}
	return
}

// Deck of cards.
type Deck struct {
	Cards    []int
	Rest     int
	Croupier rng.LCPRNG
	Cheats   []string
	Save     []int
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
	deck.Cheats = []string{}
	deck.Save = []int{}
}

func (deck *Deck) Index(c int) int {
	n := -1
	for i, k := range deck.Cards[:deck.Rest] {
		if k == c {
			n = i
			break
		}
	}
	return n
}

func (deck *Deck) AddCheats(cheats ...string) {
	deck.Cheats = append(deck.Cheats, cheats...)
}

func (deck *Deck) Hide(save ...string) {
	for _, s := range save {
		c := CardMap[s]
		n := deck.Index(c)
		if n >= 0 {
			deck.Save = append(deck.Save, c)
			deck.Cards = append(deck.Cards[:n], deck.Cards[n+1:]...)
			deck.Rest--
		}
	}
}

func (deck *Deck) Release() {
	deck.Rest += len(deck.Save)
	deck.Cards = append(deck.Save, deck.Cards...)
	deck.Save = []int{}
}

// Draw single card from deck.
func (deck *Deck) Draw(cheats ...string) (card Card) {
	deck.AddCheats(cheats...)
	if deck.Rest > 0 {
		n := -1

		if len(deck.Cheats) > 0 {
			x := CardMap[deck.Cheats[0]]
			deck.Cheats = deck.Cheats[1:]
			n = deck.Index(x)
		}

		if n < 0 {
			n = deck.Croupier.Choice(deck.Rest)
		}
		c := deck.Cards[n]
		card = CardVirtues[c]
		deck.Rest--
		deck.Cards[deck.Rest], deck.Cards[n] = deck.Cards[n], deck.Cards[deck.Rest]
	} else {
		card.Face = "★"
		card.Index = -1 // error
	}
	return
}

// Deal cards from deck (Fisher-Yates).
func (deck *Deck) Deal(cards int, cheats ...string) (hand Cards) {
	deck.AddCheats(cheats...)
	if cards > 0 {
		hand = make(Cards, cards)
		for i := range hand {
			c := deck.Draw()
			if c.Index >= 0 {
				c.Index = i
			}
			hand[i] = c
		}
	}
	return
}

// Deal cards from whole deck.
func (deck *Deck) NewDeal(cards int) Cards {
	deck.Reset()
	return deck.Deal(cards)
}

// Empty hand.
func (deck *Deck) Null() (hand Cards) {
	return
}

// Croupier with deck of cards.
var Dealer Deck

func init() {
	InitVirtues()
	Dealer.Init()
}
