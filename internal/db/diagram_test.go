package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"online_mermaid/internal/db"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "mermaid-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	cleanup := func() {
		database.Close()
		os.RemoveAll(dir)
	}
	return database, cleanup
}

func TestEnsureDefaultDiagram(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	if err := database.EnsureDefaultDiagram(ctx); err != nil {
		t.Fatalf("EnsureDefaultDiagram failed: %v", err)
	}

	items, err := database.ListDiagrams(ctx)
	if err != nil {
		t.Fatalf("ListDiagrams failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "graph1" {
		t.Fatalf("expected title graph1, got %s", items[0].Title)
	}
}

func TestAutoIncrementNaming(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_ = database.EnsureDefaultDiagram(ctx) // graph1

	next, err := database.GetNextDefaultTitle(ctx)
	if err != nil {
		t.Fatalf("GetNextDefaultTitle failed: %v", err)
	}
	if next != "graph2" {
		t.Fatalf("expected graph2, got %s", next)
	}

	// Create graph2
	d2, err := database.CreateDiagram(ctx, "", "flowchart TD\n A-->B")
	if err != nil {
		t.Fatalf("CreateDiagram failed: %v", err)
	}
	if d2.Title != "graph2" {
		t.Fatalf("expected d2 title graph2, got %s", d2.Title)
	}

	// Next should be graph3
	next3, _ := database.GetNextDefaultTitle(ctx)
	if next3 != "graph3" {
		t.Fatalf("expected graph3, got %s", next3)
	}
}

func TestDiagramCRUD(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	d, err := database.CreateDiagram(ctx, "架构图", "flowchart LR\n A-->B")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Get
	fetched, err := database.GetDiagram(ctx, d.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if fetched.Title != "架构图" || fetched.Code != "flowchart LR\n A-->B" {
		t.Fatalf("unexpected fetched diagram: %+v", fetched)
	}

	// Update
	updated, err := database.UpdateDiagram(ctx, d.ID, "修改架构图", "flowchart LR\n B-->C")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != "修改架构图" || updated.Code != "flowchart LR\n B-->C" {
		t.Fatalf("unexpected updated: %+v", updated)
	}

	// Duplicate
	dup, err := database.DuplicateDiagram(ctx, d.ID)
	if err != nil {
		t.Fatalf("duplicate failed: %v", err)
	}
	if dup.Title != "修改架构图 (副本)" || dup.Code != "flowchart LR\n B-->C" {
		t.Fatalf("unexpected duplicate: %+v", dup)
	}

	// Delete
	if err := database.DeleteDiagram(ctx, d.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = database.GetDiagram(ctx, d.ID)
	if err != db.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateDiagramPreservesCodeWhenEmpty(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	d, err := database.CreateDiagram(ctx, "架构图", "flowchart LR\n A-->B")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Rename-only update (empty code must not wipe existing code)
	updated, err := database.UpdateDiagram(ctx, d.ID, "重命名", "")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != "重命名" {
		t.Fatalf("expected title 重命名, got %s", updated.Title)
	}
	if updated.Code != "flowchart LR\n A-->B" {
		t.Fatalf("code was wiped by rename-only update: %q", updated.Code)
	}

	// Full update still replaces code
	updated, err = database.UpdateDiagram(ctx, d.ID, "重命名", "flowchart TD\n X-->Y")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Code != "flowchart TD\n X-->Y" {
		t.Fatalf("expected new code, got %q", updated.Code)
	}
}
