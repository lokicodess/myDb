package main

import "fmt"

var DefaultMinItems = 128

type Item struct {
	key   string
	value interface{}
}

type Node struct {
	bucket     *Tree
	items      []*Item
	childNodes []*Node
}

type Tree struct {
	root     *Node
	minItems int
	maxItems int
}

func newItem(key string, value interface{}) *Item {
	return &Item{key: key, value: value}
}

func NewTreeWithRoot(root *Node, minItems int) *Tree {
	bucket := &Tree{
		root:     root,
		minItems: minItems,
	}
	bucket.root.bucket = bucket
	bucket.minItems = minItems
	bucket.maxItems = minItems * 2

	return bucket
}

func NewTree(minItems int) *Tree {
	return NewTreeWithRoot(newEmptyRoot(), minItems)
}

func newEmptyRoot() *Node {
	return &Node{
		items:      []*Item{},
		childNodes: []*Node{},
	}
}

func (b *Tree) Put(key string, value interface{}) {
	i := newItem(key, value)
	insertionIndex, nodeToInsertIn, ancestorsIndexes := b.findKey(i.key, false)

	nodeToInsertIn.addItem(insertionIndex, i)
	// Get nodes form the root to the node
	ancestors := b.getNodes(ancestorsIndexes)

	for i := len(ancestorsIndexes) - 2; i >= 0; i-- {
		pnode := ancestors[i]
		node := ancestors[i+1]
		nodeIndex := ancestorsIndexes[i+1]
		if node.isOverPopulated() {
			pnode.split(node, nodeIndex)
		}
	}

	if b.root.isOverPopulated() {
		newRoot := NewNode(b, []*Item{}, []*Node{b.root})
		newRoot.split(b.root, 0)
		b.root = newRoot
	}
}

func (b *Tree) findKey(key string, exact bool) (int, *Node, []int) {
	n := b.root
	ancestorsIndexes := []int{0}

	for true {
		// checking in current node
		wasFound, index := n.findKey(key)
		if wasFound {
			return index, n, ancestorsIndexes
		} else {
			if n.isLeaf() {
				if exact {
					return -1, nil, nil
				}
				return index, n, ancestorsIndexes
			}
			nextChild := n.childNodes[index]
			ancestorsIndexes = append(ancestorsIndexes, index)
			n = nextChild
		}
	}
	return -1, nil, nil
}

func (n *Node) findKey(key string) (bool, int) {
	for i, existingItem := range n.items {
		if key == existingItem.key {
			return true, i
		}
		if key < existingItem.key {
			return false, i
		}
	}
	return false, len(n.items)
}

func (n *Node) addItem(insertionIndex int, item *Item) int {
	if len(n.items) == insertionIndex {
		n.items = append(n.items, item)
		return insertionIndex
	}

	n.items = append(n.items[:insertionIndex+1], n.items[insertionIndex:]...)
	n.items[insertionIndex] = item

	return insertionIndex
}

func (b *Tree) getNodes(indexes []int) []*Node {
	nodes := []*Node{b.root}
	child := b.root

	for i := 1; i < len(indexes); i++ {
		child = child.childNodes[i]
		nodes = append(nodes, child)
	}
	return nodes
}

func (n *Node) isLeaf() bool {
	return len(n.childNodes) == 0
}

func (n *Node) isOverPopulated() bool {
	return len(n.items) > n.bucket.maxItems
}

func NewNode(bucket *Tree, value []*Item, childNodes []*Node) *Node {
	return &Node{
		bucket,
		value,
		childNodes,
	}
}

func (n *Node) split(modifiedItem *Node, insertionIndex int) {
	i := 0
	nodeSize := n.bucket.minItems

	for modifiedItem.isOverPopulated() {
		middleItem := modifiedItem.items[nodeSize]
		var node *Node
		if modifiedItem.isLeaf() {
			node = NewNode(modifiedItem.bucket, modifiedItem.items[nodeSize+1:], []*Node{})
			modifiedItem.items = modifiedItem.items[:nodeSize]
		} else {
			node = NewNode(modifiedItem.bucket, modifiedItem.items[nodeSize+1:], modifiedItem.childNodes[i+1:])
			modifiedItem.items = modifiedItem.items[:nodeSize]
			modifiedItem.childNodes = modifiedItem.childNodes[:nodeSize+1]
		}
		n.addItem(insertionIndex, middleItem)

		if len(n.childNodes) == insertionIndex+1 {
			n.childNodes = append(n.childNodes, node)
		} else {
			n.childNodes = append(n.childNodes[:nodeSize+1], n.childNodes[nodeSize:]...)
			n.childNodes[insertionIndex] = node
		}

		i += 1
		insertionIndex += 1
		modifiedItem = node
	}
}

func main() {
	Tree := NewTree(DefaultMinItems)
	value := "0"
	Tree.Put(value, value)
	fmt.Println(Tree)
}
