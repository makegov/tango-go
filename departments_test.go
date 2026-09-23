package tango

import (
	"context"
	"errors"
	"net/http"
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

func TestGetDepartmentDecodesIntegerCode(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"abbreviation":"DOD","code":97,"name":"Department of Defense","parent":null}`))
	})
	got, err := c.GetDepartment(context.Background(), "097")
	if err != nil {
		t.Fatalf("GetDepartment: %v", err)
	}
	assertPathContains(t, capturedURL, "/api/departments/097/")
	if got.Code == nil || *got.Code != 97 {
		t.Fatalf("Code: want 97, got %v", got.Code)
	}
	if got.Name == nil || *got.Name != "Department of Defense" {
		t.Errorf("Name: got %v", got.Name)
	}
	if got.Abbreviation == nil || *got.Abbreviation != "DOD" {
		t.Errorf("Abbreviation: got %v", got.Abbreviation)
	}
	if _, ok := got.Extra["parent"]; !ok {
		t.Errorf("unknown field should land in Extra, got %#v", got.Extra)
	}
}
