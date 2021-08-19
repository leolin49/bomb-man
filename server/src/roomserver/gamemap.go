package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const (
	MaxGridNumberX = 32
	MaxGridNumberY = 32
)

type GridType uint32

const (
	GridType_Space    = GridType(iota) // 空地
	GridType_Obstacle                  // 不可摧毁的障碍物
	GridType_Box                       // 箱子（可摧毁）
	GridType_Bomb                      // 炸弹
)

type GameMap struct {
	Width    uint32
	Height   uint32
	MapArray [][]GridType
}

const (
	BGFilePath         = "./gamemap/BG.json"         // 空地，但是空地上有可能有障碍物
	BoundFilePath      = "./gamemap/Bound.json"      // 地图边界（暂时用不到）
	ForeGroundFilePath = "./gamemap/ForeGround.json" // 障碍物
)

const (
	OFFSET_X = 16
	OFFSET_Y = 8
)

func ReadFileAll(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (this *GameMap) GetGridByPos(x, y uint32) GridType {
	return this.MapArray[x][y]
}

// 自定义地图信息
// 自定义地图信息
func (this *GameMap) CustomizeInit() bool {
	glog.Infoln("[游戏地图初始化] 初始化开始")
	this.Width, this.Height = 40, 40
	this.MapArray = make([][]GridType, this.Height)
	for i := 0; i < int(this.Height); i++ {
		this.MapArray[i] = make([]GridType, this.Width)
	}

	splitPos := func(posStr string) (int, int) {
		arr := strings.Split(posStr, ",")
		x, _ := strconv.Atoi(arr[0])
		y, _ := strconv.Atoi(arr[1])
		return x + OFFSET_X, y + OFFSET_Y
	}

	// 读取地图json文件，障碍物
	jsonBuf, err := ReadFileAll(ForeGroundFilePath)
	m := make(map[string]string)
	err = json.Unmarshal(jsonBuf, &m)
	if err != nil {
		glog.Errorln("[读取地图json文件错误] ", err)
		return false
	}
	for key, value := range m {
		_ = value
		x, y := splitPos(key)
		if x < 0 || x >= 40 || y < 0 || y >= 40 {
			// 坐标不合理
			continue
		}
		this.MapArray[x][y] = GridType_Obstacle
	}
	// 读取地图json文件，可移动的空地
	jsonBuf, err = ReadFileAll(BGFilePath)
	m = make(map[string]string)
	err = json.Unmarshal(jsonBuf, &m)
	if err != nil {
		glog.Errorln("[读取地图json文件错误] ", err)
		return false
	}
	for key, value := range m {
		_ = value
		x, y := splitPos(key)
		if x < 0 || x >= 40 || y < 0 || y >= 40 {
			// 坐标不合理
			continue
		}
		if this.MapArray[x][y] == GridType_Obstacle {
			continue
		}
		this.MapArray[x][y] = GridType_Space
	}
	glog.Infoln("[游戏地图初始化] 初始化完成")

	return true
}

// func (this *GameMap) CanPass(x, y int) bool {
// 	return this.MapArray[x][y] != GridType_Box &&
// 		this.MapArray[x][y] != GridType_Obstacle
// }

// func (this *GameMap) GetWidth() uint32 {
// 	return this.Width
// }

// func (this *GameMap) GetHeight() uint32 {
// 	return this.Height
// }
