package main

import (
	"bufio"
	"os"
	"strings"
	"strconv"
	"time"
	"douniu/douniu_svr/playing"
	"douniu/douniu_svr/util"
	"douniu/douniu_svr/log"
)

func help() {
	log.Debug("-----------------help---------------------")
	log.Debug("h")
	log.Debug("exit")
	log.Debug("mycards")
	log.Debug(playing.OperateEnterRoom, int(playing.OperateEnterRoom))
	log.Debug(playing.OperateReadyRoom, int(playing.OperateReadyRoom))
	log.Debug(playing.OperateLeaveRoom, int(playing.OperateLeaveRoom))
	log.Debug(playing.OperateScramble, int(playing.OperateScramble), "4(multiple)")
	log.Debug(playing.OperateBet, int(playing.OperateBet), "1(score)")
	log.Debug(playing.OperateShowCards, int(playing.OperateShowCards))
	log.Debug(playing.OperateSeeCards, int(playing.OperateSeeCards))
	log.Debug("-----------------help---------------------")
}

type PlayerObserver struct {}
func (ob *PlayerObserver) OnMsg(player *playing.Player, msg *playing.Message) {
	log_time := time.Now().Unix()
	switch msg.Type {
	case playing.MsgEnterRoom:
		if enter_data, ok := msg.Data.(*playing.EnterRoomMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgEnterRoom, EnterPlayer:", enter_data.EnterPlayer)
		}
	case playing.MsgReadyRoom:
		if enter_data, ok := msg.Data.(*playing.ReadyRoomMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgReadyRoom, ReadyPlayer:", enter_data.ReadyPlayer)
		}
	case playing.MsgLeaveRoom:
		if enter_data, ok := msg.Data.(*playing.LeaveRoomMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgLeaveRoom, LeavePlayer:", enter_data.LeavePlayer)
		}
	case playing.MsgGameEnd:
		if _, ok := msg.Data.(*playing.GameEndMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgGameEnd")
		}
	case playing.MsgRoomClosed:
		if _, ok := msg.Data.(*playing.RoomClosedMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgRoomClosed")
		}

	case playing.MsgGetInitCards:
		if init_data, ok := msg.Data.(*playing.GetInitCardsMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgGetInitCards, PlayingCards:", init_data.PlayingCards)
		}
	case playing.MsgGetMaster:
		if init_data, ok := msg.Data.(*playing.GetMasterMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgGetMaster, Scores:", init_data.Scores)
		}
	case playing.MsgBet:
		if _, ok := msg.Data.(*playing.BetMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgBet")
		}
	case playing.MsgDispatchCard:
		if diapatched_data, ok := msg.Data.(*playing.DispatchCardMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgDispatchCard, DispatchedCard:", diapatched_data.DispatchedCard)
		}
	case playing.MsgSeeCards:
		if see_data, ok := msg.Data.(*playing.SeeCardsMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgSeeCards, SeePlayer:", see_data.SeePlayer)
		}
	case playing.MsgShowCards:
		if show_data, ok := msg.Data.(*playing.ShowCardsMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgShowCards, ShowPlayer:", show_data.ShowPlayer, show_data.Paixing, show_data.PaixingMultiple)
		}
	case playing.MsgJiesuan:
		if jiesuan_data, ok := msg.Data.(*playing.JiesuanMsgData); ok {
			log.Debug(log_time, player, "OnMsg MsgJiesuan, jiesuan_data:")
			for _, score_data := range jiesuan_data.Scores	{
				log.Debug(score_data.P, score_data.P.IsMaster(), score_data.Score, score_data.Paixing, score_data.PaixingMultiple, score_data.BaseMultiple)
			}
		}
	}
}

func main() {
	running := true

	//init room
	conf := playing.NewRoomConfig()
	conf.Init(1, 2, playing.GameTypeLunliu)
	room := playing.NewRoom(util.UniqueId(), conf)
	room.Start()

	robots := []*playing.Player{
		playing.NewPlayer(1),
		playing.NewPlayer(2),
		playing.NewPlayer(3),
	}

	for _, robot := range robots {
		robot.OperateEnterRoom(room)
		robot.AddObserver(&PlayerObserver{})
	}

	curPlayer := playing.NewPlayer(4)
	curPlayer.AddObserver(&PlayerObserver{})
	curPlayer2 := playing.NewPlayer(5)
	curPlayer2.AddObserver(&PlayerObserver{})

	go func() {
		time.Sleep(time.Second * 1)
		robots[0].OperateDoReady()
		time.Sleep(time.Second * 2)
		robots[1].OperateDoReady()
		time.Sleep(time.Second * 5)
		robots[2].OperateDoReady()
		//curPlayer.OperateDoReady()
	}()

	reader := bufio.NewReader(os.Stdin)
	for running {
		data, _, _ := reader.ReadLine()
		cmd := string(data)
		if cmd == "h" {
			help()
		} else if cmd == "exit" {
			return
		} else if cmd == "mycards" {
			log.Debug(curPlayer.GetPlayingCards())
		}
		splits := strings.Split(cmd, " ")
		c, _ := strconv.Atoi(splits[0])
		switch playing.OperateType(c) {
		case playing.OperateEnterRoom:
			curPlayer.OperateEnterRoom(room)
		case playing.OperateReadyRoom:
			curPlayer.OperateDoReady()
		case playing.OperateLeaveRoom:
			curPlayer.OperateLeaveRoom()
		case playing.OperateBet:
			score, _ := strconv.Atoi(splits[1])
			curPlayer.OperateBet(int32(score))
		case playing.OperateShowCards:
			curPlayer.OperateShowCards()
		case playing.OperateSeeCards:
			curPlayer.OperateSeeCards()
		}
	}
}
