// Package store 负责把抓取到的数据落盘为 JSON 文件。
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/luckyss6/ai-test/internal/fetcher"
)

// LatestFile 是前端读取的 web/data/latest.json 结构。
type LatestFile struct {
	GeneratedAt string          `json:"generated_at"`
	Symbols     []fetcher.Quote `json:"symbols"`
}

// WriteLatest 写入最新行情快照。
func WriteLatest(dir string, quotes []fetcher.Quote) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out := LatestFile{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Symbols:     quotes,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "latest.json"), b, 0o644)
}

// WriteHistory 写入某个标的的完整日 K 线历史（按日期升序覆盖写）。
func WriteHistory(dir, code string, bars []fetcher.Bar) error {
	if err := os.MkdirAll(filepath.Join(dir, "history"), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(bars, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "history", code+".json"), b, 0o644)
}
