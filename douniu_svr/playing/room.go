package playing

import (
	"douniu/douniu_svr/card"
	"time"
	"douniu/douniu_svr/log"
	"fmt"
	"douniu/douniu_svr/util"
)

type RoomStatusType int
const (
	RoomStatusWaitAllPlayerEnter	RoomStatusType = iota	// 等待玩家进入房间
	RoomStatusWaitAllPlayerReady				// 等待玩家准备
	RoomStatusGetMaster					// 确定庄家
	RoomStatusPlayGame					// 正在进行游戏，结束后会进入RoomStatusShowCards
	RoomStatusShowCards					// 亮牌
	RoomStatusEndPlayGame					// 游戏结束后会回到等待游戏开始状态，或者进入结束房间状态
	RoomStatusRoomEnd					// 房间结束状态
)

func (status RoomStatusType) String() string {
	switch status {
	case RoomStatusWaitAllPlayerEnter :
		return "RoomStatusWaitAllPlayerEnter"
	case RoomStatusWaitAllPlayerReady:
		return "RoomStatusWaitAllPlayerReady"
	case RoomStatusGetMaster:
		return "RoomStatusGetMaster"
	case RoomStatusPlayGame:
		return "RoomStatusPlayGame"
	case RoomStatusShowCards:
		return "RoomStatusShowCards"
	case RoomStatusEndPlayGame:
		return "RoomStatusEndPlayGame"
	case RoomStatusRoomEnd:
		return "RoomStatusRoomEnd"
	}
	return "unknow RoomStatus"
}

type RoomObserver interface {
	OnRoomClosed(room *Room)
}

type Room struct {
	id			uint64					//房间id
	config 			*RoomConfig				//房间配置
	players 		[]*Player				//当前房间的玩家列表

	observers		[]RoomObserver				//房间观察者，需要实现OnRoomClose，房间close的时候会通知它
	roomStatus		RoomStatusType				//房间当前的状态
	playedGameCnt		int					//已经玩了的游戏的次数

	//begin playingGameData, reset when start playing game
	cardPool		*card.Pool				//洗牌池
	curOperator		*Player					//获得当前操作的玩家
	masterPlayer 		*Player					//庄
	lastWinPlayer 		*Player					//庄
	//end playingGameData, reset when start playing game

	roomOperateCh		chan *Operate
	roomReadyCh		chan *Operate
	BetCh		[]chan *Operate				//下注
	ShowCardsCh	[]chan *Operate				//亮牌

	stop bool
}

func NewRoom(id uint64, config *RoomConfig) *Room {
	room := &Room{
		id:			id,
		config:			config,
		players:		make([]*Player, 0),
		cardPool:		card.NewPool(),
		observers:		make([]RoomObserver, 0),
		roomStatus:		RoomStatusWaitAllPlayerEnter,
		playedGameCnt:	0,

		roomOperateCh: make(chan *Operate, 1024),
		roomReadyCh: make(chan *Operate, 1024),
		BetCh: make([]chan *Operate, config.MaxPlayerNum),
		ShowCardsCh: make([]chan *Operate, config.MaxPlayerNum),
	}
	for idx := 0; idx < config.MaxPlayerNum; idx ++ {
		room.BetCh[idx] = make(chan *Operate, 1)
		room.ShowCardsCh[idx] = make(chan *Operate, 1)
	}
	return room
}

func (room *Room) GetId() uint64 {
	return room.id
}

func (room *Room) PlayerOperate(op *Operate) {
	pos := op.Operator.position
	log.Debug(time.Now().Unix(), room, op.Operator, "PlayerOperate", op.Op, " pos:", pos)

	switch op.Op {
	case OperateEnterRoom, OperateLeaveRoom:
		room.roomOperateCh <- op
	case OperateReadyRoom:
		room.roomReadyCh <- op
	case OperateBet:
		room.BetCh[pos] <- op
	case OperateShowCards:
		room.ShowCardsCh[pos] <- op
	}
}

func (room *Room) addObserver(observer RoomObserver) {
	room.observers = append(room.observers, observer)
}

func (room *Room) Start() {
	go func() {
		start_time := time.Now().Unix()
		for  {
			if !room.stop {
				room.checkStatus()
				time.Sleep(time.Microsecond * 10)
			}else{
				break
			}
		}
		end_time := time.Now().Unix()
		log.Debug(end_time - start_time, "over^^")
	}()
}

func (room *Room) checkStatus() {
	switch room.roomStatus {
	case RoomStatusWaitAllPlayerEnter:
		room.waitAllPlayerEnter()
	case RoomStatusWaitAllPlayerReady:
		room.waitAllPlayerReady()
	case RoomStatusGetMaster:
		room.getMaster()
	case RoomStatusPlayGame:
		room.playGame()
	case RoomStatusShowCards:
		room.showCards()
	case RoomStatusEndPlayGame:
		room.endPlayGame()
	case RoomStatusRoomEnd:
		room.close()
	}
}

func (room *Room) GetPlayerNum() int {
	return len(room.players)
}

func (room *Room) isRoomEnd() bool {
	return room.playedGameCnt >= room.config.MaxPlayGameCnt
}

func (room *Room) close() {
	log.Debug(time.Now().Unix(), room, "Room.close")
	room.stop = true
	for _, observer := range room.observers {
		observer.OnRoomClosed(room)
	}

	for _, player := range room.players {
		player.OnRoomClosed()
	}
}

func (room *Room) isEnterPlayerEnough() bool {
	length := room.GetPlayerNum()
	log.Debug(time.Now().Unix(), room, "Room.isEnterPlayerEnough, player num :", length, ", need :", room.config.NeedPlayerNum)
	return length >= room.config.NeedPlayerNum
}

func (room *Room) switchStatus(status RoomStatusType) {
	log.Debug(time.Now().Unix(), room, "room status switch,", room.roomStatus, " =>", status)
	room.roomStatus = status
	log.Debug("---------------------------------------")
}

//等待游戏开局
func (room *Room) waitAllPlayerEnter() {
	log_time := time.Now().Unix()
	log.Debug(log_time, room, "waitAllPlayerEnter")
	breakTimerTime := time.Duration(0)
	timeout := time.Duration(room.config.WaitPlayerEnterRoomTimeout) * time.Second
	for {
		timer := timeout - breakTimerTime
		select {
		case <-time.After(timer):
			log.Debug(time.Now().Unix(), room, "waitAllPlayerEnter timeout", timeout)
			room.switchStatus(RoomStatusRoomEnd) //超时发现没有足够的玩家都进入房间了，则结束
			return
		case op := <-room.roomOperateCh:
			if op.Op == OperateEnterRoom || op.Op == OperateLeaveRoom{
				log.Debug(time.Now().Unix(), room, "waitAllPlayerEnter catch operate:", op)
				room.dealPlayerOperate(op)
				go room.waitInitPlayerReady(op.Operator)

				if room.isEnterPlayerEnough() && room.isAllPlayerReady() && room.roomStatus == RoomStatusWaitAllPlayerEnter{
					room.switchStatus(RoomStatusGetMaster)
					go room.waitPlayerJoin()
					return
				}
			}
		case op := <-room.roomReadyCh:
			log.Debug(time.Now().Unix(), room, "waitAllPlayerEnter catch operate:", op)
			room.dealPlayerOperate(op)
			if room.isEnterPlayerEnough() && room.isAllPlayerReady() && room.roomStatus == RoomStatusWaitAllPlayerEnter{
				room.switchStatus(RoomStatusGetMaster)
				go room.waitPlayerJoin()
				return
			}
		}
	}
}

func (room *Room) waitPlayerJoin() {
	log.Debug(time.Now().Unix(), room, "waitPlayerJoin")
	for {
		select {
		case op := <-room.roomOperateCh:
			if op.Op == OperateEnterRoom{
				log.Debug(time.Now().Unix(), room, "waitAllPlayerEnter catch operate:", op)
				room.dealPlayerOperate(op)
			}
		}
		log.Debug(time.Now().Unix(), room, "waitPlayerJoin for")
	}
}

func (room *Room) waitBet(player *Player) bool{
	for{
		select {
		case <- time.After(time.Second * room.config.WaitBetSec):
			data := &OperateBetData{Score:1}
			op := NewOperateBet(player, data)
			log.Debug(player, "waitBet do PlayerOperate")
			room.PlayerOperate(op)
			continue
		case op := <-room.BetCh[player.position]:
			log.Debug(time.Now().Unix(), player, "Player.waitBet:", op.Data)
			room.dealPlayerOperate(op)
			return true
		}
	}

	log.Debug(time.Now().Unix(), player, "Player.waitBet fasle")
	return false
}

func (room *Room) waitShowCards(player *Player) bool{
	for{
		select {
		case <- time.After(time.Second * room.config.WaitShowCardsSec):
			data := &OperateShowCardsData{}
			op := NewOperateShowCards(player, data)
			log.Debug(player, "waitShowCards do PlayerOperate")
			room.PlayerOperate(op)
			continue
		case op := <-room.ShowCardsCh[player.position]:
			log.Debug(time.Now().Unix(), player, "Player.waitShowCards:", op.Data)
			room.dealPlayerOperate(op)
			return true
		}
	}

	log.Debug(time.Now().Unix(), player, "Player.waitShowCards fasle")
	return false
}

func (room *Room) waitInitPlayerReady(player *Player) {
	time.Sleep(time.Second * room.config.WaitReadySec)
	if room.roomStatus == RoomStatusWaitAllPlayerEnter && !player.GetIsReady() {
		data := &OperateReadyRoomData{}
		op := NewOperateReadyRoom(player, data)
		log.Debug(player, "waitInitPlayerReady do PlayerOperate")
		room.PlayerOperate(op)
	}
}

func (room *Room) waitPlayerReady(player *Player) bool {
	for{
		select {
		case <- time.After(time.Second * room.config.WaitReadySec):
			data := &OperateReadyRoomData{}
			op := NewOperateReadyRoom(player, data)
			log.Debug(player, "waitPlayerReady do PlayerOperate")
			room.PlayerOperate(op)
			continue
		case op := <-room.roomReadyCh:
			log.Debug(time.Now().Unix(), player, "Player.waitPlayerReady")
			room.dealPlayerOperate(op)
			return true
		}
	}

	log.Debug(time.Now().Unix(), player, "Player.waitPlayerReady fasle")
	return false
}

func (room *Room) waitAllPlayerReady() {
	log.Debug(time.Now().Unix(), room, "Room.waitAllPlayerReady")
	/*if room.playedGameCnt == 0 {
		room.roomStatus = RoomStatusStartPlayGame
		log.Debug(time.Now().Unix(), room, "RoomStatusStartPlayGame")
		return
	}*/
	//等待所有玩家准备
	for _, player := range room.players {
		go room.waitPlayerReady(player)
	}
	for  {
		if room.isAllPlayerReady() {
			room.roomStatus = RoomStatusGetMaster
			log.Debug(time.Now().Unix(), room, "RoomStatusStartPlayGame")
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (room *Room) getMaster() {
	log.Debug(time.Now().Unix(), room, "Room.getMaster", room.playedGameCnt)

	// 重置牌池, 洗牌
	room.cardPool.ReGenerate()

	// 随机一个玩家首先开始
	room.masterPlayer = room.selectMasterPlayer()
	room.curOperator = room.masterPlayer

	//通知所有玩家确定了庄家
	for _, player := range room.players {
		player.OnGetMaster()
	}

	//等待所有玩家下注，庄家除外
	for _, player := range room.players {
		if !player.IsMaster(){
			go room.waitBet(player)
		}
	}

	room.switchStatus(RoomStatusPlayGame)
}

func (room *Room) playGame() {
	log.Debug(time.Now().Unix(), room, "Room.playGame", room.playedGameCnt)

	cnt := 0
	for !room.isAllPlayerBet() {
		time.Sleep(time.Millisecond * 100)
		//log.Debug(time.Now().Unix(), "cnt:", cnt)
		cnt ++
	}

	//通知所有玩家下注金额
	for _, player := range room.players {
		player.OnAllBet()
	}

	time.Sleep(time.Second * room.config.AfterBetSleep)

	//发初始牌给所有玩家
	room.putInitCardsToPlayers()

	//通知所有玩家手上的牌
	for _, player := range room.players {
		player.OnGetInitCards()
	}

	//等待所有玩家亮牌
	for _, player := range room.players {
		go room.waitShowCards(player)
	}

	room.switchStatus(RoomStatusShowCards)
}

func (room *Room) showCards() {
	log.Debug(time.Now().Unix(), room, "Room.showCards", room.playedGameCnt)

	cnt := 0
	for !room.isAllPlayerShow() {
		time.Sleep(time.Millisecond * 500)
		log.Debug(time.Now().Unix(), "cnt:", cnt)
		cnt ++
	}

	time.Sleep(time.Second)

	//结算
	msg := room.jiesuan()

	//通知所有玩家结算结果
	for _, player := range room.players {
		player.OnJiesuan(msg)
	}

	time.Sleep(time.Second * room.config.AfterShowCardsSleep)
	room.switchStatus(RoomStatusEndPlayGame)
}

func (room *Room) endPlayGame() {
	room.playedGameCnt++
	log.Debug(time.Now().Unix(), room, "Room.endPlayGame cnt :", room.playedGameCnt)
	if room.isRoomEnd() {
		log.Debug(time.Now().Unix(), room, "Room.endPlayGame room end")
		room.switchStatus(RoomStatusRoomEnd)
	} else {
		for _, player := range room.players {
			player.OnEndPlayGame()
		}
		log.Debug(time.Now().Unix(), room, "Room.endPlayGame restart play game")
		room.switchStatus(RoomStatusWaitAllPlayerReady)
	}
}

func (room *Room) jiesuan() *Message {
	master_player := room.masterPlayer
	master_paixing := master_player.GetPaixing()
	master_paixing_multiple := master_player.GetPaixingMultiple()
	master_base_multiple := master_player.GetBaseMultiple()
	master_maxid := master_player.GetMaxid()
	master_jiesuan_data := &PlayerJiesuanData{
		P:master_player,
		Score:0,
		Paixing:master_paixing,
		PaixingMultiple:master_paixing_multiple,
		BaseMultiple:int(master_base_multiple),
	}
	max_paixing := master_paixing
	max_id := master_maxid
	master_player.SetLeizhu(0)

	data := &JiesuanMsgData{}
	data.Scores = make([]*PlayerJiesuanData, 0)
	for _, player := range room.players {
		if player != master_player {
			player_paixing := player.GetPaixing()
			player_maxid := player.GetMaxid()
			player_paixing_multiple := player.GetPaixingMultiple()
			player_base_multiple := player.GetBaseMultiple()
			player_bet_score := player.GetBetScore()
			round_score := player_bet_score * player_base_multiple * int32(player_paixing_multiple)
			player_jiesuan_data := &PlayerJiesuanData{
				P:player,
				Score:0,
				Paixing:player_paixing,
				PaixingMultiple:player_paixing_multiple,
				BaseMultiple:int(player_base_multiple),
			}

			if player_paixing > master_paixing || (player_paixing == master_paixing && player_maxid > master_maxid){
				player.SetRoundScore(round_score)
				player.AddTotalScore(round_score)
				player_jiesuan_data.Score = round_score

				master_player.SetRoundScore(-round_score)
				master_player.AddTotalScore(-round_score)
				master_jiesuan_data.Score -= round_score

				//垒注
				if player.GetLeizhu() != 0 {
					player.SetLeizhu(0)
				}else{
					if round_score + player_bet_score > room.GetScoreHigh() {
						player.SetLeizhu(round_score + player_bet_score)
					}else {
						player.SetLeizhu(0)
					}
				}

				//牛牛上庄需记录上庄谁得了牛牛
				if player_paixing >= card.DouniuType_Niuniu{
					if player_paixing > max_paixing || (player_paixing == max_paixing && player_maxid > max_id){
						max_paixing = player_paixing
						max_id = player_maxid
						room.lastWinPlayer = player
					}
				}

			}else{
				player.SetRoundScore(-round_score)
				player.AddTotalScore(-round_score)
				player_jiesuan_data.Score = -round_score

				master_player.SetRoundScore(round_score)
				master_player.AddTotalScore(round_score)
				master_jiesuan_data.Score += round_score

				player.SetLeizhu(0)
			}
			data.Scores = append(data.Scores, player_jiesuan_data)
		}
	}
	data.Scores = append(data.Scores, master_jiesuan_data)
	return NewJiesuanMsg(nil, data)
}

//取指定玩家的下一个玩家
func (room *Room) nextPlayer(player *Player) *Player {
	pos := player.GetPosition()

	max_player_num := int32(room.config.MaxPlayerNum)
	for next_pos := pos + 1; next_pos < max_player_num; next_pos++ {
		for _, room_player := range room.players {
			if room_player.GetPosition() == next_pos {
				log.Debug(time.Now().Unix(), ", nextPlayer", ", pos:", pos, ", next_pos:", next_pos)
				return room_player
			}
		}
	}

	for next_pos := int32(0); next_pos < pos; next_pos++ {
		for _, room_player := range room.players {
			if room_player.GetPosition() == next_pos {
				log.Debug(time.Now().Unix(), ", nextPlayer", ", pos:", pos, ", next_pos:", next_pos)
				return room_player
			}
		}
	}

	log.Debug(time.Now().Unix(), ", nextPlayer", ", pos:", pos, ", next_pos:", 0)
	return room.players[0]
}

func (room *Room) isAllPlayerReady() bool{
	for _, player := range room.players {
		if !player.isReady {
			return false
		}
	}
	return true
}

func (room *Room) isAllPlayerBet() bool{
	for _, player := range room.players {
		if player.GetIsPlaying() && !player.IsMaster() && !player.GetIsBet() {
			return false
		}
	}
	return true
}

func (room *Room) isAllPlayerShow() bool{
	for _, player := range room.players {
		if player.GetIsPlaying() && !player.GetIsShowCards() {
			return false
		}
	}
	return true
}

//处理玩家操作
func (room *Room) dealPlayerOperate(op *Operate) bool{
	log_time := time.Now().Unix()
	log.Debug(log_time, room, "Room.dealPlayerOperate :", op)
	switch op.Op {
	case OperateEnterRoom:
		if _, ok := op.Data.(*OperateEnterRoomData); ok {
			if room.addPlayer(op.Operator) {
				//玩家进入成功
				player_pos := room.getMinUsablePosition()
				op.Operator.EnterRoom(room, player_pos)
				log.Debug(log_time, room, "Room.dealPlayerOperate player enter :", op.Operator)
				op.ResultCh <- true
				room.broadcastPlayerSuccessOperated(op)
				return true
			}
		}

	case OperateReadyRoom:
		if _, ok := op.Data.(*OperateReadyRoomData); ok {
			if room.readyPlayer(op.Operator) { //	玩家确认开始游戏
				op.Operator.ReadyRoom(room)
				log.Debug(log_time, room, "Room.dealPlayerOperate player ready :", op.Operator)
				op.ResultCh <- true
				room.broadcastPlayerSuccessOperated(op)
				return true
			}
		}

	case OperateLeaveRoom:
		if _, ok := op.Data.(*OperateLeaveRoomData); ok {
			log.Debug(log_time, room, "Room.dealPlayerOperate player leave :", op.Operator)
			room.delPlayer(op.Operator)
			op.Operator.LeaveRoom()
			op.ResultCh <- true
			room.broadcastPlayerSuccessOperated(op)
			return true
		}

	case OperateBet:
		if bet_data, ok := op.Data.(*OperateBetData); ok {
			log.Debug(log_time, room, "Room.dealPlayerOperate player bet :", op.Operator)
			op.Operator.Bet(bet_data.Score)
			op.ResultCh <- true
			room.broadcastPlayerSuccessOperated(op)
			return true
		}

	case OperateShowCards:
		if show_data, ok := op.Data.(*OperateShowCardsData); ok {
			log.Debug(log_time, room, "Room.dealPlayerOperate player show cards :", op.Operator)
			op.Operator.ShowCards()
			op.ResultCh <- true
			show_data.Paixing = op.Operator.GetPaixing()
			show_data.PaixingMultiple = op.Operator.GetPaixingMultiple()
			room.broadcastPlayerSuccessOperated(op)
			return true
		}

	case OperateSeeCards:
		if _, ok := op.Data.(*OperateSeeCardsData); ok {
			log.Debug(log_time, room, "Room.dealPlayerOperate player see cards :", op.Operator)
			op.Operator.SeeCards()
			op.ResultCh <- true
			room.broadcastPlayerSuccessOperated(op)
			return true
		}

	}
	op.ResultCh <- false
	return false
}

//查找房间中未被占用的最新的position
func (room *Room) getMinUsablePosition() (int32)  {
	log.Debug(time.Now().Unix(), room, "getMinUsablePosition")
	//获取所有已经被占用的position
	player_positions := make([]int32, 0)
	for _, room_player := range room.players {
		player_positions = append(player_positions, room_player.GetPosition())
	}

	//查找未被占用的position中最小的
	room_max_position := int32(room.config.MaxPlayerNum - 1)
	for i := int32(0); i <= room_max_position; i++ {
		is_occupied := false
		for _, occupied_pos := range player_positions{
			if occupied_pos == i {
				is_occupied = true
				break
			}
		}
		if !is_occupied {
			return i
		}
	}
	return room_max_position
}

//给所有玩家发初始化的5张牌
func (room *Room) putInitCardsToPlayers() {
	log.Debug(time.Now().Unix(), room, "Room.initAllPlayer")
	/*for _, player := range room.players {
		player.Reset()
	}*/

	for num := 0; num < card.NIUNIU_INIT_CARD_NUM; num++ {
		for _, player := range room.players {
			room.putCardToPlayer(player)
		}
	}
}

//添加玩家
func (room *Room) addPlayer(player *Player) bool {
	/*if room.roomStatus != RoomStatusWaitAllPlayerEnter {
		return false
	}*/
	if room.GetPlayerNum() >= room.config.MaxPlayerNum {
		return false
	}
	room.players = append(room.players, player)
	return true
}

func (room *Room) readyPlayer(player *Player) bool {
	if room.roomStatus != RoomStatusWaitAllPlayerEnter && room.roomStatus != RoomStatusWaitAllPlayerReady{
		return false
	}
	player.SetIsReady(true)
	return true
}

func (room *Room) delPlayer(player *Player)  {
	for idx, p := range room.players {
		if p == player {
			room.players = append(room.players[0:idx], room.players[idx+1:]...)
			return
		}
	}
}

func (room *Room) broadcastPlayerSuccessOperated(op *Operate) {
	log.Debug(time.Now().Unix(), room, "Room.broadcastPlayerSucOp :", op)
	for _, player := range room.players {
		player.OnPlayerSuccessOperated(op)
	}
}

//发牌给指定玩家
func (room *Room) putCardToPlayer(player *Player) *card.Card {
	card := room.cardPool.PopFront()
	if card == nil {
		return nil
	}
	player.AddCard(card)
	return card
}

func (room *Room) randomPlayer() *Player {
	idx := util.RandomN(len(room.players))
	log.Debug(time.Now().Unix(), room, "Room.randomPlayer", room.players[idx])
	return room.players[idx]
}

//选择庄家
func (room *Room) selectMasterPlayer() *Player {
	log.Debug(time.Now().Unix(), room, "Room.selectMasterPlayer")
	if room.playedGameCnt == 0 { //第一盘，随机一个做东
		return room.randomPlayer()
	}

	if room.config.PlayGameType == GameTypeNiuniu {
		if room.lastWinPlayer == nil {//流局，上一盘没有人胡牌
			return room.randomPlayer()
		}

		return room.lastWinPlayer
	}else if room.config.PlayGameType == GameTypeLunliu {
		return room.nextPlayer(room.masterPlayer)
	}
	return room.lastWinPlayer
}

func (room *Room) String() string {
	if room == nil {
		return "{room=nil}"
	}
	return fmt.Sprintf("{room=%v}", room.GetId())
}

func (room *Room) GetScoreLow() int32 {
	return int32(room.config.ScoreLow)
}

func (room *Room) GetScoreHigh() int32 {
	return int32(room.config.ScoreHigh)
}

func (room *Room) clearChannel() {
	for idx := 0 ; idx < room.config.MaxPlayerNum; idx ++ {
		select {
		case op := <-room.BetCh[idx]:
			op.ResultCh <- false
		default:
		}

		select {
		case op := <-room.ShowCardsCh[idx]:
			op.ResultCh <- false
		default:
		}
	}
}
