package card

type Card struct {
	CardType int 	//牌类型
	CardNo   int 	//牌编号
	CardId   int    //牌唯一标识符
}

//是否同一类型的牌
func (card *Card) SameCardTypeAs(other *Card) bool {
	if other == nil || card == nil {
		return false
	}
	return other.CardType == card.CardType
}

func (card *Card) SameCardNoAs(other *Card) bool {
	if other == nil || card == nil {
		return false
	}
	return other.CardNo == card.CardNo
}

func (card *Card) SameAs(other *Card) bool {
	if other == nil || card == nil {
		return false
	}
	if other.CardType != card.CardType {
		return false
	}
	if other.CardNo != card.CardNo {
		return false
	}
	return true
}

func (card *Card) MakeKey() int64 {
	var ret int64
	ret = int64(card.CardNo << 48) | int64(card.CardType << 32) | int64(card.CardId)
	return ret
}

func (card *Card) MakeID(num int) int {
	var ret int
	ret = card.CardNo * 100 + card.CardType * 10 + num
	return ret
}


func (card *Card) Next() *Card {
	if card == nil {
		return nil
	}
	if card.CardNo == 13 {
		return nil
	}
	return &Card{
		CardType: card.CardType,
		CardNo: card.CardNo + 1,
	}
}

func (card *Card) Prev() *Card {
	if card == nil {
		return nil
	}
	if card.CardNo == 1 {
		return nil
	}
	return &Card{
		CardType: card.CardType,
		CardNo: card.CardNo - 1,
	}
}

func (card *Card) String() string {
	if card == nil {
		return "nil"
	}
	cardNameMap := cardNameMap()
	noNameMap, ok1 := cardNameMap[card.CardType]
	if !ok1 {
		return "unknow card type"
	}

	name, ok2 := noNameMap[card.CardNo]
	if !ok2 {
		return "unknow card no"
	}
	return name
}

func (card *Card) GetScore() int{
	switch card.CardNo {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	case 6:
		return 6
	case 7:
		return 7
	case 8:
		return 8
	case 9:
		return 9
	case 10:
		return 10
	case 11:
		return 10
	case 12:
		return 10
	case 13:
		return 10
	}
	return 1
}

func cardNameMap() map[int]map[int]string {
	return map[int]map[int]string{
		CardType_Fangpian: {
			1: 		"A方片",
			2:  		"2方片",
			3:   		"3方片",
			4:  		"4方片",
			5:		"5方片",
			6:		"6方片",
			7:		"7方片",
			8:		"8方片",
			9:		"9方片",
			10:		"10方片",
			11:		"J方片",
			12:		"Q方片",
			13:		"K方片",
		},
		CardType_Meihua: {
			1: 		"A梅花",
			2:  		"2梅花",
			3:   		"3梅花",
			4:  		"4梅花",
			5:		"5梅花",
			6:		"6梅花",
			7:		"7梅花",
			8:		"8梅花",
			9:		"9梅花",
			10:		"10梅花",
			11:		"J梅花",
			12:		"Q梅花",
			13:		"K梅花",
		},
		CardType_Hongtao: {
			1: 		"A红桃",
			2:  		"2红桃",
			3:   		"3红桃",
			4:  		"4红桃",
			5:		"5红桃",
			6:		"6红桃",
			7:		"7红桃",
			8:		"8红桃",
			9:		"9红桃",
			10:		"10红桃",
			11:		"J红桃",
			12:		"Q红桃",
			13:		"K红桃",
		},
		CardType_Heitao: {
			1: 		"A黑桃",
			2:  		"2黑桃",
			3:   		"3黑桃",
			4:  		"4黑桃",
			5:		"5黑桃",
			6:		"6黑桃",
			7:		"7黑桃",
			8:		"8黑桃",
			9:		"9黑桃",
			10:		"10黑桃",
			11:		"J黑桃",
			12:		"Q黑桃",
			13:		"K黑桃",
		},
		CardType_Xiaowang: {
			14:		"小王",
		},
		CardType_Dawang: {
			14:		"大王",
		},
	}
}
