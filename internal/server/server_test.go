package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"online_mermaid/internal/auth"
	"online_mermaid/internal/config"
	"online_mermaid/internal/db"
	"online_mermaid/internal/server"
	"online_mermaid/internal/webui"
)

func setupTestServer(t *testing.T) (*httptest.Server, *db.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "mermaid-server-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_ = database.EnsureDefaultDiagram(context.Background())

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	sessions := auth.NewStore(time.Hour, false)
	authHandler := auth.NewHandler("admin", hash, sessions, auth.NewLoginLimiter(10, time.Minute))

	assets, _ := webui.Dist()
	srv := server.New(server.Config{
		AppConfig: config.Config{
			Addr:         "127.0.0.1:0",
			DBPath:       dbPath,
			Username:     "admin",
			PasswordHash: hash,
			SessionTTL:   time.Hour,
		},
		Auth:     authHandler,
		Sessions: sessions,
		DB:       database,
		Assets:   assets,
	})

	ts := httptest.NewServer(srv.Handler())
	cleanup := func() {
		ts.Close()
		database.Close()
		os.RemoveAll(dir)
	}
	return ts, database, cleanup
}

func TestAuthAndDiagramAPIs(t *testing.T) {
	ts, _, cleanup := setupTestServer(t)
	defer cleanup()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := ts.Client()
	client.Jar = jar

	// 1. Initial session check (not logged in)
	resp, err := client.Get(ts.URL + "/api/auth/session")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	var sess struct {
		Authenticated bool   `json:"authenticated"`
		Username      string `json:"username"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&sess)
	resp.Body.Close()
	if sess.Authenticated {
		t.Fatalf("expected unauthenticated, got true")
	}

	// 2. Unauthenticated request to /api/diagrams should return 401
	resp, err = client.Get(ts.URL + "/api/diagrams")
	if err != nil {
		t.Fatalf("get diagrams: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	// 3. Login with wrong password
	loginPayload, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/auth/login", bytes.NewReader(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("login wrong: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong password, got %d", resp.StatusCode)
	}

	// 4. Login with correct password
	loginPayload, _ = json.Marshal(map[string]string{"username": "admin", "password": "secret123"})
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/api/auth/login", bytes.NewReader(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("login ok: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 5. Session check (authenticated)
	resp, _ = client.Get(ts.URL + "/api/auth/session")
	_ = json.NewDecoder(resp.Body).Decode(&sess)
	resp.Body.Close()
	if !sess.Authenticated || sess.Username != "admin" {
		t.Fatalf("expected authenticated as admin, got %+v", sess)
	}

	// 6. List diagrams
	resp, err = client.Get(ts.URL + "/api/diagrams")
	if err != nil {
		t.Fatalf("list diagrams: %v", err)
	}
	var listResp struct {
		Items []db.DiagramMeta `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()
	if len(listResp.Items) != 1 || listResp.Items[0].Title != "graph1" {
		t.Fatalf("unexpected list response: %+v", listResp)
	}

	// 7. Create new diagram
	createPayload, _ := json.Marshal(map[string]string{"title": "graph2", "code": "flowchart TD\n A-->B"})
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/api/diagrams", bytes.NewReader(createPayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("create diagram: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d", resp.StatusCode)
	}
	var created db.Diagram
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	if created.Title != "graph2" {
		t.Fatalf("expected title graph2, got %s", created.Title)
	}

	// 8. Update diagram
	updatePayload, _ := json.Marshal(map[string]string{"title": "自定义名称", "code": "flowchart LR\n X-->Y"})
	req, _ = http.NewRequest(http.MethodPut, ts.URL+"/api/diagrams/"+created.ID, bytes.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("update diagram: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 ok, got %d", resp.StatusCode)
	}
	var updated db.Diagram
	_ = json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()
	if updated.Title != "自定义名称" || updated.Code != "flowchart LR\n X-->Y" {
		t.Fatalf("unexpected updated: %+v", updated)
	}

	// 9. Delete diagram
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/api/diagrams/"+created.ID, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete diagram: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 ok on delete, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
