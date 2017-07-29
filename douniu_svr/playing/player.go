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
	isScramble		bool
	isBet			bool
	isShowCards		bool
	isSeeCards		bool
	isPlaying		bool
	BetScore             	int32                   //每局下注分数
	BaseMultiple            int32                   //抢庄倍数
	Paixing		        int                     //牌型
	PaixingMultiple	        int                     //牌型倍数
	maxid		        int                     //手牌最大的id
	roundScore              int32                   //本轮得分
	totalScore              int32                   //总得分
	leizhu		        int32                   //垒注

	playingCards 	*card.PlayingCards	//玩家手上的牌
	niuCards         []*card.Card
	observers	 []PlayerObserver
}

func NewPlayer(id uint64) *Player {
	player :=  &Player{
		id:		id,
		position:       10,
		isReady:        false,
		isScramble:     false,
		isBet:      	false,
		isShowCards:   	false,
		isSeeCards:   	false,
		isPlaying:     	false,
		BetScore:     	1,
		BaseMultiple:   1,
		maxid:   1,
		roundScore:     0,
		totalScore:     0,
		Paixing:   	card.DouniuType_Meiniu,
		PaixingMultiple:1,
		playingCards:	card.NewPlayingCards(),
		observers:	make([]PlayerObserver, 0),
		niuCards:       make([]*card.Card, 0),
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

func (player *Player) GetTotalScore() int32 {
	return player.totalScore
}

func (player *Player) AddTotalScore(add int32) int32 {
	player.totalScore += add
	return player.totalScore
}

func (player *Player) GetRoundScore() int32 {
	return player.roundScore
}

func (player *Player) SetRoundScore(round_score int32) {
	player.roundScore = round_score
}

func (player *Player) GetLeizhu() int32 {
	return player.leizhu
}

func (player *Player) SetLeizhu(leizhu int32) {
	player.leizhu = leizhu
}

func (player *Player) GetIsScramble() bool {
	return player.isScramble
}

func (player *Player) SetIsScramble(is_scramble bool) {
	player.isScramble = is_scramble
}

func (player *Player) GetBaseMultiple() int32 {
	return player.BaseMultiple
}

func (player *Player) SetBaseMultiple(multiple int32) {
	player.BaseMultiple = multiple
}

func (player *Player) GetPaixingMultiple() int {
	return player.PaixingMultiple
}

func (player *Player) SetPaixingMultiple(multiple int) {
	player.PaixingMultiple = multiple
}

func (player *Player) GetPaixing() int {
	return player.Paixing
}

func (player *Player) SetPaixing(paixing int) {
	player.Paixing = paixing
}

func (player *Player) GetMaxid() int {
	return player.maxid
}

func (player *Player) SetMaxid(maxid int) {
	player.maxid = maxid
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

func (player *Player) GetIsShowCards() bool {
	return player.isShowCards
}

func (player *Player) SetIsShowCards(is_showcards bool) {
	player.isShowCards = is_showcards
}

func (player *Player) GetIsSeeCards() bool {
	return player.isSeeCards
}

func (player *Player) SetIsSeeCards(is_seecards bool) {
	player.isSeeCards = is_seecards
}

func (player *Player) GetIsPlaying() bool {
	return player.isPlaying
}

func (player *Player) SetIsPlaying(is_playing bool) {
	player.isPlaying = is_playing
}

func (player *Player) GetNiuCards() []*card.Card {
	return player.niuCards
}

func (player *Player) SetNiuCards(niu_cards []*card.Card) {
	player.niuCards = niu_cards
}

func (player *Player) Reset() {
	//log.Debug(time.Now().Unix(), player,"Player.Reset")
	player.playingCards.Reset()
	player.SetIsReady(false)
	player.SetIsBet(false)
	player.SetIsScramble(false)
	player.SetIsShowCards(false)
	player.SetIsSeeCards(false)
	player.SetIsPlaying(false)
}

func (player *Player) AddObserver(ob PlayerObserver) {
	player.observers = append(player.observers, ob)
}

func (player *Player) AddCard(card *card.Card) {
	//log.Debug(time.Now().Unix(), player, "Player.AddCard :", card)
	player.playingCards.AddCard(card)
}

func (player *Player) OperateEnterRoom(room *Room) bool{
	//log.Debug(time.Now().Unix(), player, "OperateEnterRoom room :", room)
	for _, room_player := range room.players{
		if room_player == player{
			log.Error("Player already in room:", player)
			return false
		}
	}

	data := &OperateEnterRoomData{}
	op := NewOperateEnterRoom(player, data)
	room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateLeaveRoom() bool{
	//log.Debug(time.Now().Unix(), player, "OperateLeaveRoom", player.room)
	if player.room == nil {
		return true
	}
	room_status := player.room.roomStatus
	if room_status > RoomStatusWaitAllPlayerEnter {
		log.Error("Wrong room status:", room_status)
		return false
	}

	data := &OperateLeaveRoomData{}
	op := NewOperateLeaveRoom(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateDoReady() bool{
	//log.Debug(time.Now().Unix(), player, "OperateDoReady", player.room)
	if player.room == nil || player.GetIsReady(){
		return false
	}
	room_status := player.room.roomStatus
	if room_status != RoomStatusWaitAllPlayerEnter && room_status != RoomStatusWaitAllPlayerReady {
		log.Error("Wrong room status:", room_status)
		return false
	}

	data := &OperateReadyRoomData{}
	op := NewOperateReadyRoom(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateScramble(scramble_multiple int32) bool{
	log.Debug(time.Now().Unix(), player, "OperateScramble", player.room)
	if player.room == nil || player.GetIsScramble(){
		return false
	}
	room_status := player.room.roomStatus
	if room_status != RoomStatusGetMaster {
		log.Error("Wrong room status:", room_status)
		return false
	}
	if !player.GetIsPlaying() {
		log.Error("Player is not playing", player)
		return false
	}

	data := &OperateScrambleData{ScrambleMultiple:scramble_multiple}
	op := NewOperateScramble(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateBet(score int32) bool{
	log.Debug(time.Now().Unix(), player, "OperateBet", player.room)
	if player.room == nil  || player.GetIsBet(){
		return false
	}
	room_status := player.room.roomStatus
	if room_status != RoomStatusPlayGame {
		log.Error("Wrong room status:", room_status)
		return false
	}
	if !player.GetIsPlaying() {
		log.Error("Player is not playing", player)
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
	room_status := player.room.roomStatus
	if room_status != RoomStatusShowCards {
		log.Error("Wrong room status:", room_status)
		return false
	}
	if !player.GetIsPlaying() {
		log.Error("Player is not playing", player)
		return false
	}

	data := &OperateShowCardsData{}
	op := NewOperateShowCards(player, data)
	player.room.PlayerOperate(op)
	return player.waitResult(op.ResultCh)
}

func (player *Player) OperateSeeCards() bool{
	log.Debug(time.Now().Unix(), player, "OperateSeeCards", player.room)
	if player.room == nil {
		return false
	}
	room_status := player.room.roomStatus
	if room_status != RoomStatusShowCards {
		log.Error("Wrong room status:", room_status)
		return false
	}
	if !player.GetIsPlaying() {
		log.Error("Player is not playing", player)
		return false
	}

	data := &OperateSeeCardsData{}
	op := NewOperateSeeCards(player, data)
	player.room.dealPlayerOperate(op)
	return true
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

func (player *Player) Scramble(multiple int32) {
	log.Debug(time.Now().Unix(), player, "Scramble", player.room)
	player.SetBaseMultiple(multiple)
	player.SetIsScramble(true)
}

func (player *Player) ShowCards() {
	//log.Debug(time.Now().Unix(), player, "showcards", player.room)
	player.SetIsShowCards(true)

	paixing, niu_cards := card.GetPaixing(player.playingCards.CardsInHand.GetData())
	maxid := card.GetCardsMaxid(player.playingCards.CardsInHand.GetData())
	paixing_multiple := card.GetPaixingMultiple(paixing)
	player.SetPaixing(paixing)
	player.SetPaixingMultiple(paixing_multiple)
	player.SetMaxid(maxid)
	player.SetNiuCards(niu_cards)
}

func (player *Player) SeeCards() {
	log.Debug(time.Now().Unix(), player, "seecards", player.room)
	player.SetIsSeeCards(true)
}

func (player *Player) String() string{
	if player == nil {
		return "{player=nil}"
	}
	return fmt.Sprintf("{player=%v, pos=%v}", player.id, player.position)
}

//玩家成功操作的通知
func (player *Player) OnPlayerSuccessOperated(op *Operate) {
	//log.Debug(time.Now().Unix(), player, "OnPlayerSuccessOperated", op)
	switch op.Op {
	case OperateEnterRoom:
		player.onPlayerEnterRoom(op)
	case OperateLeaveRoom:
		player.onPlayerLeaveRoom(op)
	case OperateReadyRoom:
		player.onPlayerReadyRoom(op)
	case OperateShowCards:
		player.onShowCards(op)
	case OperateSeeCards:
		player.onSeeCards(op)
	}
}

func (player *Player) notifyObserver(msg *Message) {
	//log.Debug(time.Now().Unix(), player, "notifyObserverMsg", msg)
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

func (player *Player) onPlayerReadyRoom(op *Operate) {
	//log.Debug(time.Now().Unix(), player, "onPlayerReadyRoom")

	data := &ReadyRoomMsgData{
		ReadyPlayer:op.Operator,
	}
	player.notifyObserver(NewReadyRoomMsg(player, data))
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
	//log.Debug(time.Now().Unix(), player, "OnAllBet")

	data := &BetMsgData{}
	player.notifyObserver(NewBetMsg(player, data))
}

func (player *Player) OnJiesuan(msg *Message) {
	//log.Debug(time.Now().Unix(), player, "OnJiesuan")

	player.notifyObserver(msg)
}

func (player *Player) onShowCards(op *Operate) {
	//log.Debug(time.Now().Unix(), player, "onShowCards")
	if show_data, ok := op.Data.(*OperateShowCardsData); ok {
		data := &ShowCardsMsgData{
			ShowPlayer:op.Operator,
			Paixing:show_data.Paixing,
			PaixingMultiple:show_data.PaixingMultiple,
			PlayingCards:show_data.PlayingCards,
			NiuCards:show_data.NiuCards,
		}
		player.notifyObserver(NewShowCardsMsg(player, data))
	}
}

func (player *Player) onSeeCards(op *Operate) {
	//log.Debug(time.Now().Unix(), player, "onSeeCards")

	data := &SeeCardsMsgData{
		SeePlayer:op.Operator,
	}
	player.notifyObserver(NewSeeCardsMsg(player, data))
}

func (player *Player) OnGetInitCards() {
	//log.Debug(time.Now().Unix(), player, "OnGetInitCards", player.playingCards)

	data := &GetInitCardsMsgData{
		PlayingCards: player.playingCards,
	}
	player.notifyObserver(NewGetInitCardsMsg(player, data))
}

func (player *Player) OnGetDispatchedCard(dispatched_card *card.Card) {
	//log.Debug(time.Now().Unix(), player, "OnGetDispatchedCard", dispatched_card)

	data := &DispatchCardMsgData{
		DispatchedCard: dispatched_card,
	}
	player.notifyObserver(NewDispatchCardMsg(player, data))
}

func (player *Player) OnGetMaster(highest_players []*Player, master_player *Player) {
	//log.Debug(time.Now().Unix(), player, "OnGetMaster")
	//player.SetIsPlaying(true)

	data := &GetMasterMsgData{
		MasterPlayer:master_player,
		HighestPlayers:highest_players,
	}
	data.Scores = make([]int32, 0)
	if player != player.room.masterPlayer {
		data.Scores = append(data.Scores, player.room.GetScoreLow())
		data.Scores = append(data.Scores, player.room.GetScoreHigh())
		lastLeizhu := player.GetLeizhu()
		if lastLeizhu > 0 {
			data.Scores = append(data.Scores, lastLeizhu)
		}
	}

	//log.Debug(time.Now().Unix(), player, data.Scores)
	player.notifyObserver(NewGetMasterMsg(player, data))
}

func (player *Player) OnRoomClosed() {
	//log.Debug(time.Now().Unix(), player, "OnRoomClosed")
	player.room = nil
	//player.Reset()

	data := &RoomClosedMsgData{}
	player.notifyObserver(NewRoomClosedMsg(player, data))
}

func (player *Player) OnEndPlayGame() {
	//log.Debug(time.Now().Unix(), player, "OnPlayingGameEnd")
	player.Reset()
	data := &GameEndMsgData{}
	player.notifyObserver(NewGameEndMsg(player, data))
}
