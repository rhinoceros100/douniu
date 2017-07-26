package playing

import (
	"fmt"
	"douniu/douniu_svr/card"
)

type MsgType	int
const  (
	MsgGetInitCards	MsgType = iota + 1
	MsgGetMaster
	MsgBet
	MsgShowCards

	MsgEnterRoom
	MsgLeaveRoom
	MsgGameEnd
	MsgRoomClosed
)

func (msgType MsgType) String() string {
	switch msgType {
	case MsgGetInitCards:
		return "MsgGetInitCards"
	case MsgGetMaster:
		return "MsgGetMaster"
	case MsgBet:
		return "MsgBet"
	case MsgShowCards:
		return "MsgShowCards"
	case MsgEnterRoom:
		return "MsgEnterRoom"
	case MsgLeaveRoom:
		return "MsgEnterRoom"
	case MsgGameEnd:
		return "MsgGameEnd"
	case MsgRoomClosed:
		return "MsgRoomClosed"
	}
	return "unknow MsgType"
}

type Message struct {
	Type		MsgType
	Owner 	*Player
	Data 	interface{}
}

func (data *Message) String() string {
	if data == nil {
		return "{nil Message}"
	}
	return fmt.Sprintf("{type=%v, Owner=%v}", data.Type, data.Owner)
}

func newMsg(t MsgType, owner *Player, data interface{}) *Message {
	return &Message{
		Owner:	owner,
		Type: t,
		Data: data,
	}
}

//玩家获得初始牌的消息
type GetMasterMsgData struct {
	Scores   []int32
}
func NewGetMasterMsg(owner *Player, data *GetMasterMsgData) *Message {
	return newMsg(MsgGetMaster, owner, data)
}

//玩家获得初始牌的消息
type GetInitCardsMsgData struct {
	PlayingCards	*card.PlayingCards
}
func NewGetInitCardsMsg(owner *Player, data *GetInitCardsMsgData) *Message {
	return newMsg(MsgGetInitCards, owner, data)
}

//玩家下注的消息
type BetMsgData struct {}
func NewBetMsg(owner *Player, data *BetMsgData) *Message {
	return newMsg(MsgBet, owner, data)
}

//玩家亮牌的消息
type ShowCardsMsgData struct {}
func NewShowCardsMsg(owner *Player, data *ShowCardsMsgData) *Message {
	return newMsg(MsgShowCards, owner, data)
}

//玩家进入房间的消息
type EnterRoomMsgData struct {
	EnterPlayer *Player
	AllPlayer 	[]*Player
}
func NewEnterRoomMsg(owner *Player, data *EnterRoomMsgData) *Message {
	return newMsg(MsgEnterRoom, owner, data)
}

//玩家离开房间的消息
type LeaveRoomMsgData struct {
	LeavePlayer *Player
	AllPlayer 	[]*Player
}
func NewLeaveRoomMsg(owner *Player, data *LeaveRoomMsgData) *Message {
	return newMsg(MsgLeaveRoom, owner, data)
}

//一盘游戏结束的消息
type GameEndMsgData struct {}
func NewGameEndMsg(owner *Player, data *GameEndMsgData) *Message{
	return newMsg(MsgGameEnd, owner, data)
}

//房间结束的消息
type RoomClosedMsgData struct {}
func NewRoomClosedMsg(owner *Player, data *RoomClosedMsgData) *Message{
	return newMsg(MsgRoomClosed, owner, data)
}
