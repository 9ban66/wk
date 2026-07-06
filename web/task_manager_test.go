package web

import "testing"

func TestTaskStoreSubmitAndList(t *testing.T) {
	store := newTaskStore()
	task := store.enqueue(Task{Platform: "haiqikeji", Account: "demo", Password: "secret", PreURL: "https://example.com"})
	if task.ID == "" {
		t.Fatal("expected task id to be generated")
	}
	if task.Status != "queued" {
		t.Fatalf("expected status queued, got %s", task.Status)
	}
	items := store.list()
	if len(items) != 1 {
		t.Fatalf("expected 1 task, got %d", len(items))
	}
	if items[0].Platform != "haiqikeji" {
		t.Fatalf("expected platform haiqikeji, got %s", items[0].Platform)
	}
}

func TestTaskStoreAppendLogOnlyKeepsMessage(t *testing.T) {
	store := newTaskStore()
	task := store.enqueue(Task{Platform: "haiqikeji", Account: "demo", Password: "secret", Message: "failed reason"})

	if !store.appendLogOnly(task.ID, "info", "执行结束") {
		t.Fatal("expected log append to succeed")
	}

	updated, ok := store.get(task.ID)
	if !ok {
		t.Fatal("expected task to exist")
	}
	if updated.Message != "failed reason" {
		t.Fatalf("expected message to stay unchanged, got %q", updated.Message)
	}
	if len(updated.Logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(updated.Logs))
	}
}

func TestTaskStoreControlTimestamps(t *testing.T) {
	store := newTaskStore()
	task := store.enqueue(Task{Platform: "haiqikeji", Account: "demo", Password: "secret"})

	if !store.start(task.ID) {
		t.Fatal("expected start to succeed")
	}
	running, _ := store.get(task.ID)
	if running.Status != TaskRunning {
		t.Fatalf("expected running status, got %s", running.Status)
	}
	if running.StartedAt == nil {
		t.Fatal("expected started time to be recorded")
	}

	if !store.pause(task.ID) {
		t.Fatal("expected pause to succeed")
	}
	paused, _ := store.get(task.ID)
	if paused.Status != TaskPaused {
		t.Fatalf("expected paused status, got %s", paused.Status)
	}

	if !store.stop(task.ID) {
		t.Fatal("expected stop to succeed")
	}
	stopped, _ := store.get(task.ID)
	if stopped.Status != TaskStopped {
		t.Fatalf("expected stopped status, got %s", stopped.Status)
	}
	if stopped.EndedAt == nil {
		t.Fatal("expected ended time to be recorded")
	}
}
