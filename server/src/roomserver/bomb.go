package main

import (
	"paopao/server/src/common"
	"time"

	"github.com/golang/glog"
)

type Bomb struct {
	id    uint32          // 炸弹id，主要用于做Map的key
	pos   *common.GridPos // 位置
	owner *ScenePlayer    // 所有者
	scene *Scene          // 场景指针
}

func NewBomb(player *ScenePlayer) *Bomb {
	row, col := player.GetCurrentGrid()
	bomb := &Bomb{
		id:    row*player.scene.gameMap.Width + col,
		pos:   &common.GridPos{X: row, Y: col},
		owner: player,
		scene: player.scene,
	}
	// go func() {
	// 	ticker := time.NewTicker(BOMB_MAXTIME * time.Second)
	// 	<-ticker.C

	// 	bomb.Explode()

	// 	return
	// }()
	return bomb
}

// 倒计时
func (this *Bomb) CountDown() {
	// ticker := time.NewTicker(BOMB_MAXTIME * time.Second)
	// <-ticker.C
	time.Sleep(5 * time.Second)
	this.Explode()
}

// 爆炸
func (this *Bomb) Explode() {
	glog.Infof("[%v的炸弹爆炸] x:%v, y:%v",
		this.owner.name, this.pos.X, this.pos.Y)
	// 计算伤害范围
	// 1. 上下左右
	up := this.pos.Y + this.owner.power
	down := this.pos.Y - this.owner.power
	left := this.pos.X - this.owner.power
	right := this.pos.X + this.owner.power
	// 遍历所有炸弹，判断是否在当前炸弹的范围内(一颗炸弹引爆另一颗炸弹)
	for _, b := range this.scene.BombMap {
		if b.pos.Y == this.pos.Y && left <= b.pos.X && b.pos.X <= right {
			// b.Explode()
		}
		if b.pos.X == this.pos.X && down <= b.pos.Y && b.pos.Y <= up {
			// b.Explode()
		}
	}
	// 遍历所有角色，判断是否在当前炸弹的范围内
	for _, p := range this.scene.players {

		x, y := p.GetCurrentGrid()
		glog.Infof("[%v的炸弹爆炸时%v位置] x:%v, y:%v",
			this.owner.name, p.name, x, y)
		if y == this.pos.Y && left <= x && x <= right {
			// 水平方向
			glog.Infof("[炸弹造成<%v,%v>水平方向伤害]", left, right)
			this.owner.AddScore(HurtScore)
			p.BeHurt(this.owner)
		} else if x == this.pos.X && down <= y && y <= up {
			// 垂直方向
			glog.Infof("[炸弹造成<%v,%v>垂直方向伤害]", down, up)
			this.owner.AddScore(HurtScore)
			p.BeHurt(this.owner)
		}
	}
	// 在场景中删除炸弹
	this.scene.DelBomb(this)
	//
	this.owner.curbomb--
}
