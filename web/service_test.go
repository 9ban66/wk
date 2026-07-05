package web

import (
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
