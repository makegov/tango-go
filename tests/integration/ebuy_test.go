//go:build integration

package integration

import (
	"testing"
)

func TestLiveEbuyAccess(t *testing.T) {
	c := newLiveClient(t)
	access, err := c.GetEbuyAccess(ctx(t))
	if err != nil {
		t.Fatalf("GetEbuyAccess: %v", err)
	}
	if access.Enabled != (access.Reason == nil) {
		t.Errorf("Enabled=%v disagrees with Reason=%v", access.Enabled, access.Reason)
	}
	if access.Enabled && len(access.Contracts) == 0 {
		t.Error("expected at least one contract when access is enabled")
	}
}
