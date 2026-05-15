package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListDepartmentsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	resp, err := c.ListDepartments(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestListDepartmentsWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListDepartments(context.Background(), &ListOptions{Page: 3, Limit: 10})
	assertQueryContains(t, capturedURL, map[string]string{"page": "3", "limit": "10"}, nil)
}

func TestListDepartmentsBuildsCorrectPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListDepartments(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/departments/")
}

func TestGetDepartmentRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetDepartment(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetDepartmentBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetDepartment(context.Background(), "097")
	assertPathContains(t, capturedURL, "/api/departments/097/")
}
