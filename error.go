package tgf

import "errors"

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ群 7400585
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/3/16
//***************************************************

type GameError interface {
	Error() string
	Code() int32
}

var (
	ErrorRPCTimeOut = errors.New("rpc time out")
)
