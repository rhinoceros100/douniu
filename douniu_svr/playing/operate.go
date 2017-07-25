package playing

import "fmt"

type OperateType int

const (
	OperateEnterRoom	OperateType = iota + 1
	OperateReadyRoom
	OperateLeaveRoom

	OperateXiazhu
)

func (operateType OperateType) String() string {
	switch operateType {
	case OperateEnterRoom :
		return "OperateEnterRoom"
	case OperateReadyRoom :
		return "OperateReadyRoom"
	case OperateLeaveRoom:
		return "OperateLeaveRoom"
	case OperateXiazhu:
		return "OperateXiazhu"
	}
	return "unknow OperateType"
}

type Operate struct {//玩家操作
	Op			OperateType
	Operator	*Player				//操作者
	Data		interface{}
	ResultCh		chan bool
}

func (op *Operate) String() string {
	if op == nil {
		return "{operator=nil, op=nil}"
	}
	return fmt.Sprintf("{operator=%v, op=%v}", op.Operator, op.Op)
}

func newOperate(op OperateType, operator *Player, data interface{}) *Operate{
	return &Operate{
		Op:	op,
		Data: data,
		Operator: operator,
		ResultCh: make(chan bool, 1),
	}
}

type OperateEnterRoomData struct {
}
func NewOperateEnterRoom(operator *Player, data *OperateEnterRoomData) *Operate {
	return newOperate(OperateEnterRoom, operator, data)
}

type OperateReadyRoomData struct {
}
func NewOperateReadyRoom(operator *Player, data *OperateReadyRoomData) *Operate {
	return newOperate(OperateReadyRoom, operator, data)
}

type OperateLeaveRoomData struct {
}
func NewOperateLeaveRoom(operator *Player, data *OperateLeaveRoomData) *Operate {
	return newOperate(OperateLeaveRoom, operator, data)
}

type OperateXiazhuData struct {
	Score int32
}
func NewOperateXiazhu(operator *Player, data *OperateXiazhuData) *Operate {
	return newOperate(OperateXiazhu, operator, data)
}
