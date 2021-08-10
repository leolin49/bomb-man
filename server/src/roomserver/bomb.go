package main

import (
	"paopao/server/src/common"
	"time"
)

type Bomb struct {
	id    uint32           // 炸弹id，主要用于做Map的key
	pos   *common.Position // 位置
	owner *ScenePlayer     // 所有者
	scene *Scene           // 场景指针
}

func NewBomb(player *ScenePlayer, bombId uint32) *Bomb {
	bomb := &Bomb{
		id:    bombId,
		pos:   &common.Position{X: player.curPos.X, Y: player.curPos.Y},
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
	ticker := time.NewTicker(BOMB_MAXTIME * time.Second)
	<-ticker.C
	this.Explode()
}

// 爆炸
func (this *Bomb) Explode() {
	// 计算伤害范围
	// 1. 上下左右
	up := this.pos.Y + float64(this.owner.power)
	down := this.pos.Y - float64(this.owner.power)
	left := this.pos.X - float64(this.owner.power)
	right := this.pos.X + float64(this.owner.power)
	// 遍历所有炸弹，判断是否在当前炸弹的范围内
	for _, b := range *this.scene.BombMap {
		if b.pos.Y == this.pos.Y && left <= b.pos.X && b.pos.X <= right {
			b.Explode()
		}
		if b.pos.X == this.pos.X && down <= b.pos.Y && b.pos.Y <= up {
			b.Explode()
		}
	}
	// 遍历所有角色，判断是否在当前炸弹的范围内
	for _, p := range this.scene.players {
		if p.curPos.Y == this.pos.Y && left <= p.curPos.X && p.curPos.X <= right {
			this.owner.AddScore(HurtScore)
			p.BeHurt(this.owner)
		}
		if p.curPos.X == this.pos.X && down <= p.curPos.Y && p.curPos.Y <= up {
			this.owner.AddScore(HurtScore)
			p.BeHurt(this.owner)
		}
	}
	// 在bombMap中删除炸弹
	delete(*this.scene.BombMap, this.id)
	//
	this.owner.curbomb--
}
