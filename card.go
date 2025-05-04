package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Suit (マーク)
type Suit int

const (
	Spade Suit = iota
	Heart
	Diamond
	Club
)

func (s Suit) String() string {
	switch s {
	case Spade:
		return "♠"
	case Heart:
		return "♥"
	case Diamond:
		return "♦"
	case Club:
		return "♣"
	default:
		return "?"
	}
}

// Rank (数字)
type Rank int

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return strconv.Itoa(int(r))
	}
}

// Value (カードの点数)
func (r Rank) Value() (int, int) {
	switch r {
	case Ace:
		return 1, 11 // Aceは1または11
	case Jack, Queen, King:
		return 10, 10
	default:
		return int(r), int(r)
	}
}

// Card (カード)
type Card struct {
	Suit Suit
	Rank Rank
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Suit.String(), c.Rank.String())
}

// Deck (デッキ)
type Deck []Card

// NewDeck creates a standard 52-card deck
func NewDeck() Deck {
	deck := make(Deck, 0, 52)
	for s := Spade; s <= Club; s++ {
		for r := Ace; r <= King; r++ {
			deck = append(deck, Card{Suit: s, Rank: r})
		}
	}
	return deck
}

// Shuffle shuffles the deck
func (d Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
}

// Draw draws a card from the deck
func (d *Deck) Draw() (Card, error) {
	if len(*d) == 0 {
		return Card{}, fmt.Errorf("deck is empty")
	}
	card := (*d)[0]
	*d = (*d)[1:]
	return card, nil
}

// Hand (手札)
type Hand []Card

// AddCard adds a card to the hand
func (h *Hand) AddCard(c Card) {
	*h = append(*h, c)
}

// Score calculates the score of the hand
func (h Hand) Score() int {
	lowScore := 0
	aceCount := 0
	for _, card := range h {
		low, _ := card.Rank.Value()
		if card.Rank == Ace {
			aceCount++
			lowScore += low // まずAceを1として計算
		} else {
			lowScore += low // lowとhighはAce以外同じ
		}
	}

	// Aceを11として使えるだけ使う
	for i := 0; i < aceCount; i++ {
		if lowScore+10 <= 21 { // 11にしても21を超えない場合
			lowScore += 10
		}
	}
	return lowScore
}

// String returns the string representation of the hand
func (h Hand) String() string {
	var s []string
	for _, c := range h {
		s = append(s, c.String())
	}
	return strings.Join(s, " ")
}

// IsBust checks if the hand score is over 21
func (h Hand) IsBust() bool {
	return h.Score() > 21
}

// IsBlackjack checks if the hand is a blackjack (2 cards totaling 21)
func (h Hand) IsBlackjack() bool {
	return len(h) == 2 && h.Score() == 21
}
