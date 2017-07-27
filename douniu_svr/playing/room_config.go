package playing

import (
	"time"
)

type GameType int
const (
	GameTypeNiuniu	GameType = iota			//牛牛上庄
	GameTypeLunliu					//轮流上庄
	GameTypeMingpai					//明牌上庄
)

type RoomConfig struct {
	PlayGameType			GameType        `json:"play_game_type"`

	NeedPlayerNum			int        `json:"need_player_num"`
	MaxPlayerNum			int        `json:"max_player_num"`
	WaitPlayerEnterRoomTimeout	int        `json:"wait_player_enter_room_timeout"`
	WaitPlayerOperateTimeout	int        `json:"wait_player_operate_timeout"`
	MaxPlayGameCnt			int        `json:"max_play_game_cnt"`	//最大的游戏局数
	ScoreLow			int        `json:"score_low"`	//下注底分
	ScoreHigh			int        `json:"score_high"`	//下注高分

	WaitBetSec                 	time.Duration      `json:"wait_bet_sec"`	//等待下注时长
	WaitShowCardsSec              	time.Duration      `json:"wait_show_cards_sec"`	//等待亮牌时长
	WaitReadySec              	time.Duration      `json:"wait_ready_sec"`	//等待准备时长

	AfterBetSleep                   time.Duration      `json:"after_bet_sleep"`	//下注后sleep时长
	AfterShowCardsSleep             time.Duration      `json:"after_show_cards_sleep"`	//亮牌后sleep时长
}

func NewRoomConfig() *RoomConfig {
	return &RoomConfig{}
}

func (config *RoomConfig) Init(score_low, score_high int, play_game_type GameType) {
	config.PlayGameType = play_game_type

	config.ScoreLow = score_low
	config.ScoreHigh = score_high
	config.NeedPlayerNum = 3
	config.MaxPlayerNum = 6
	config.WaitPlayerEnterRoomTimeout = 300
	config.WaitPlayerOperateTimeout = 300
	config.MaxPlayGameCnt = 3

	config.WaitBetSec = 15
	config.WaitShowCardsSec = 15
	config.WaitReadySec = 15

	config.AfterBetSleep = 4
	config.AfterShowCardsSleep = 5
}