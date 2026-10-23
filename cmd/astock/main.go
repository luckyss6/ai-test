// astock 是 A 股数据定时抓取与展示工具。
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/luckyss6/ai-test/internal/config"
	"github.com/luckyss6/ai-test/internal/fetcher"
	"github.com/luckyss6/ai-test/internal/store"
)

func main() {
	// 子命令作为第一个非 - 开头的参数，之后的参数交给 FlagSet 解析。
	args := os.Args[1:]
	cmd := "fetch"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	}

	fs := flag.NewFlagSet("astock", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "配置文件路径")
	every := fs.Duration("every", time.Hour, "调度周期（仅 schedule 子命令）")
	addr := fs.String("addr", ":8080", "监听地址（仅 serve 子命令）")
	fs.Usage = usage
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	switch cmd {
	case "fetch":
		if err := runFetch(cfg); err != nil {
			log.Fatalf("抓取失败: %v", err)
		}
		log.Println("抓取完成")
	case "serve":
		if err := runServe(cfg, *addr); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	case "schedule":
		if err := runSchedule(cfg, *every); err != nil {
			log.Fatalf("调度失败: %v", err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `astock —— A 股数据定时抓取与展示

用法:
  astock fetch            抓取一次数据并写入 web/data
  astock serve -addr :8080  启动本地静态服务器（默认 :8080）
  astock schedule -every 1h 按周期定时抓取（默认每小时）

参数:
  -config string   配置文件路径（默认 config.json）
  -every duration  调度周期（仅 schedule，默认 1h）
  -addr string     监听地址（仅 serve，默认 :8080）
`)
}

// runFetch 拉取所有标的的日 K 线并落盘。
func runFetch(cfg *config.Config) error {
	quotes := make([]fetcher.Quote, 0, len(cfg.Symbols))
	for i := range cfg.Symbols {
		sym := cfg.Symbols[i]
		bars, err := fetcher.FetchHistory(sym.SecID)
		if err != nil {
			log.Printf("抓取 %s(%s) 失败: %v", sym.Name, sym.Code, err)
			continue
		}

		last, prev, ok := fetcher.LatestBar(bars)
		if !ok {
			log.Printf("%s(%s) 无数据", sym.Name, sym.Code)
			continue
		}
		change := last.Close - prev
		changePct := 0.0
		if prev != 0 {
			changePct = change / prev * 100
		}
		quotes = append(quotes, fetcher.Quote{
			Code:      sym.Code,
			Name:      sym.Name,
			SecID:     sym.SecID,
			Date:      last.Date,
			Open:      last.Open,
			Close:     last.Close,
			High:      last.High,
			Low:       last.Low,
			Volume:    last.Volume,
			PrevClose: prev,
			Change:    round2(change),
			ChangePct: round2(changePct),
		})

		if err := store.WriteHistory(cfg.DataDir, sym.Code, bars); err != nil {
			return err
		}
		if i < len(cfg.Symbols)-1 {
			time.Sleep(300 * time.Millisecond) // 轻微限速
		}
	}

	if len(quotes) == 0 {
		return fmt.Errorf("没有任何标的抓取成功")
	}
	return store.WriteLatest(cfg.DataDir, quotes)
}

// runServe 启动静态文件服务器，根目录为 data_dir 的上一级（即 web/）。
func runServe(cfg *config.Config, addr string) error {
	root := webRoot(cfg)
	log.Printf("静态服务已启动: http://localhost%s（根目录 %s）", addr, root)
	return http.ListenAndServe(addr, http.FileServer(http.Dir(root)))
}

// runSchedule 先立即抓取一次，随后按固定周期循环抓取。
func runSchedule(cfg *config.Config, every time.Duration) error {
	if every <= 0 {
		return fmt.Errorf("调度周期必须大于 0")
	}
	log.Printf("调度器启动：每 %s 抓取一次", every)
	if err := runFetch(cfg); err != nil {
		log.Printf("首次抓取失败: %v", err)
	}
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for range ticker.C {
		if err := runFetch(cfg); err != nil {
			log.Printf("抓取失败: %v", err)
		}
	}
	return nil
}

func webRoot(cfg *config.Config) string {
	dir := cfg.DataDir
	if dir == "" {
		dir = "web/data"
	}
	return filepath.Dir(filepath.Clean(dir))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
