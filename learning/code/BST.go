/*
IMPLEMENTING THE BINARY SEARCH TREE
  - insert
*/
package main

import "fmt"

type Node struct {
	left  *Node
	right *Node
	key   int
}

func (n *Node) insert(val int) {
	if val > n.key {
		// Right
		if n.right != nil {
			n.right.insert(val)
		} else {
			n.right = &Node{key: val}
		}
	} else {
		// Left
		if n.left != nil {
			n.left.insert(val)
		} else {
			n.left = &Node{key: val}
		}
	}
}

func (n *Node) search(k int) bool {
	if n.key != 0 {
		if n.key == k {
			return true
		} else if k < n.key {
			if n.left != nil {
				if n.left.key == k {
					return true
				} else {
					return n.left.search(k)
				}
			}
		} else {
			if n.right != nil {
				if n.right.key == k {
					return true
				} else {
					return n.right.search(k)
				}
			}
		}
	}
	return false
}

func (n *Node) delete(key int) {
	if n.key != 0 {
		if n.key == key {
			n.key = 0
			n.left = nil
			n.right = nil
		} else if key < n.key {
			// left
			if n.left.key != 0 {
				if n.left.key == key {
					n.key = 0
					n.left.left = nil
					n.right.right = nil
				} else {
					n.left.delete(key)
				}
			}
		} else {
			if n.right.key != 0 {
				if n.right.key == key {
					n.key = 0
					n.right.right = nil
					n.right.right = nil
				} else {
					n.right.delete(key)
				}
			}
		}
	}
}

func bst() {
	bst := &Node{key: 10}
	bst.insert(20)
	bst.insert(5)
	bst.insert(2)
	bst.insert(1000)
	bst.insert(100)
	bst.insert(300)

	fmt.Println(bst)
	fmt.Println(bst.search(300))
	fmt.Println(bst.search(20))
	bst.delete(5)
	fmt.Println(bst.search(5))
	fmt.Println(bst.search(2))
}
