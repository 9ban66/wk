package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

func TestCountXueXiTongTaskPoints(t *testing.T) {
	videoDTOs := []xuexitongApi.PointVideoDto{{IsSet: true}, {IsSet: true}}
	documentDTOs := []xuexitongApi.PointDocumentDto{{IsSet: true}}
	hyperlinkDTOs := []xuexitongApi.PointHyperlinkDto{}
	liveDTOs := []xuexitongApi.PointLiveDto{{IsSet: true}}

	count := countXueXiTongTaskPoints(videoDTOs, documentDTOs, hyperlinkDTOs, liveDTOs)
	if count != 4 {
		t.Fatalf("expected 4 task points, got %d", count)
	}
}

func TestTaskStorePreservesAISettings(t *testing.T) {
	store := newTaskStore()
	task := store.enqueue(Task{Platform: "xuexitong", Account: "demo", Password: "secret", AIURL: "https://example.com", AIModel: "gpt-4o", AIKey: "secret", AIType: "OPENAI"})
	if task.AIURL != "https://example.com" || task.AIModel != "gpt-4o" || task.AIKey != "secret" || task.AIType != "OPENAI" {
		t.Fatalf("expected AI settings to be preserved, got %+v", task)
	}
}

func TestCleanStringList(t *testing.T) {
	got := cleanStringList([]string{" 101 ", "", "102", "101", "  "})
	if len(got) != 2 || got[0] != "101" || got[1] != "102" {
		t.Fatalf("unexpected cleaned values: %#v", got)
	}
}

func TestAdminTasksRequireKey(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/admin/tasks", nil)
	rec := httptest.NewRecorder()

	server.adminTasksHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rec.Code)
	}
}

func TestAdminLoginAndClearLogs(t *testing.T) {
	t.Setenv("ADMIN_KEY", "test-key")
	server := NewServer()
	task := server.store.enqueue(Task{Platform: "haiqikeji", Account: "demo"})
	server.store.appendLog(task.ID, "info", "hello")

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString(`{"key":"test-key"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	server.adminLoginHandler(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login ok, got %d: %s", loginRec.Code, loginRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/admin/tasks/"+task.ID+"/logs", nil)
	for _, cookie := range loginRec.Result().Cookies() {
		deleteReq.AddCookie(cookie)
	}
	deleteRec := httptest.NewRecorder()
	server.adminTaskHandler(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("expected clear logs ok, got %d: %s", deleteRec.Code, deleteRec.Body.String())
	}
	updated, _ := server.store.get(task.ID)
	if len(updated.Logs) != 0 {
		t.Fatalf("expected logs to be cleared, got %d", len(updated.Logs))
	}
}
