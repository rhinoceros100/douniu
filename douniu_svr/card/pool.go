package card

import "douniu/douniu_svr/util"

type Pool struct {
	cards *Cards
}

func NewPool() *Pool {
	pool := &Pool{
		cards:	NewCards(),
	}
	return pool
}

func (pool *Pool) generate() {
	for cardNo := 1 ; cardNo <= 13; cardNo ++ {
		for num := 0; num < 1; num ++ {
			card_fang := &Card{
				CardType_Fangpian,
				cardNo,
				0,
			}
			card_fang.CardId = card_fang.MakeID(num)

			card_mei := &Card{
				CardType_Meihua,
				cardNo,
				0,
			}
			card_mei.CardId = card_mei.MakeID(num)

			card_hong := &Card{
				CardType_Hongtao,
				cardNo,
				0,
			}
			card_hong.CardId = card_hong.MakeID(num)

			card_hei := &Card{
				CardType_Heitao,
				cardNo,
				0,
			}
			card_hei.CardId = card_hei.MakeID(num)

			pool.cards.AppendCard(card_fang)
			pool.cards.AppendCard(card_mei)
			pool.cards.AppendCard(card_hong)
			pool.cards.AppendCard(card_hei)
		}
	}
}

func (pool *Pool) ReGenerate() {
	pool.cards.Clear()
	pool.generate()
	pool.shuffle()
}

//洗牌，打乱牌
func (pool *Pool) shuffle() {
	length := pool.cards.Len()
	for cnt := 0; cnt < length*3; cnt++ {
		i := util.RandomN(length)
		j := util.RandomN(length)
		pool.cards.Swap(i, j)
	}
}

func (pool *Pool) PopFront() *Card {
	return pool.cards.PopFront()
}

func (pool *Pool) At(idx int) *Card {
	return pool.cards.At(idx)
}

func (pool *Pool) GetCardNum() int {
	return pool.cards.Len()
}
