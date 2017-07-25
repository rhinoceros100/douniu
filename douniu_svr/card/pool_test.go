package card

import (
	"testing"
	"time"
	"github.com/bmizerany/assert"
)

func TestPool(t *testing.T) {
	start := time.Now()
	pool := NewPool()
	pool.ReGenerate()
	t.Log(pool.cards)

	t.Log(pool.GetCardNum())
	beforeGet := NewCards()
	beforeGet.AppendCards(pool.cards)
	newCards := NewCards()
	for{
		card := pool.PopFront()
		if card == nil {
			break
		}
		t.Log(card, card.CardId, pool.GetCardNum())
		newCards.AppendCard(card)
		//time.Sleep(time.Second)
	}
	if pool.GetCardNum() != 0 {
		t.Fatal("card num of pool should be 0")
	}

	t.Log("time:", time.Now().Sub(start).Seconds())

	assert.Equal(t, newCards.SameAs(beforeGet), true)
}
