package main

import (
	"sync"

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
	t.scenePlayer = nil
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
