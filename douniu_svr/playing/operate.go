package playing

import (
	"fmt"
	"douniu/douniu_svr/card"
)

type OperateType int

const (
	OperateEnterRoom	OperateType = iota + 1
	OperateReadyRoom
	OperateLeaveRoom

	OperateScramble
	OperateBet
	OperateShowCards
	OperateSeeCards
)

func (operateType OperateType) String() string {
	switch operateType {
	case OperateEnterRoom :
		return "OperateEnterRoom"
	case OperateReadyRoom :
		return "OperateReadyRoom"
	case OperateLeaveRoom:
		return "OperateLeaveRoom"
	case OperateScramble:
		return "OperateScramble"
	case OperateBet:
		return "OperateBet"
	case OperateShowCards:
		return "OperateShowCards"
	case OperateSeeCards:
		return "OperateSeeCards"
	}
	return "unknow OperateType"
}

type Operate struct {//玩家操作
	Op			OperateType
	Operator	*Player				//操作者
	Data		interface{}
	ResultCh		chan bool
}

func (op *Operate) String() string {
	if op == nil {
		return "{operator=nil, op=nil}"
	}
	return fmt.Sprintf("{operator=%v, op=%v}", op.Operator, op.Op)
}

func newOperate(op OperateType, operator *Player, data interface{}) *Operate{
	return &Operate{
		Op:	op,
		Data: data,
		Operator: operator,
		ResultCh: make(chan bool, 1),
	}
}

type OperateEnterRoomData struct {
}
func NewOperateEnterRoom(operator *Player, data *OperateEnterRoomData) *Operate {
	return newOperate(OperateEnterRoom, operator, data)
}

type OperateReadyRoomData struct {
}
func NewOperateReadyRoom(operator *Player, data *OperateReadyRoomData) *Operate {
	return newOperate(OperateReadyRoom, operator, data)
}

type OperateLeaveRoomData struct {
}
func NewOperateLeaveRoom(operator *Player, data *OperateLeaveRoomData) *Operate {
	return newOperate(OperateLeaveRoom, operator, data)
}

type OperateScrambleData struct {
	ScrambleMultiple int32
}
func NewOperateScramble(operator *Player, data *OperateScrambleData) *Operate {
	return newOperate(OperateScramble, operator, data)
}

type OperateBetData struct {
	Score int32
}
func NewOperateBet(operator *Player, data *OperateBetData) *Operate {
	return newOperate(OperateBet, operator, data)
}

type OperateShowCardsData struct {
	Paixing int
	PaixingMultiple int
	PlayingCards	*card.PlayingCards
	NiuCards	[]*card.Card
}
func NewOperateShowCards(operator *Player, data *OperateShowCardsData) *Operate {
	return newOperate(OperateShowCards, operator, data)
}

type OperateSeeCardsData struct {}
func NewOperateSeeCards(operator *Player, data *OperateSeeCardsData) *Operate {
	return newOperate(OperateSeeCards, operator, data)
}
