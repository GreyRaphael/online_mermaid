package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Diagram struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	SortOrder int       `json:"sortOrder"`
}

type DiagramMeta struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	SortOrder int       `json:"sortOrder"`
}

var (
	ErrNotFound = errors.New("diagram not found")
	graphRegex  = regexp.MustCompile(`^graph(\d+)$`)
)

func generateID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "d_" + hex.EncodeToString(b[:])
}

const defaultSampleCode = `flowchart TD
    Start([🚀 开始]) --> Input[输入 Mermaid 源码]
    Input --> Mode{选择视图模式}
    
    Mode -->|👁 预览| Preview[专注全屏画布预览]
    Mode -->|✏️ 编辑| Edit[多功能代码编辑器]
    Mode -->|📑 分屏| Split[左侧编辑 · 右侧实时渲染]
    
    Preview --> Export[导出/复制]
    Split --> Export
    
    subgraph 快捷操作
        Export --> PNG[🖼️ 复制白底 PNG]
        Export --> TransparentPNG[💾 导出透明 PNG]
        Export --> SVG[📐 导出 SVG 矢量图]
        Export --> CopySrc[📋 复制 Mermaid 源码]
    end
    
    Split --> Save[💾 自动/快捷键保存至 SQLite]
    Save --> End([🏁 完成])

    style Start fill:#dceadf,stroke:#356a4b,stroke-width:2px,color:#214f36
    style End fill:#dceadf,stroke:#356a4b,stroke-width:2px,color:#214f36
`

func (db *DB) EnsureDefaultDiagram(ctx context.Context) error {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM diagrams").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		now := time.Now().UTC()
		_, err := db.ExecContext(ctx,
			"INSERT INTO diagrams (id, title, code, created_at, updated_at, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
			generateID(), "graph1", defaultSampleCode, now, now, 0,
		)
		return err
	}
	return nil
}

func (db *DB) ListDiagrams(ctx context.Context) ([]DiagramMeta, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, created_at, updated_at, sort_order FROM diagrams ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DiagramMeta
	for rows.Next() {
		var item DiagramMeta
		if err := rows.Scan(&item.ID, &item.Title, &item.CreatedAt, &item.UpdatedAt, &item.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []DiagramMeta{}
	}
	return list, nil
}

func (db *DB) GetDiagram(ctx context.Context, id string) (*Diagram, error) {
	var d Diagram
	err := db.QueryRowContext(ctx,
		"SELECT id, title, code, created_at, updated_at, sort_order FROM diagrams WHERE id = ?",
		id,
	).Scan(&d.ID, &d.Title, &d.Code, &d.CreatedAt, &d.UpdatedAt, &d.SortOrder)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (db *DB) CreateDiagram(ctx context.Context, title, code string) (*Diagram, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		nextTitle, err := db.GetNextDefaultTitle(ctx)
		if err != nil {
			title = "graph1"
		} else {
			title = nextTitle
		}
	}
	if strings.TrimSpace(code) == "" {
		code = "flowchart TD\n    A[节点 A] --> B[节点 B]\n"
	}

	id := generateID()
	now := time.Now().UTC()

	_, err := db.ExecContext(ctx,
		"INSERT INTO diagrams (id, title, code, created_at, updated_at, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
		id, title, code, now, now, 0,
	)
	if err != nil {
		return nil, err
	}

	return &Diagram{
		ID:        id,
		Title:     title,
		Code:      code,
		CreatedAt: now,
		UpdatedAt: now,
		SortOrder: 0,
	}, nil
}

func (db *DB) UpdateDiagram(ctx context.Context, id, title, code string) (*Diagram, error) {
	d, err := db.GetDiagram(ctx, id)
	if err != nil {
		return nil, err
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle != "" {
		d.Title = trimmedTitle
	}
	if strings.TrimSpace(code) != "" {
		d.Code = code
	}
	d.UpdatedAt = time.Now().UTC()

	res, err := db.ExecContext(ctx,
		"UPDATE diagrams SET title = ?, code = ?, updated_at = ? WHERE id = ?",
		d.Title, d.Code, d.UpdatedAt, id,
	)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return d, nil
}

func (db *DB) DeleteDiagram(ctx context.Context, id string) error {
	res, err := db.ExecContext(ctx, "DELETE FROM diagrams WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *DB) DuplicateDiagram(ctx context.Context, id string) (*Diagram, error) {
	src, err := db.GetDiagram(ctx, id)
	if err != nil {
		return nil, err
	}

	newTitle := src.Title + " (副本)"
	newID := generateID()
	now := time.Now().UTC()

	_, err = db.ExecContext(ctx,
		"INSERT INTO diagrams (id, title, code, created_at, updated_at, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
		newID, newTitle, src.Code, now, now, 0,
	)
	if err != nil {
		return nil, err
	}

	return &Diagram{
		ID:        newID,
		Title:     newTitle,
		Code:      src.Code,
		CreatedAt: now,
		UpdatedAt: now,
		SortOrder: 0,
	}, nil
}

func (db *DB) GetNextDefaultTitle(ctx context.Context) (string, error) {
	rows, err := db.QueryContext(ctx, "SELECT title FROM diagrams WHERE title LIKE 'graph%'")
	if err != nil {
		return "graph1", err
	}
	defer rows.Close()

	maxNum := 0
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err == nil {
			matches := graphRegex.FindStringSubmatch(title)
			if len(matches) == 2 {
				if num, err := strconv.Atoi(matches[1]); err == nil && num > maxNum {
					maxNum = num
				}
			}
		}
	}

	return fmt.Sprintf("graph%d", maxNum+1), nil
}
