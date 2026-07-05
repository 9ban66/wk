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
