package main

import (
	"paopao/server/src/common"
	"runtime/debug"
	"sync"
	"time"

	"github.com/golang/glog"
)

type PlayerTaskManager struct {
	mutex sync.Mutex
	tasks map[uint64]*PlayerTask
}

var mPlayerTaskMgr *PlayerTaskManager

func PlayerTaskManager_GetMe() *PlayerTaskManager {
	if mPlayerTaskMgr == nil {
		mPlayerTaskMgr = &PlayerTaskManager{
			tasks: make(map[uint64]*PlayerTask),
		}
		// go mPlayerTaskMgr.iTimeAction()
	}
	return mPlayerTaskMgr
}

func (this *PlayerTaskManager) Add(task *PlayerTask) bool {
	if task == nil {
		return false
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.tasks[task.id] = task
	return true
}

func (this *PlayerTaskManager) Remove(task *PlayerTask) bool {
	if task == nil {
		return false
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	t, ok := this.tasks[task.id]
	if !ok {
		return false
	}
	if t != task {
		glog.Errorln("[PlayerTaskManager Remove] error ")
		return false
	}

	delete(this.tasks, task.id)
	return true
}

func (this *PlayerTaskManager) GetTask(uid uint64) *PlayerTask {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	user, ok := this.tasks[uid]
	if !ok {
		return nil
	}
	return user
}

func (this *PlayerTaskManager) GetNum() int32 {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return int32(len(this.tasks))
}

// 玩家超时无操作
func (this *PlayerTaskManager) iTimeAction() {
	var (
		timeTicker    = time.NewTicker(time.Second)
		loop          uint64
		noActionTasks []*PlayerTask // 无操作玩家列表
	)
	defer func() {
		timeTicker.Stop()
		if err := recover(); err != nil {
			glog.Errorln("[异常] 定时线程错误 ", err, "\n", string(debug.Stack()))
		}
	}()

	for {
		select {
		case <-timeTicker.C:
			if 0 == loop%5 {
				now := time.Now()

				this.mutex.Lock()
				for _, task := range this.tasks {
					if now.Sub(task.activeTime) > common.PlayerTaskTimeOut*time.Second {
						noActionTasks = append(noActionTasks, task)
					}
				}
				this.mutex.Unlock()
				// 删除玩家
				for _, task := range noActionTasks {
					if !task.tcptask.IsClosed() {
						this.Remove(task)
					}
					glog.Infof("[iTimeAction] player %v connect timeout", task.id)
				}
				noActionTasks = noActionTasks[:0]
			}
			loop += 1
		}
	}
}
