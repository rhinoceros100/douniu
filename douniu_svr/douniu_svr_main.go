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
	log.Debug("-----------------help---------------------")
}

type PlayerObserver struct {}
func (ob *PlayerObserver) OnMsg(player *playing.Player, msg *playing.Message) {
	log_time := time.Now().Unix()
	log.Debug(log_time, player, "receive msg", msg)
	log.Debug(log_time, player, "playingcards :", player.GetPlayingCards())
}

func main() {
	running := true

	//init room
	conf := playing.NewRoomConfig()
	conf.Init(3)
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

	go func() {
		time.Sleep(time.Second * 1)
		robots[0].OperateDoReady()
		time.Sleep(time.Second * 2)
		robots[1].OperateDoReady()
		time.Sleep(time.Second * 8)
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
		}
	}
}
