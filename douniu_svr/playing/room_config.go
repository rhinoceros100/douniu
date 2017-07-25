package playing

type RoomConfig struct {
	NeedPlayerNum			int        `json:"need_player_num"`
	MaxPlayerNum			int        `json:"max_player_num"`
	WaitPlayerEnterRoomTimeout	int        `json:"wait_player_enter_room_timeout"`
	WaitPlayerOperateTimeout	int        `json:"wait_player_operate_timeout"`
	MaxPlayGameCnt			int        `json:"max_play_game_cnt"`	//最大的游戏局数
}

func NewRoomConfig() *RoomConfig {
	return &RoomConfig{}
}

func (config *RoomConfig) Init(need_player_num int) {
	config.NeedPlayerNum = need_player_num
	config.MaxPlayerNum = 6
	config.WaitPlayerEnterRoomTimeout = 300
	config.WaitPlayerOperateTimeout = 300
	config.MaxPlayGameCnt = 5
}