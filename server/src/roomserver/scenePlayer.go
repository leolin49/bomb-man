package main

import (
	"paopao/server/src/common"
)

type ScenePlayer struct {
	id      uint64          // 玩家id
	sid     uint32          // 场景id
	name    string          // 玩家昵称
	key     string          // 登录token
	score   float64         // 分数
	pos     common.Position // 当前位置
	hp      uint32          //生命值
	curbomb uint32          // 当前已经放置的炸弹数量
	maxbomb uint32          // 能放置的最大炸弹数量
	power   uint32          // 炸弹威力
	speed   uint32          // 移动速度
}
