package main

// --------------------------------AOI（九宫格）-------------------------------- //
const (
	AoiGridSize = 2                  // 每个aoi格子长度等于多少个游戏格子
	INT_MAX     = int(^uint(0) >> 1) // 0x7fffffff
)

type AoiGrid struct {
	index   uint32
	players map[uint64]*ScenePlayer
}

// 根据玩家坐标计算其所在aoi格子的编号
func CaculateAoiGrid(x, y uint32) uint32 {
	return (y/AoiGridSize)*(MaxGridNumberX/AoiGridSize) + (x/AoiGridSize + 1)
}

func (this *AoiGrid) AddPlayer(player *ScenePlayer) {
	this.players[player.id] = player
}

func (this *AoiGrid) DelPlayer(player *ScenePlayer) {
	delete(this.players, player.id)
}

// TODO 十字链表加入哨兵节点
// --------------------------------AOI（十字链表）-------------------------------- //
type ListNode struct {
	xPrev *ListNode
	xNext *ListNode

	yPrev *ListNode
	yNext *ListNode

	// 玩家位置信息
	x int
	y int

	p *ScenePlayer // 场景玩家指针
	// 哨兵标志 0代表是对象节点，1代表是前哨兵，2代表是后哨兵
	// flag byte
}

// 双向链表
type CrossLinkedList struct {
	head *ListNode // 头节点
	tail *ListNode // 尾节点
}

func (this *CrossLinkedList) Init() {
	this.head = &ListNode{x: 0, y: 0}
	this.tail = &ListNode{x: INT_MAX, y: INT_MAX}
	this.head.xNext = this.tail
	this.tail.xPrev = this.head
	this.head.yNext = this.tail
	this.tail.yPrev = this.head
}

func (this *CrossLinkedList) InsertPlayer(player *ScenePlayer) {
	x, y := player.GetCurrentGrid()
	node := &ListNode{x: int(x), y: int(y), p: player}
	// 新建两个哨兵节点
	gPrevNode := &ListNode{
		x: node.x - int(node.p.vision),
		y: node.y - int(node.p.vision),
		p: player}
	gNextNode := &ListNode{
		x: node.x + int(node.p.vision),
		y: node.y + int(node.p.vision),
		p: player}
	// 按顺序插入哨兵和节点
	this.Add(gPrevNode)
	this.Add(node)
	this.Add(gNextNode)
}

func (this *CrossLinkedList) PlayerMove(node *ListNode, x, y int) {
	// TODO
	this.RemovePlayer(node)
	node.x = x
	node.y = y
	this.Add(node)
}

func (this *CrossLinkedList) RemovePlayer(node *ListNode) {
	// x轴处理
	node.xPrev.xNext = node.xNext
	node.xNext.xPrev = node.xPrev
	// y轴处理
	node.yPrev.yNext = node.yNext
	node.yNext.yPrev = node.yPrev

	node.xNext, node.xPrev = nil, nil
	node.yNext, node.yPrev = nil, nil
}

func (this *CrossLinkedList) Add(node *ListNode) {
	// x轴处理
	cur := this.head.xNext
	for cur != nil {
		if cur.x > node.x {
			node.xNext = cur
			node.xPrev = cur.xPrev
			cur.xPrev.xNext = node
			cur.xPrev = node
			break
		}
		cur = cur.xNext
	}
	// y轴处理
	cur = this.head.yNext
	for cur != nil {
		if cur.y > node.y {
			node.yNext = cur
			node.yPrev = cur.yPrev
			cur.yPrev.yNext = node
			cur.yPrev = node
			break
		}
		cur = cur.yNext
	}
}

func (this *CrossLinkedList) GetRangeSet(node *ListNode) []*ScenePlayer {
	var cur *ListNode

	var xMap map[uint64]*ScenePlayer
	// X轴上往后遍历
	cur = node.xNext
	for cur != this.tail {
		// 距离超出当前玩家视野，跳出循环
		if cur.x-node.x > node.p.vision {
			break
		}
		xMap[cur.p.id] = cur.p
		cur = cur.xNext
	}
	// X轴上往前遍历
	cur = node.xPrev
	for cur != this.head {
		if node.x-cur.x > node.p.vision {
			break
		}
		xMap[cur.p.id] = cur.p
		cur = cur.xPrev
	}

	var res []*ScenePlayer
	// y轴上往后遍历
	cur = node.yNext
	for cur != this.tail {
		if cur.y-node.y > node.p.vision {
			break
		}
		if _, ok := xMap[cur.p.id]; ok {
			res = append(res, cur.p)
		}
		cur = cur.yNext
	}
	// y轴上往前遍历
	cur = node.yPrev
	for cur != this.head {
		if node.y-cur.y > node.p.vision {
			break
		}
		if _, ok := xMap[cur.p.id]; ok {
			res = append(res, cur.p)
		}
		cur = cur.yPrev
	}
	return res
}
