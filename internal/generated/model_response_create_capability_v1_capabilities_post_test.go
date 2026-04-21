// Copyright (c) Trifork

package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestResponseCreateCapability_DispatchesOnType(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		capType          string
		completionPrompt string
		wantChat         bool
		wantCompletion   bool
		wantExtraction   bool
		wantSpeechToText bool
	}{
		{name: "chat", capType: "chat", wantChat: true},
		{name: "completion with non-empty prompt", capType: "completion", completionPrompt: "hello", wantCompletion: true},
		{name: "completion with empty prompt", capType: "completion", completionPrompt: "", wantCompletion: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload map[string]interface{}
			switch tt.capType {
			case "chat":
				cc := NewChatCapability("n", "id", "u", "u", now, now, "o", "sys")
				b, _ := json.Marshal(cc)
				_ = json.Unmarshal(b, &payload)
			case "completion":
				cc := NewCompletionCapability("n", "id", "u", "u", now, now, "o", "sys", tt.completionPrompt, "text")
				b, _ := json.Marshal(cc)
				_ = json.Unmarshal(b, &payload)
			}
			raw, _ := json.Marshal(payload)

			var resp ResponseCreateCapabilityV1CapabilitiesPost
			if err := json.Unmarshal(raw, &resp); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if (resp.ChatCapability != nil) != tt.wantChat {
				t.Errorf("ChatCapability populated=%v, want=%v", resp.ChatCapability != nil, tt.wantChat)
			}
			if (resp.CompletionCapability != nil) != tt.wantCompletion {
				t.Errorf("CompletionCapability populated=%v, want=%v", resp.CompletionCapability != nil, tt.wantCompletion)
			}
			if (resp.ExtractionCapability != nil) != tt.wantExtraction {
				t.Errorf("ExtractionCapability populated=%v, want=%v", resp.ExtractionCapability != nil, tt.wantExtraction)
			}
			if (resp.SpeechToTextCapability != nil) != tt.wantSpeechToText {
				t.Errorf("SpeechToTextCapability populated=%v, want=%v", resp.SpeechToTextCapability != nil, tt.wantSpeechToText)
			}

			if tt.capType == "completion" && resp.CompletionCapability != nil {
				if resp.CompletionCapability.CompletionPrompt != tt.completionPrompt {
					t.Errorf("CompletionPrompt=%q, want=%q", resp.CompletionCapability.CompletionPrompt, tt.completionPrompt)
				}
			}
		})
	}
}
