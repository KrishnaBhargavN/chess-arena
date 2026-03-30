package matchmaking

import (
	"slices"
	"sync"
)

type Queue struct {
	waiting []string
	mu sync.Mutex
}

type MatchResult struct {
	PlayerA string
	PlayerB string
}


func(q *Queue) JoinQueue(playerID string) *MatchResult {
	q.mu.Lock()
	defer q.mu.Unlock()
	if slices.Contains(q.waiting, playerID) {
		return nil
	}
	if len(q.waiting) > 0 {
		playerA := q.waiting[0]
		q.waiting = q.waiting[1:]
		return &MatchResult{
			PlayerA: playerA,
			PlayerB: playerID,
		}
	}
	q.waiting = append(q.waiting, playerID)
	return nil
}

func (q *Queue) LeaveQueue(playerID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	idx := slices.Index(q.waiting, playerID)
	if idx == -1 {
		return
	}
	q.waiting = slices.Delete(q.waiting, idx, idx+1)
	q.mu.Unlock()
}

func (q *Queue) ClearQueue() {
	q.mu.Lock()
	q.waiting = make([]string, 0)
	q.mu.Unlock()
}

func NewQueue() *Queue {
	return &Queue{
		waiting: make([]string, 0),
	}
}