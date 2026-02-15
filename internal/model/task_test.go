package model

import "testing"

func TestCreateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateTaskRequest
		wantErr string
	}{
		{
			name:    "valid request",
			req:     CreateTaskRequest{Title: "Test", Priority: "high"},
			wantErr: "",
		},
		{
			name:    "empty title",
			req:     CreateTaskRequest{Title: "", Priority: "medium"},
			wantErr: "title is required",
		},
		{
			name:    "invalid priority",
			req:     CreateTaskRequest{Title: "Test", Priority: "urgent"},
			wantErr: "priority must be low, medium, or high",
		},
		{
			name:    "default priority",
			req:     CreateTaskRequest{Title: "Test"},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.Validate()
			if got != tt.wantErr {
				t.Errorf("Validate() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b"}, "a") {
		t.Error("expected true")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Error("expected false")
	}
}
