package taskscheduler

type PriorityQueue struct {
}

func NewPriorityQueue() *PriorityQueue {
	return nil
}

func (pq *PriorityQueue) Push(task *Task) {
}

func (pq *PriorityQueue) Pop() *Task {
	return nil
}

func (pq *PriorityQueue) Peek() *Task {
	return nil
}

func (pq *PriorityQueue) Len() int {
	return 0
}

func (pq *PriorityQueue) Remove(taskID string) bool {
	return false
}

func (pq *PriorityQueue) Update(taskID string, newPriority Priority) {
}
