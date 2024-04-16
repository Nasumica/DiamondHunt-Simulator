package main

// Author: Srbislav D. Nešić, srbislav.nesic@fincore.com

import (
	"crypto/rand"
	"math/big"
)

type Xaoc struct{} // to slow

func (rnd *Xaoc) Choice(n int) int {
	if n > 1 {
		b, _ := rand.Int(rand.Reader, new(big.Int).SetUint64(uint64(n)))
		n = int(b.Uint64())
	} else {
		n--
	}
	return n
}

func (rnd *Xaoc) Deck() (d []int) {
	d = make(list, 52)
	for i := range d {
		j := rnd.Choice(i + 1)
		d[i], d[j] = d[j], i+1
	}
	return
}
