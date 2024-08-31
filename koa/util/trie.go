package util

import "sync"

type TrieNode struct {
	children map[string]*TrieNode
	value    any
	end      bool
}

func NewTrieNode(e any) *TrieNode {
	ret := make(map[string]*TrieNode)
	return &TrieNode{
		children: ret,
		value:    e,
		end:      false,
	}
}

type Trie struct {
	sync.RWMutex
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		root: NewTrieNode(nil),
	}
}

func (t *Trie) Insert(word []string, value any) {
	current := t.root

	for i := 0; i < len(word); i++ {
		var ch = word[i]
		if child, ok := current.children[ch]; ok {
			current = child
		} else {
			current.children[ch] = NewTrieNode(nil)
			current = current.children[ch]
		}
	}

	current.value = value
	current.end = true
}

// 同时支持，搜索，前缀匹配，最长前缀匹配
func (t *Trie) Search(word []string) (latestValue any, isEnd bool) {
	current := t.root
	if current.end {
		latestValue = current.value
	}
	for i := 0; i < len(word); i++ {
		node := current.children[word[i]]
		if node == nil {
			return
		}
		current = node
		if current.end {
			latestValue = current.value
		}
	}
	isEnd = current.end
	return
}
