package bufferpool_lru

import doublyll "github.com/nishanthgowda/btree/lru/doubly-ll"

type BufferPool struct {
	capacity       int
	cache          map[int]*doublyll.Node
	list           *doublyll.DoublyLinkedList
	MidPoint       *doublyll.Node
	OldRatio       float64
	MaxOldSize     int // Renamed from OldListSize for clarity (The Target)
	currentOldSize int // The actual counter
}

func NewBufferPool(capacity int, ratio float64) *BufferPool {
	return &BufferPool{
		capacity:       capacity,
		cache:          make(map[int]*doublyll.Node),
		list:           doublyll.NewDoublyLinkedList(),
		MidPoint:       nil,
		OldRatio:       ratio,
		MaxOldSize:     int(float64(capacity) * ratio),
		currentOldSize: 0,
	}
}

// insertAtMidpoint inserts a node at the head of the "Old" sublist
func (bp *BufferPool) insertAtMidpoint(node *doublyll.Node) {

	// Mark the node as "Old"
	node.IsOld = true

	// Case 1: If the list is empty, just add it to the front
	if bp.list.Head == nil {
		bp.list.AddFront(node)
		bp.MidPoint = node
		bp.currentOldSize++
		return
	}

	// Case 2: If MidPoint is currently the Head (meaning everything is "Old")
	// We AddFront, and this new node stays the "Old" head (MidPoint)
	if bp.MidPoint == bp.list.Head {
		bp.list.AddFront(node)
		bp.MidPoint = node
		bp.currentOldSize++
		return
	}

	// Case 3: MidPoint is somewhere in the middle
	// We physically splice the node into the list right BEFORE the current MidPoint

	// 1. Set the new node's pointers
	node.Next = bp.MidPoint      // It points forward to the old MidPoint
	node.Prev = bp.MidPoint.Prev // It points backward to the New List's tail

	// 2. Update the "New List" tail to point to our new node
	if bp.MidPoint.Prev != nil {
		bp.MidPoint.Prev.Next = node
	}

	// 3. Update the old MidPoint to point back to our new node
	bp.MidPoint.Prev = node

	// 4. Finally, update the MidPoint marker to this new node
	bp.MidPoint = node
	bp.currentOldSize++
}

func (bp *BufferPool) Put(key int, value int) {

	// 1. Handle Update (If key exists) Treat it as a Get
	if node, ok := bp.cache[key]; ok {
		node.Value = value
        
        // Exact same promotion logic as Get
        if node.IsOld {
            node.IsOld = false
            bp.currentOldSize--
            if bp.MidPoint == node {
                bp.MidPoint = node.Next
            }
        }
        
        bp.list.RemoveNode(node)
        bp.list.AddFront(node)
        return
	}

	// 2. Capacity Check & Eviction
	if len(bp.cache) == bp.capacity {
		nodeToEvict := bp.list.Tail
		bp.list.RemoveNode(nodeToEvict)
		delete(bp.cache, nodeToEvict.Key)
		bp.currentOldSize--

		// Update MidPoint if it was evicted
		if bp.MidPoint == nodeToEvict {
			bp.MidPoint = nil
		}
	}

	// 3. Create and Insert New Node
	newNode := &doublyll.Node{Key: key, Value: value}
	bp.cache[key] = newNode
	bp.insertAtMidpoint(newNode) // Use our helper

	// 4. Move MidPoint if OldListSize exceeds MaxOldSize
	if bp.currentOldSize > bp.MaxOldSize {
		if bp.MidPoint != nil {
			bp.MidPoint = bp.MidPoint.Next
			bp.currentOldSize--
		}
	}
}

func (bp *BufferPool) Get(key int) int {
	node, ok := bp.cache[key]
	if !ok {
		return -1
	}

	if node.IsOld {
		node.IsOld = false
		bp.currentOldSize--
		if node == bp.MidPoint {
			bp.MidPoint = bp.MidPoint.Next
		}
	}

	bp.list.RemoveNode(node)
	bp.list.AddFront(node)

	return node.Value
}
