// Package config 负责读取项目配置。
package config

import (
	"encoding/json"
	"os"
)

// Symbol 表示一个待跟踪的证券标的。
type Symbol struct {
	Code  string `json:"code"`  // 交易代码，例如 600519
	Name  string `json:"name"`  // 名称，例如 贵州茅台
	SecID string `json:"secid"` // 市场+代码标识，例如 1.600519（1=沪，0=深）
}

// Config 是项目根目录下 config.json 的结构。
type Config struct {
	DataDir string   `json:"data_dir"` // 数据输出目录，相对仓库根目录
	Symbols []Symbol `json:"symbols"`  // 跟踪的标的列表
}

// Default 返回内置默认配置，保证缺少 config.json 时程序也能运行。
func Default() *Config {
	return &Config{
		DataDir: "web/data",
		Symbols: []Symbol{
			{Code: "000001", Name: "上证指数", SecID: "1.000001"},
			{Code: "399001", Name: "深证成指", SecID: "0.399001"},
			{Code: "000300", Name: "沪深300", SecID: "1.000300"},
			{Code: "399006", Name: "创业板指", SecID: "0.399006"},
			{Code: "600519", Name: "贵州茅台", SecID: "1.600519"},
			{Code: "601318", Name: "中国平安", SecID: "1.601318"},
			{Code: "600036", Name: "招商银行", SecID: "1.600036"},
			{Code: "000858", Name: "五粮液", SecID: "0.000858"},
			{Code: "300750", Name: "宁德时代", SecID: "0.300750"},
			{Code: "002594", Name: "比亚迪", SecID: "0.002594"},
		},
	}
}

// Load 读取配置文件；文件不存在时返回默认配置。
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.DataDir == "" {
		c.DataDir = "web/data"
	}
	if len(c.Symbols) == 0 {
		c.Symbols = Default().Symbols
	}
	return &c, nil
}
