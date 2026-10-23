// Package fetcher 负责从公开行情接口拉取 A 股数据。
package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Bar 表示一条日 K 线。
type Bar struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"` // 成交量（手）
}

// Quote 是写入 latest.json、供前端展示的一条最新行情。
type Quote struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	SecID     string  `json:"secid"`
	Date      string  `json:"date"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Volume    float64 `json:"volume"`
	PrevClose float64 `json:"prev_close"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_pct"`
}

const klineURL = "https://web.ifzq.gtimg.cn/appstock/app/kline/kline"

// client 复用连接，并设置超时。
var client = &http.Client{Timeout: 20 * time.Second}

// tencentSymbol 把“市场.代码”标识转成腾讯的 sh/sz 前缀代码。
func tencentSymbol(secid string) string {
	parts := strings.Split(secid, ".")
	if len(parts) != 2 {
		return ""
	}
	if parts[0] == "1" {
		return "sh" + parts[1]
	}
	return "sz" + parts[1]
}

type tencentResp struct {
	Code int `json:"code"`
	Data map[string]struct {
		// 部分历史行会附带除权事件对象作为第 7 个字段，因此用 interface{} 兜底。
		Day [][]interface{} `json:"day"`
	} `json:"data"`
}

// FetchHistory 拉取指定标的的日 K 线历史（不复权），返回按日期升序的 K 线。
func FetchHistory(secid string) ([]Bar, error) {
	sym := tencentSymbol(secid)
	if sym == "" {
		return nil, fmt.Errorf("invalid secid %s", secid)
	}
	// 抓取最近约 2000 个交易日（含当日）。注意不要传 end 日期，否则会漏掉当日 K 线。
	u := fmt.Sprintf("%s?param=%s,day,,,2000", klineURL, sym)

	var body []byte
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		body, err = httpGet(u)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	var resp tencentResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("tencent code %d", resp.Code)
	}
	data, ok := resp.Data[sym]
	if !ok || len(data.Day) == 0 {
		return nil, fmt.Errorf("empty data for %s", sym)
	}

	bars := make([]Bar, 0, len(data.Day))
	for _, row := range data.Day {
		if b, ok := parseKline(row); ok {
			bars = append(bars, b)
		}
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("no valid bars for %s", sym)
	}
	return bars, nil
}

// parseKline 解析 ["date","open","close","high","low","volume", ...] 形式的一行。
// 只取前 6 个字段，忽略可能存在的除权事件等附加字段。
func parseKline(row []interface{}) (Bar, bool) {
	if len(row) < 6 {
		return Bar{}, false
	}
	str := func(v interface{}) string {
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
	f := func(v interface{}) float64 {
		val, _ := strconv.ParseFloat(str(v), 64)
		return val
	}
	return Bar{
		Date:   str(row[0]),
		Open:   f(row[1]),
		Close:  f(row[2]),
		High:   f(row[3]),
		Low:    f(row[4]),
		Volume: f(row[5]),
	}, true
}

// LatestBar 返回最后一条 K 线以及用于计算涨跌的前收价。
func LatestBar(bars []Bar) (Bar, float64, bool) {
	if len(bars) == 0 {
		return Bar{}, 0, false
	}
	last := bars[len(bars)-1]
	prev := 0.0
	if len(bars) >= 2 {
		prev = bars[len(bars)-2].Close
	} else {
		prev = last.Close
	}
	return last, prev, true
}

// httpGet 发起一次 GET 请求并返回响应体。
func httpGet(u string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://gu.qq.com/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
