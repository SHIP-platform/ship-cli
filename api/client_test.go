package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetApplicationsBuildsScopedRequest(t *testing.T) {
	t.Parallel()
	const projectID = "project/id with space"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/applications" {
			t.Errorf("path = %q, want /api/applications", r.URL.Path)
		}
		if got := r.URL.Query().Get("projectId"); got != projectID {
			t.Errorf("projectId = %q, want %q", got, projectID)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"app-1","name":"Demo","projectId":"project/id with space"}]`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL+"/", "test-token")
	apps, err := client.GetApplications(projectID)
	if err != nil {
		t.Fatalf("GetApplications: %v", err)
	}
	if len(apps) != 1 || apps[0].ID != "app-1" || apps[0].ProjectID != projectID {
		t.Fatalf("applications = %#v, want one scoped application", apps)
	}
}

func TestClient_GetApplicationsRejectsNonOKResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "test-token")
	if _, err := client.GetApplications("missing-project"); err == nil {
		t.Fatal("GetApplications should reject a non-200 response")
	}
}
