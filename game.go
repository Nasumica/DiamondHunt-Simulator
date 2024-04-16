package main

import (
	"fmt"
	"time"
)

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
	h := s.Hand[s.Swaped]
	if h.Load > d.Load {
		s.Swap(s.Swaped)
		s.Swaped++
	}
}

func (s *Screen) Play() []Card {
	s.Deal()
	s.Hunt()
	s.Hunt()
	s.Hunt()
	s.Hunt()
	return s.Diam
}

func SpeedTest(n int) {
	start := time.Now()
	var scr Screen
	for i := 0; i < n; i++ {
		scr.Play()
	}
	elapsed := time.Since(start).Seconds()
	speed := float64(n) / elapsed
	fmt.Printf("elapsed = %.3f\",  speed = %.0f deals / s\n", elapsed, speed)
}
