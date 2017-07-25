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
	RoomStatusConfirmMaster					// 确定庄家
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
	case RoomStatusConfirmMaster:
		return "RoomStatusConfirmMaster"
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
	xiazhuCh		[]chan *Operate				//解散房间

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
		xiazhuCh: make([]chan *Operate, config.MaxPlayerNum),
	}
	for idx := 0; idx < config.MaxPlayerNum; idx ++ {
		room.xiazhuCh[idx] = make(chan *Operate, 1)
	}
	return room
}

func (room *Room) GetId() uint64 {
	return room.id
}

func (room *Room) PlayerOperate(op *Operate) {
	//idx := op.Operator.position
	switch op.Op {
	case OperateEnterRoom, OperateLeaveRoom, OperateReadyRoom:
		room.roomOperateCh <- op
	}
}

func (room *Room) addObserver(observer RoomObserver) {
	room.observers = append(room.observers, observer)
}

func (room *Room) Start() {
	go func() {
		for  {
			if !room.stop {
				room.checkStatus()
				time.Sleep(time.Microsecond * 10)
			}else{
				break
			}
		}
	}()
}

func (room *Room) checkStatus() {
	switch room.roomStatus {
	case RoomStatusWaitAllPlayerEnter:
		room.waitAllPlayerEnter()
	case RoomStatusWaitAllPlayerReady:
		room.waitAllPlayerReady()
	case RoomStatusConfirmMaster:
		room.confirmMaster()
	case RoomStatusPlayGame:
		room.playGame()
	case RoomStatusShowCards:
		room.showCards()
	/*case RoomStatusEndPlayGame:
		room.endPlayGame()
	case RoomStatusRoomEnd:
		room.close()*/
	}
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

	/*for _, player := range room.players {
		player.OnRoomClosed()
	}*/
}

func (room *Room) isEnterPlayerEnough() bool {
	length := len(room.players)
	log.Debug(time.Now().Unix(), room, "Room.isEnterPlayerEnough, player num :", length, ", need :", room.config.NeedPlayerNum)
	return length >= room.config.NeedPlayerNum
}

func (room *Room) switchStatus(status RoomStatusType) {
	log.Debug(time.Now().Unix(), room, "room status switch,", room.roomStatus, " =>", status)
	room.roomStatus = status
}

//等待玩家进入
func (room *Room) waitAllPlayerEnter() {
	log_time := time.Now().Unix()
	log.Debug(log_time, room, "Room.waitAllPlayerEnter")
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
			if op.Op == OperateEnterRoom || op.Op == OperateLeaveRoom || op.Op == OperateReadyRoom{
				log.Debug(time.Now().Unix(), room, "Room.waitAllPlayerEnter catch operate:", op)
				room.dealPlayerOperate(op)
				if room.isEnterPlayerEnough() && room.isAllPlayerReady(){
					room.switchStatus(RoomStatusConfirmMaster)
					return
				}
			}
		}
	}
}

func (room *Room) waitXiazhu(player *Player) bool{
	select {
	case <- time.After(time.Second * 10):
		data := &OperateXiazhuData{Score:1}
		op := NewOperateXiazhu(player, data)
		room.PlayerOperate(op)
	case op := <-room.xiazhuCh[player.position]:
		log.Debug(time.Now().Unix(), player, "Player.waitXiazhu:", op.Data)
		if xiazhu_data, ok := op.Data.(*OperateXiazhuData); ok {
			player.xiazhuScore = xiazhu_data.Score
			return true
		}
		return false
	}
	log.Debug(time.Now().Unix(), player, "Player.waitXiazhu fasle")
	return false
}

//TODO
func (room *Room) waitAllPlayerReady() {
	log.Debug(time.Now().Unix(), room, "Room.waitAllPlayerReady")
	/*if room.playedGameCnt == 0 {
		room.roomStatus = RoomStatusStartPlayGame
		log.Debug(time.Now().Unix(), room, "RoomStatusStartPlayGame")
		return
	}*/
	for  {
		if room.isAllPlayerReady() {
			room.roomStatus = RoomStatusConfirmMaster
			log.Debug(time.Now().Unix(), room, "RoomStatusStartPlayGame")
			return
		}
		time.Sleep(time.Second)
	}
}

func (room *Room) confirmMaster() {
	log.Debug(time.Now().Unix(), room, "Room.confirmMaster")

	// 重置牌池, 洗牌
	room.cardPool.ReGenerate()

	// 随机一个玩家首先开始
	room.masterPlayer = room.selectMasterPlayer()
	room.curOperator = room.masterPlayer

	//通知所有玩家确定了庄家
	for _, player := range room.players {
		player.OnConfirmMaster()
	}

	room.switchStatus(RoomStatusPlayGame)

	//等待所有玩家下注
	for _, player := range room.players {
		go room.waitXiazhu(player)
	}
}

func (room *Room) playGame() {
	log.Debug(time.Now().Unix(), room, "Room.playGame")

	//发初始牌给所有玩家
	room.putInitCardsToPlayers()

	//通知所有玩家手上的牌
	for _, player := range room.players {
		player.OnGetInitCards()
	}

	room.switchStatus(RoomStatusShowCards)
}

//TODO
func (room *Room) showCards() {

}

func (room *Room) isAllPlayerReady() bool{
	for _, player := range room.players {
		if !player.isReady {
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
				//	玩家进入成功
				op.Operator.EnterRoom(room, len(room.players) - 1)
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

	}
	op.ResultCh <- false
	return false
}

//给所有玩家发初始化的5张牌
func (room *Room) putInitCardsToPlayers() {
	log.Debug(time.Now().Unix(), room, "Room.initAllPlayer")
	for _, player := range room.players {
		player.Reset()
	}

	for num := 0; num < card.NIUNIU_INIT_CARD_NUM; num++ {
		for _, player := range room.players {
			room.putCardToPlayer(player)
		}
	}
}

//添加玩家
func (room *Room) addPlayer(player *Player) bool {
	if room.roomStatus != RoomStatusWaitAllPlayerEnter {
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
	log.Debug(time.Now().Unix(), room, "Room.broadcastPlayerSuccessOperated :", op)
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

	if room.lastWinPlayer == nil {//流局，上一盘没有人胡牌
		return room.randomPlayer()
	}

	return room.lastWinPlayer
}

func (room *Room) String() string {
	if room == nil {
		return "{room=nil}"
	}
	return fmt.Sprintf("{room=%v}", room.GetId())
}
