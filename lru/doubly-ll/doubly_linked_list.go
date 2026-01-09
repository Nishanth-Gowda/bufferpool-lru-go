package doublyll

// Node represents a node in the doubly linked list
type Node struct {
	Key   int
	Value int
	Prev  *Node
	Next  *Node
	IsOld bool
}

// DoublyLinkedList represents a doubly linked list
type DoublyLinkedList struct {
	Head *Node
	Tail *Node
}

func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{
		Head: nil,
		Tail: nil,
	}
}

func (list *DoublyLinkedList) AddFront(node *Node) {
	if list.Head == nil {
		list.Head = node
		list.Tail = node
		node.Prev = nil
		node.Next = nil
		return
	}

	node.Prev = nil // Clear prev pointer to avoid dangling references
	node.Next = list.Head
	list.Head.Prev = node
	list.Head = node
}

func (list *DoublyLinkedList) RemoveNode(node *Node) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	}

	if node == list.Head {
		list.Head = node.Next
	}

	if node == list.Tail {
		list.Tail = node.Prev
	}

	// Clear the node's pointers to help GC and prevent dangling references
	node.Prev = nil
	node.Next = nil
}
