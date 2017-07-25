package playing

import (
	"douniu/douniu_svr/card"
	"douniu/douniu_svr/log"
	"time"
	"fmt"
)

type PlayerObserver interface {
	OnMsg(player *Player, msg *Message)
}

type Player struct {
	id			uint64			//玩家id
	position		int			//玩家在房间的位置
	room			*Room			//玩家所在的房间
	isReady			bool
	xiazhuScore             int32                   //每局下注分数

	playingCards 	*card.PlayingCards	//玩家手上的牌
	observers	 []PlayerObserver
}

func NewPlayer(id uint64) *Player {
	player :=  &Player{
		id:		id,
		position:       0,
		playingCards:	card.NewPlayingCards(),
		observers:	make([]PlayerObserver, 0),
	}
	return player
}

func (player *Player) GetXiazhuScore() int32 {
	return player.xiazhuScore
}

func (player *Player) SetXiazhuScore(score int32) {
	player.xiazhuScore = score
}

func (player *Player) Reset() {
	log.Debug(time.Now().Unix(), player,"Player.Reset")
	player.playingCards.Reset()
	player.isReady = false
}

func (player *Player) AddObserver(ob PlayerObserver) {
	player.observers = append(player.observers, ob)
}

func (player *Player) AddCard(card *card.Card) {
	log.Debug(time.Now().Unix(), player, "Player.AddCard :", card)
	player.playingCards.AddCard(card)
}

func (player *Player) OperateEnterRoom(room *Room) bool{
	log.Debug(time.Now().Unix(), player, "OperateEnterRoom room :", room)
	data := &OperateEnterRoomData{}
	op := NewOperateEnterRoom(player, data)
	room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateLeaveRoom() bool{
	log.Debug(time.Now().Unix(), player, "OperateLeaveRoom", player.room)
	if player.room == nil {
		return true
	}

	data := &OperateLeaveRoomData{}
	op := NewOperateLeaveRoom(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateDoReady() bool{
	log.Debug(time.Now().Unix(), player, "OperateDoReady", player.room)
	if player.room == nil {
		return false
	}

	data := &OperateReadyRoomData{}
	op := NewOperateReadyRoom(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) GetIsReady() bool {
	return player.isReady
}

func (player *Player) SetIsReady(is_ready bool) {
	player.isReady = is_ready
}

func (player *Player) GetPlayingCards() *card.PlayingCards {
	return player.playingCards
}

func (player *Player) waitResult(resultCh chan bool) bool{
	log_time := time.Now().Unix()
	select {
	case <- time.After(time.Second * 10):
		close(resultCh)
		log.Debug(time.Now().Unix(), player, "Player.waitResult timeout")
		return false
	case result := <- resultCh:
		log.Debug(log_time, player, "Player.waitResult result :", result)
		return result
	}
	log.Debug(log_time, player, "Player.waitResult fasle")
	return false
}

func (player *Player) EnterRoom(room *Room, position int) {
	log.Debug(time.Now().Unix(), player, "enter", room)
	player.room = room
	player.position = position
}

func (player *Player) ReadyRoom(room *Room) {
	log.Debug(time.Now().Unix(), player, "ready", room)
}

func (player *Player) LeaveRoom() {
	log.Debug(time.Now().Unix(), player, "leave", player.room)
	player.room = nil
	player.position = -1
}

func (player *Player) String() string{
	if player == nil {
		return "{player=nil}"
	}
	return fmt.Sprintf("{player=%v}", player.id)
}

//玩家成功操作的通知
func (player *Player) OnPlayerSuccessOperated(op *Operate) {
	log.Debug(time.Now().Unix(), player, "OnPlayerSuccessOperated", op)
	switch op.Op {
	case OperateEnterRoom:
		player.onPlayerEnterRoom(op)
	case OperateLeaveRoom:
		player.onPlayerLeaveRoom(op)
	}
}

func (player *Player) notifyObserver(msg *Message) {
	log.Debug(time.Now().Unix(), player, "notifyObserverMsg", msg)
	for _, ob := range player.observers {
		ob.OnMsg(player, msg)
	}
}

//begin player operate event handler

func (player *Player) onPlayerEnterRoom(op *Operate) {
	if _, ok := op.Data.(*OperateEnterRoomData); ok {
		if player.room == nil {
			return
		}
		msgData := &EnterRoomMsgData{
			EnterPlayer : op.Operator,
			AllPlayer: player.room.players,
		}
		player.notifyObserver(NewEnterRoomMsg(player, msgData))
	}
}

func (player *Player) onPlayerLeaveRoom(op *Operate) {
	if _, ok := op.Data.(*OperateLeaveRoomData); ok {
		if op.Operator == player {
			return
		}
		if player.room == nil {
			return
		}
		msgData := &LeaveRoomMsgData{
			LeavePlayer : op.Operator,
			AllPlayer: player.room.players,
		}
		player.notifyObserver(NewLeaveRoomMsg(player, msgData))
	}
}

func (player *Player) OnGetInitCards() {
	log.Debug(time.Now().Unix(), player, "OnGetInitCards", player.playingCards)

	data := &GetInitCardsMsgData{
		PlayingCards: player.playingCards,
	}
	player.notifyObserver(NewGetInitCardsMsg(player, data))
}

func (player *Player) OnConfirmMaster() {
	log.Debug(time.Now().Unix(), player, "OnConfirmMaster")

	data := &ConfirmMasterMsgData{
	}
	player.notifyObserver(NewConfirmMasterMsg(player, data))
}
