package util

import (
	"cmp"
)

type Heap[T cmp.Ordered] []T

func (this Heap[T]) Len() int { return len(this) }

func (this Heap[T]) Less(i, j int) bool { return this[i] < this[j] }

func (this Heap[T]) Swap(i, j int) { this[i], this[j] = this[j], this[i] }

func (this *Heap[T]) Push(x interface{}) { *this = append(*this, x.(T)) }

func (this *Heap[T]) Pop() interface{} {
	a := (*this)
	x := a[len(a)-1]
	(*this) = a[:len(a)-1]
	return x
}
