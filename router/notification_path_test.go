package router

import "testing"

// KVolt: POST .../:id/read matches but returns nil Params (use .../read/:id).
// KVolt: POST .../notifications/read-all conflicts with notification routes (use a separate path).
func TestNotificationRoutesPattern(t *testing.T) {
	r := New()
	r.AddRoute("POST", "/member/notifications/read/:id", "single")
	r.AddRoute("POST", "/member/mark-notifications-read", "mark-all")

	h1, ps1, ok1 := r.Find("POST", "/member/notifications/read/550e8400-e29b-41d4-a716-446655440000")
	if !ok1 || h1 != "single" || ps1.Get("id") != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("read/:id: ok=%v h=%v id=%q", ok1, h1, ps1.Get("id"))
	}

	h2, _, ok2 := r.Find("POST", "/member/mark-notifications-read")
	if !ok2 || h2 != "mark-all" {
		t.Fatalf("mark-all: ok=%v h=%v", ok2, h2)
	}
}
