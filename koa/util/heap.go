package util

import "container/heap"

type HeapCompareFunc[T any] func(a, b T) bool

type heapList[T any] struct {
	compare HeapCompareFunc[T]
	data    []T
}

func (this heapList[T]) Len() int { return len(this.data) }

func (this heapList[T]) Less(i, j int) bool { return this.compare(this.data[i], this.data[j]) }

func (this heapList[T]) Swap(i, j int) { this.data[i], this.data[j] = this.data[j], this.data[i] }

func (this *heapList[T]) Push(x interface{}) { this.data = append(this.data, x.(T)) }

func (this *heapList[T]) Pop() interface{} {
	array := (this.data)
	popedItem := array[len(array)-1]
	this.data = array[:len(array)-1]
	return popedItem
}

type Heap[T any] struct {
	heapData heapList[T]
}

func NewHeap[T any](compare HeapCompareFunc[T], data []T) *Heap[T] {
	return &Heap[T]{
		heapData: heapList[T]{compare, data},
	}
}

func (this *Heap[T]) Len() int {
	return this.heapData.Len()
}

func (this *Heap[T]) Push(x T) {
	heap.Push(&this.heapData, x)
}

func (this *Heap[T]) Pop() T {
	return heap.Pop(&this.heapData).(T)
}

func (this *Heap[T]) Init() {
	heap.Init(&this.heapData)
}

func (this *Heap[T]) Fix(i int) {
	heap.Fix(&this.heapData, i)
}

// func mycmp(a, b int) bool { return a < b }
func (this *Heap[T]) Remove(i int) T {
	return heap.Remove(&this.heapData, i).(T)
}

/**
使用示例
score := []int{1, 2, 3}
var h = NewHeap(HeapCompareFunc[int](func(a, b int) bool { return a < b }), score)
h.Init()

这里不为比较比较函数单独定义类型也是可以的。因类alias不能用泛型，所以这里定义了新类型。
*/
