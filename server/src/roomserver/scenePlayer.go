package main

import (
	"math/rand"
	"paopao/server/src/common"
	"time"
)

// 初始数值
const (
	INIT_POWER = 1 // 初始炸弹威力
	INIT_SPEED = 3 // 初始移动速度
)

type ScenePlayer struct {
	id      uint64          // 玩家id
	name    string          // 玩家昵称
	key     string          // 登录token
	score   float64         // 分数
	pos     common.Position // 当前位置
	hp      uint32          // 生命值
	curbomb uint32          // 当前已经放置的炸弹数量
	maxbomb uint32          // 能放置的最大炸弹数量
	power   uint32          // 炸弹威力
	speed   uint32          // 移动速度

	playerTask *PlayerTask // 用于通信
}

func NewScenePlayer(player *PlayerTask, scene *Scene) *ScenePlayer {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	_ = r
	s := &ScenePlayer{
		id:      player.id,
		name:    player.name,
		key:     player.key,
		score:   0,
		pos:     common.Position{X: 0, Y: 0},
		hp:      3,
		curbomb: 0,
		maxbomb: 1,
		power:   INIT_POWER,
		speed:   INIT_SPEED,

		playerTask: player,
	}
	return s
}

// ----------------------玩家角色事件-------------------------- //

// 放置炸弹
func (this *ScenePlayer) PutBomb() {

}

// 移动
func (this *ScenePlayer) Move() {

}
