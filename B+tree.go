package BTREE

import (
	"bytes"
	"encoding/binary"

	"github.com/lokicodess/myDb/utils"
)

const (
	HEADER             = 4
	BTREE_PAGE_SIZE    = 4096 // analogy to node size
	BTREE_MAX_KEY_SIZE = 1000
	BTREE_MAX_VAL_SIZE = 3000
	BNODE_NODE         = 1
	BNODE_LEAF         = 2
)

type BNode []byte

type BTree struct {
	root uint64
	get  func(uint64) []byte
	new  func(BNode) uint64 // prev it was accepting bytes
	del  func(uint64)
}

// HEADER Manipulation
// Endianess meaning storing bytes in the sequence/order in the memory
// binary.LittleEndian.Uint16 only reads the 2 bytes
func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node[0:3])
}

func (node BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

func (node BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], btype)
	binary.LittleEndian.PutUint16(node[2:4], nkeys)
}

// POINTERS

func (node BNode) getPtr(idx uint16) uint64 {
	utils.Assert(idx > node.nkeys(), "getPtr: Index out of Bounds!")
	pos := HEADER + 8*idx
	return binary.LittleEndian.Uint64(node[pos:])
}

func (node BNode) setPtr(idex uint16, val uint64)

// OFFSET

func offset(node BNode, idx uint16) uint16 {
	utils.Assert(1 <= idx && idx <= node.nkeys(), "offset: Index out of Bounds!")
	return HEADER + 8*node.nkeys() + 2*(idx-1) // only position
}

// @getOffset returns the offset postion
func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}
	return binary.LittleEndian.Uint16(node[offset(node, idx):])
}

func (node BNode) setOffset(idx, offset uint16)

// @kvPos returns the kvPositon(key,val -->  len), not actual key and value
func (node BNode) kvPos(idx uint16) uint16 {
	utils.Assert(idx <= node.nkeys(), "kvPos: Index out of Bounds!")
	return HEADER + 8*node.nkeys() + 2*node.nkeys() + node.getOffset(idx) // getoffset only returns the 2Bytes
}

func (node BNode) getKey(idx uint16) []byte { // actually  returns the original key
	utils.Assert(idx < node.nkeys(), "getKey: Index out of Bounds!")
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node[pos:])
	return node[pos+4:][klen:] // +4 because 2 for keyLen and 2 for valLen
}

func (node BNode) getVal(idx uint16) []byte

// node size
func (node BNode) nbytes() uint16 {
	return node.kvPos(node.nkeys()) // n is nth index
}

// @nodeLookupLE finds the key in the current node
func nodeLookupLE(node BNode, key []byte) uint16 {
	nkeys := node.nkeys()
	found := uint16(0)

	for i := uint16(1); i < nkeys; i++ {
		cmp := bytes.Compare(node.getKey(i), key)
		if cmp <= 0 {
			found = i
		}
		if cmp >= 0 {
			break
		}
	}
	return found
}

// NODE APPEND FUNCTION

func leafInsert(
	new BNode, old BNode, idx uint16,
	key []byte, val []byte,
) {
	// maybe temporarly storing n + 1 keys instead of n/2
	// it will help in redistribution
	new.setHeader(BNODE_LEAF, old.nkeys()+1)
	nodeAppendRange(new, old, 0, 0, idx)
	nodeAppendKV(new, idx, 0, key, val)
	nodeAppendRange(new, old, idx+1, idx, old.nkeys()-idx)
}

func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, val []byte) {
	new.setPtr(idx, ptr)
	pos := new.kvPos(idx)                                        // will be pointing to starting of keyLen
	binary.LittleEndian.PutUint16(new[pos+0:], uint16(len(key))) // should start from here
	binary.LittleEndian.PutUint16(new[pos+2:], uint16(len(val)))
	copy(new[pos+4:], key)
	copy(new[pos+4+uint16(len(key)):], val)
	new.setOffset(idx+1, new.getOffset(idx)+4+uint16((len(key)+len(val)))) // will see when implement the function
}

func nodeAppendRange(new BNode, old BNode, dstNew uint16, srcOld uint16, n uint16)

// internal nodes store keys and pointers to child nodes
func nodeReplaceKidN(
	tree *BTree, new BNode, old BNode, idx uint16,
	kids ...BNode,
) {
	inc := uint16(len(kids))
	new.setHeader(BNODE_NODE, old.nkeys()+inc-1)
	nodeAppendRange(new, old, 0, 0, idx)
	for i, node := range kids {
		nodeAppendKV(new, idx+uint16(i), tree.new(node), node.getKey(0), nil)
	}
	nodeAppendRange(new, old, idx+inc, idx+1, old.nkeys()-(idx+1))
}

// SPLIT FUNCTIONS

func nodeSplit2(left BNode, right BNode, old BNode) {
}

func nodeSplit3(old BNode) (uint16, [3]BNode) {
	if old.nbytes() <= BTREE_PAGE_SIZE {
		old = old[:BTREE_PAGE_SIZE]
		return 1, [3]BNode{old}
	}
	left := BNode(make([]byte, 2*BTREE_PAGE_SIZE))
	right := BNode(make([]byte, 2*BTREE_PAGE_SIZE))
	nodeSplit2(left, right, old)
	if left.nbytes() <= BTREE_PAGE_SIZE {
		left = left[:BTREE_PAGE_SIZE]
		return 2, [3]BNode{left, right} // 2 nodes
	}
	// After splitting its still > 4KB
	leftleft := BNode(make([]byte, BTREE_PAGE_SIZE)) // splitting left more
	middle := BNode(make([]byte, BTREE_PAGE_SIZE))
	nodeSplit2(leftleft, middle, left)
	utils.Assert(leftleft.nbytes() <= BTREE_PAGE_SIZE, "nodeSplit3: leftleft.nbytes smaller than node size")
	return 3, [3]BNode{leftleft, middle, right}
}

// B+Tree insertion

func treeInsert(tree *BTree, node BNode, key []byte, val []byte) BNode {
	new := make(BNode, 2*BTREE_PAGE_SIZE)
	idx := nodeLookupLE(node, key)
	switch node.btype() {
	case BNODE_LEAF:
		if bytes.Equal(key, node.getKey(idx)) {
			leafUpdate(new, node, idx, key, val)
		} else {
			leafInsert(new, node, idx+1, key, val)
		}
	case BNODE_NODE:
		nodeInsert(tree, new, node, idx, key, val)
	default:
		panic("bad node!")

	}
	return new
}

func nodeInsert(
	tree *BTree, new BNode, node BNode, idx uint16, key []byte, val []byte,
) {
	kptr := node.getPtr(idx)
	knode := treeInsert(tree, tree.get(kptr), key, val) // child node
	nsplit, split := nodeSplit3(knode)
	tree.del(kptr) // removing refrence
	nodeReplaceKidN(tree, new, node, idx, split[:nsplit]...)
}

func leafUpdate(
	new BNode, old BNode, idx uint16,
	key []byte, val []byte,
) {
	// maybe temporarly storing n + 1 keys instead of n/2
	// it will help in redistribution
	new.setHeader(BNODE_LEAF, old.nkeys()+1)
	nodeAppendRange(new, old, 0, 0, idx)
	nodeAppendKV(new, idx, 0, key, val)
	nodeAppendRange(new, old, idx+1, idx, old.nkeys()-idx)
}

func (tree *BTree) Insert(key []byte, val []byte) {
	if tree.root == 0 {
		// create the first node
		root := make(BNode, BTREE_PAGE_SIZE)
		root.setHeader(BNODE_LEAF, 2)
		nodeAppendKV(root, 0, 0, nil, nil)
		nodeAppendKV(root, 1, 0, key, val)
		tree.root = tree.new(root)
		return
	}

	node := treeInsert(tree, tree.get(tree.root), key, val)
	nsplit, split := nodeSplit3(node)
	tree.del(tree.root)
	if nsplit > 1 {
		// the root was split, add a new level
		root := make(BNode, BTREE_PAGE_SIZE)
		root.setHeader(BNODE_NODE, nsplit)
		for i, knode := range split[:nsplit] {
			ptr, key := tree.new(knode), knode.getKey(0)
			nodeAppendKV(root, uint16(i), ptr, key, nil)
		}
		tree.root = tree.new(root)
	} else {
		tree.root = tree.new(split[0])
	}
}
func (tree *BTree) Delete(key []byte)

func leafDelete(new BNode, old BNode, idx uint16)
func nodeMerge(new BNode, left BNode, right BNode)
func nodeReplace2Kid(
	new BNode, old BNode, idx uint16, ptr uint64, key []byte,
)
func shouldMerge(tree *BTree, node BNode, idx uint16, updated BNode) (int, BNode) {
	if updated.nbytes() > BTREE_PAGE_SIZE/4 {
		return 0, BNode{}
	}

	if idx > 0 {
		sibling := BNode(tree.get(node.getPtr(idx - 1)))
		merged := sibling.nbytes() + updated.nbytes() - HEADER
		if merged <= BTREE_PAGE_SIZE {
			return -1, sibling // left
		}
	}
	if idx+1 < node.nkeys() {
		sibling := BNode(tree.get(node.getPtr(idx + 1)))
		merged := sibling.nbytes() + updated.nbytes() - HEADER
		if merged <= BTREE_PAGE_SIZE {
			return +1, sibling // right
		}
	}
	return 0, BNode{}
}

func treeDelete(tree *BTree, node BNode, key []byte) BNode

func nodeDelete(tree *BTree, node BNode, idx uint16, key []byte) BNode {
	kptr := node.getPtr(idx)
	updated := treeDelete(tree, tree.get(kptr), key)
	if len(updated) == 0 {
		return BNode{}
	}
	tree.del(kptr)
	new := make(BNode, BTREE_PAGE_SIZE)
	mergeDir, sibling := shouldMerge(tree, node, idx, updated)
	switch {
	case mergeDir < 0:
		merged := make(BNode, BTREE_PAGE_SIZE)
		nodeMerge(merged, sibling, updated)
		tree.del(node.getPtr(idx - 1))
		nodeReplace2Kid(new, node, idx-1, tree.new(merged), merged.getKey(0))
	case mergeDir > 0:
		merged := make(BNode, BTREE_PAGE_SIZE)
		nodeMerge(merged, updated, sibling)
		tree.del(node.getPtr(idx + 1))
		nodeReplace2Kid(new, node, idx, tree.new(merged), merged.getKey(0))
	case mergeDir == 0 && updated.nkeys() == 0:
		utils.Assert(node.nkeys() == 1, "nodeDelete: node.nkeys() != 1")
		new.setHeader(BNODE_NODE, 0)
	case mergeDir == 0 && updated.nkeys() > 0:
		nodeReplaceKidN(tree, new, node, idx, updated)
	}
	return new
}
