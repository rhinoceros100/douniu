package card

import "fmt"

const NIUNIU_INIT_CARD_NUM int = 5      //牛牛起手牌数

type PlayingCards struct {
	CardsInHand			*Cards		//手上的牌
}

func NewPlayingCards() *PlayingCards {
	return  &PlayingCards{
		CardsInHand: NewCards(),
	}
}

func (playingCards *PlayingCards) Reset() {
	playingCards.CardsInHand.Clear()
}

func (playingCards *PlayingCards) AddCards(cards *Cards) {
	playingCards.CardsInHand.AppendCards(cards)
	playingCards.CardsInHand.Sort()
}

//增加一张牌
func (playingCards *PlayingCards) AddCard(card *Card) {
	playingCards.CardsInHand.AddAndSort(card)
}

func (playingCards *PlayingCards) String() string{
	return fmt.Sprintf(
		"{%v}",
		playingCards.CardsInHand,
	)
}
