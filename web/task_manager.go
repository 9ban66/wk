package web

import (
	"fmt"
	"sync"
	"time"
)

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string     `json:"id"`
	Platform  string     `json:"platform"`
	Account   string     `json:"account"`
	Password  string     `json:"password"`
	PreURL    string     `json:"preUrl"`
	AIURL     string     `json:"aiUrl"`
	AIModel   string     `json:"aiModel"`
	AIKey     string     `json:"aiKey"`
	AIType    string     `json:"aiType"`
	Status    TaskStatus `json:"status"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type taskStore struct {
	mu      sync.Mutex
	tasks   []Task
	byID    map[string]*Task
	counter int
}

func newTaskStore() *taskStore {
	return &taskStore{tasks: []Task{}, byID: make(map[string]*Task)}
}

func (s *taskStore) enqueue(task Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", s.counter+1)
		s.counter++
	}
	task.Status = TaskQueued
	task.CreatedAt = time.Now()
	task.UpdatedAt = task.CreatedAt
	s.tasks = append(s.tasks, task)
	s.byID[task.ID] = &s.tasks[len(s.tasks)-1]
	return s.tasks[len(s.tasks)-1]
}

func (s *taskStore) list() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Task, len(s.tasks))
	copy(out, s.tasks)
	return out
}

func (s *taskStore) get(id string) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.byID[id]
	if !ok {
		return Task{}, false
	}
	return *task, true
}

func (s *taskStore) update(id string, fn func(*Task)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.byID[id]
	if !ok {
		return false
	}
	fn(task)
	task.UpdatedAt = time.Now()
	return true
}
