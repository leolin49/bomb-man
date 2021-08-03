package main

type Item interface{}

type Queue struct {
	Items []Item
}

type IQueue interface {
	New() Queue
	Push(t Item)
	Pop() *Item
	Empty() bool
	Size() int
}

func (q *Queue) New() *Queue {
	q.Items = []Item{}
	return q
}

func (q *Queue) Push(data Item) {
	q.Items = append(q.Items, data)
}

func (q *Queue) Pop() *Item {
	item := q.Items[0]
	q.Items = q.Items[1:len(q.Items)]
	return &item
}

func (q *Queue) Empty() bool {
	return len(q.Items) == 0
}

func (q *Queue) Size() int {
	return len(q.Items)
}
