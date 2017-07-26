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
	position		int32			//玩家在房间的位置
	room			*Room			//玩家所在的房间
	isReady			bool
	isBet			bool
	isShow			bool
	BetScore             int32                   //每局下注分数

	playingCards 	*card.PlayingCards	//玩家手上的牌
	observers	 []PlayerObserver
}

func NewPlayer(id uint64) *Player {
	player :=  &Player{
		id:		id,
		position:       10,
		isReady:        false,
		isBet:       false,
		isShow:     false,
		playingCards:	card.NewPlayingCards(),
		observers:	make([]PlayerObserver, 0),
	}
	return player
}

func (player *Player) IsMaster() bool {
	return player == player.room.masterPlayer
}

func (player *Player) GetId() uint64 {
	return player.id
}

func (player *Player) GetPosition() int32 {
	return player.position
}

func (player *Player) GetIsBet() bool {
	return player.isBet
}

func (player *Player) SetIsBet(is_bet bool) {
	player.isBet = is_bet
}

func (player *Player) GetBetScore() int32 {
	return player.BetScore
}

func (player *Player) SetBetScore(score int32) {
	player.BetScore = score
}

func (player *Player) GetIsShow() bool {
	return player.isShow
}

func (player *Player) SetIsShow(is_show bool) {
	player.isShow = is_show
}

func (player *Player) Reset() {
	log.Debug(time.Now().Unix(), player,"Player.Reset")
	player.playingCards.Reset()
	player.SetIsReady(false)
	player.SetIsBet(false)
	player.SetIsShow(false)
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

func (player *Player) OperateBet(score int32) bool{
	log.Debug(time.Now().Unix(), player, "OperateBet", player.room)
	if player.room == nil {
		return false
	}

	data := &OperateBetData{Score:score}
	op := NewOperateBet(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateShowCards() bool{
	log.Debug(time.Now().Unix(), player, "OperateShowCards", player.room)
	if player.room == nil {
		return false
	}

	data := &OperateShowCardsData{}
	op := NewOperateShowCards(player, data)
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

func (player *Player) EnterRoom(room *Room, position int32) {
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

func (player *Player) Bet(score int32) {
	log.Debug(time.Now().Unix(), player, "bet", player.room)
	player.SetBetScore(score)
	player.SetIsBet(true)
}

func (player *Player) ShowCards() {
	log.Debug(time.Now().Unix(), player, "showcards", player.room)
	player.SetIsShow(true)
}

func (player *Player) String() string{
	if player == nil {
		return "{player=nil}"
	}
	return fmt.Sprintf("{player=%v, pos=%v}", player.id, player.position)
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

func (player *Player) OnAllBet() {
	log.Debug(time.Now().Unix(), player, "OnAllBet")

	data := &BetMsgData{}
	player.notifyObserver(NewBetMsg(player, data))
}

func (player *Player) OnAllShowCards() {
	log.Debug(time.Now().Unix(), player, "OnAllShowCards")

	data := &ShowCardsMsgData{}
	player.notifyObserver(NewShowCardsMsg(player, data))
}

func (player *Player) OnGetInitCards() {
	log.Debug(time.Now().Unix(), player, "OnGetInitCards", player.playingCards)

	data := &GetInitCardsMsgData{
		PlayingCards: player.playingCards,
	}
	player.notifyObserver(NewGetInitCardsMsg(player, data))
}

func (player *Player) OnGetMaster() {
	log.Debug(time.Now().Unix(), player, "OnGetMaster")

	data := &GetMasterMsgData{}
	data.Scores = make([]int32, 0)
	data.Scores = append(data.Scores, player.room.GetScoreLow())
	data.Scores = append(data.Scores, player.room.GetScoreHigh())
	log.Debug(time.Now().Unix(), player, data.Scores)
	player.notifyObserver(NewGetMasterMsg(player, data))
}

func (player *Player) OnRoomClosed() {
	log.Debug(time.Now().Unix(), player, "OnRoomClosed")
	player.room = nil
	//player.Reset()

	data := &RoomClosedMsgData{}
	player.notifyObserver(NewRoomClosedMsg(player, data))
}

func (player *Player) OnEndPlayGame() {
	log.Debug(time.Now().Unix(), player, "OnPlayingGameEnd")
	//player.Reset()
	data := &GameEndMsgData{}
	player.notifyObserver(NewGameEndMsg(player, data))
}
