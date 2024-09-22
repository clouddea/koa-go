package util

type LRUCache[T comparable] struct {
	size       int
	capacity   int
	cache      map[T]*DLinkedNode[T]
	head, tail *DLinkedNode[T]
}

type DLinkedNode[T comparable] struct {
	key        T
	value      any
	prev, next *DLinkedNode[T]
}

func initDLinkedNode[T comparable](key T, value any) *DLinkedNode[T] {
	return &DLinkedNode[T]{
		key:   key,
		value: value,
	}
}

func NewLRUCache[T comparable](capacity int) *LRUCache[T] {
	var initVal T
	l := LRUCache[T]{
		cache:    map[T]*DLinkedNode[T]{},
		head:     initDLinkedNode[T](initVal, nil),
		tail:     initDLinkedNode[T](initVal, nil),
		capacity: capacity,
	}
	l.head.next = l.tail
	l.tail.prev = l.head
	return &l
}

func (this *LRUCache[T]) Get(key T) (any, bool) {
	if _, ok := this.cache[key]; !ok {
		return nil, false
	}
	node := this.cache[key]
	this.moveToHead(node)
	return node.value, true
}

func (this *LRUCache[T]) Put(key T, value any) {
	if _, ok := this.cache[key]; !ok {
		node := initDLinkedNode(key, value)
		this.cache[key] = node
		this.addToHead(node)
		this.size++
		if this.size > this.capacity {
			removed := this.removeTail()
			delete(this.cache, removed.key)
			this.size--
		}
	} else {
		node := this.cache[key]
		node.value = value
		this.moveToHead(node)
	}
}

func (this *LRUCache[T]) addToHead(node *DLinkedNode[T]) {
	node.prev = this.head
	node.next = this.head.next
	this.head.next.prev = node
	this.head.next = node
}

func (this *LRUCache[T]) removeNode(node *DLinkedNode[T]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (this *LRUCache[T]) moveToHead(node *DLinkedNode[T]) {
	this.removeNode(node)
	this.addToHead(node)
}

func (this *LRUCache[T]) removeTail() *DLinkedNode[T] {
	node := this.tail.prev
	this.removeNode(node)
	return node
}
