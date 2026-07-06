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
	TaskPaused    TaskStatus = "paused"
	TaskStopped   TaskStatus = "stopped"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
)

type Task struct {
	ID             string     `json:"id"`
	Platform       string     `json:"platform"`
	Account        string     `json:"account"`
	Password       string     `json:"password"`
	PreURL         string     `json:"preUrl"`
	CourseIDs      []string   `json:"courseIds"`
	AIURL          string     `json:"aiUrl"`
	AIModel        string     `json:"aiModel"`
	AIKey          string     `json:"aiKey"`
	AIType         string     `json:"aiType"`
	Status         TaskStatus `json:"status"`
	Message        string     `json:"message"`
	Logs           []TaskLog  `json:"logs"`
	CreatedAt      time.Time  `json:"createdAt"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	EndedAt        *time.Time `json:"endedAt,omitempty"`
	RuntimeSeconds int64      `json:"runtimeSeconds"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type TaskLog struct {
	At      time.Time `json:"at"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type taskStore struct {
	mu      sync.Mutex
	tasks   []*Task
	byID    map[string]*Task
	counter int
}

func newTaskStore() *taskStore {
	return &taskStore{tasks: []*Task{}, byID: make(map[string]*Task)}
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
	item := &task
	s.tasks = append(s.tasks, item)
	s.byID[task.ID] = item
	return *item
}

func (s *taskStore) list() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		out = append(out, *task)
	}
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

func (s *taskStore) start(id string) bool {
	now := time.Now()
	return s.update(id, func(task *Task) {
		if task.Status == TaskQueued && task.StartedAt == nil {
			task.StartedAt = &now
		}
		if task.Status == TaskQueued || task.Status == TaskPaused {
			task.Status = TaskRunning
			task.Message = "任务运行中"
		}
	})
}

func (s *taskStore) pause(id string) bool {
	return s.update(id, func(task *Task) {
		if task.Status == TaskRunning {
			task.Status = TaskPaused
			task.Message = "任务已暂停"
		}
	})
}

func (s *taskStore) stop(id string) bool {
	now := time.Now()
	return s.update(id, func(task *Task) {
		if task.Status == TaskSucceeded || task.Status == TaskFailed || task.Status == TaskStopped {
			return
		}
		task.Status = TaskStopped
		task.EndedAt = &now
		task.RuntimeSeconds = task.runtimeSeconds(now)
		task.Message = "任务已停止"
	})
}

func (s *taskStore) finish(id string, status TaskStatus, message string) bool {
	now := time.Now()
	return s.update(id, func(task *Task) {
		if task.Status == TaskStopped {
			return
		}
		task.Status = status
		task.EndedAt = &now
		task.RuntimeSeconds = task.runtimeSeconds(now)
		task.Message = message
	})
}

func (t Task) runtimeSeconds(now time.Time) int64 {
	if t.StartedAt == nil {
		return 0
	}
	end := now
	if t.EndedAt != nil {
		end = *t.EndedAt
	}
	if end.Before(*t.StartedAt) {
		return 0
	}
	return int64(end.Sub(*t.StartedAt).Seconds())
}

func (s *taskStore) appendLog(id, level, message string) bool {
	return s.update(id, func(task *Task) {
		task.Logs = append(task.Logs, TaskLog{
			At:      time.Now(),
			Level:   level,
			Message: message,
		})
		task.Message = message
	})
}

func (s *taskStore) appendLogOnly(id, level, message string) bool {
	return s.update(id, func(task *Task) {
		task.Logs = append(task.Logs, TaskLog{
			At:      time.Now(),
			Level:   level,
			Message: message,
		})
	})
}
