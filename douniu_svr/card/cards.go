package card

import (
	"sort"
	"douniu/douniu_svr/util"
)

type Cards struct {
	Data 	[]*Card			`json:"data"`
}

//创建一个Cards对象
func NewCards(allCard ...*Card) *Cards{
	cards := &Cards{
		Data :	make([]*Card, 0),
	}
	for _, card := range allCard {
		cards.AddAndSort(card)
	}
	return cards
}

func CreateNewCards(cardSlice []*Card) *Cards{
	newCardSlice := make([]*Card, 0)
	for _, new_card := range cardSlice {
		newCardSlice = append(newCardSlice, new_card)
	}
	return &Cards{
		Data: newCardSlice,
	}
}

//获取cards的数据
func (cards *Cards) GetData() []*Card {
	return cards.Data
}

//获取第idx个牌
func (cards *Cards) At(idx int) *Card {
	if idx >= cards.Len() {
		return nil
	}
	return cards.Data[idx]
}

//cards的长度，牌数
func (cards *Cards) Len() int {
	return len(cards.Data)
}

//比较指定索引对应的两个牌的大小
func (cards *Cards) Less(i, j int) bool {
	cardI := cards.At(i)
	cardJ := cards.At(j)
	if cardI == nil || cardJ == nil{
		return false
	}

	if cardI.CardNo != cardJ.CardNo {
		return cardI.CardNo < cardJ.CardNo
	}

	if cardI.CardType < cardJ.CardType {
		return true
	}
	return false
}

//交换索引为，j的两个数据
func (cards *Cards) Swap(i, j int) {
	if i == j {
		return
	}
	length := cards.Len()
	if i >= length || j >= length {
		return
	}
	swap := cards.At(i)
	cards.Data[i] = cards.At(j)
	cards.Data[j] = swap
}

//追加一张牌
func (cards *Cards) AppendCard(card *Card) {
	if card == nil {
		return
	}
	cards.Data = append(cards.Data, card)
}

//增加一张牌并排序
func (cards *Cards) AddAndSort(card *Card){
	if card == nil {
		return
	}
	cards.AppendCard(card)
	cards.Sort()//default sort
}

//追加一个cards对象
func (cards *Cards) AppendCards(other *Cards) {
	cards.Data = append(cards.Data, other.Data...)
}

//取走一张指定的牌，并返回成功或者失败
func (cards *Cards) TakeWay(drop *Card) bool {
	if drop == nil {
		return true
	}
	for idx, card := range cards.Data {
		if card.SameAs(drop) {
			cards.Data = append(cards.Data[0:idx], cards.Data[idx+1:]...)
			return true
		}
	}
	return false
}

//取走第一张牌
func (cards *Cards) PopFront() *Card {
	if cards.Len() == 0 {
		return nil
	}
	card := cards.At(0)
	cards.Data = cards.Data[1:]
	return card
}

//取走最后一张牌
func (cards *Cards) Tail() *Card {
	if cards.Len() == 0 {
		return nil
	}
	return cards.At(cards.Len()-1)
}

//随机取走一张牌
func (cards *Cards) RandomTakeWayOne() *Card {
	length := cards.Len()
	if length == 0 {
		return nil
	}
	idx := util.RandomN(length)
	card := cards.At(idx)
	cards.Data = append(cards.Data[0:idx], cards.Data[idx+1:]...)
	return card
}

//清空牌
func (cards *Cards) Clear() {
	cards.Data = cards.Data[0:0]
}

//排序
func (cards *Cards)Sort() {
	sort.Sort(cards)
}

//是否是一样的牌组
func (cards *Cards) SameAs(other *Cards) bool {
	if cards == nil || other == nil {
		return false
	}

	length := other.Len()
	if cards.Len() != length {
		return false
	}

	for idx := 0; idx < length; idx++ {
		if !cards.At(idx).SameAs(other.At(idx)) {
			return false
		}
	}
	return true
}

func (cards *Cards) String() string {
	str := ""
	for _, card := range cards.Data{
		str += card.String() + ","
	}
	return str
}

//检查是否存在子集subCards
func (cards *Cards) hasCards(subCards *Cards) bool {
	if subCards.Len() == 0 {
		return true
	}
	tmpCards := CreateNewCards(cards.GetData())
	for _, subCard := range subCards.Data {
		if !tmpCards.HasCard(subCard) {
			return false
		}
		tmpCards.TakeWay(subCard)
	}

	return true
}

func (cards *Cards) HasCard(card *Card) bool{
	for _, tmp := range cards.Data {
		if tmp.SameAs(card) {
			return true
		}
	}
	return false
}
