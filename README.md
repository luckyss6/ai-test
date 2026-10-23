# ai-test · A 股数据看板

一个用 **Go** 编写的个人项目，用于**定时抓取 A 股行情数据并记录**，并提供**前端页面展示**，通过 **GitHub Pages** 发布。

> 仅供学习研究使用，不构成投资建议。

## 功能特性

- ✅ Go 实现，无第三方运行时依赖（标准库完成抓取、存储、调度、静态服务）
- ✅ 定时抓取 A 股日 K 线数据并记录为 JSON 文件（留档）
- ✅ 静态前端页面展示行情表格与收盘走势图（零依赖 Canvas 绘图）
- ✅ GitHub Actions 定时抓取 + 自动提交数据 + 自动发布到 GitHub Pages
- ✅ 打 `v*` 标签自动构建多平台二进制并发布 Release
- ✅ Issue 模板用于 bug 反馈，形成“bug → issue → 修复 → 发版”的迭代闭环

## 架构

```
腾讯行情公开接口 (ifzq.gtimg.cn)
        │  HTTP
        ▼
┌────────────────┐    写入     ┌──────────────────────────┐
│  fetcher(Go)   │ ──────────▶ │ web/data/latest.json      │
└────────────────┘            │ web/data/history/<code>.json│
        │                      └──────────────────────────┘
        │ GitHub Actions 定时            │
        ▼                                ▼
   提交到仓库                        GitHub Pages
                                  (web/ 静态站点)
```

## 目录结构

```
ai-test/
├── cmd/astock/            # 命令行入口(fetch / serve / schedule)
├── internal/
│   ├── config/            # 配置读取
│   ├── fetcher/           # 行情抓取
│   └── store/             # JSON 落盘
├── web/                   # 前端站点(Pages 根目录)
│   ├── index.html
│   ├── app.js
│   ├── style.css
│   └── data/              # 抓取生成的行情数据(提交留档)
├── .github/
│   ├── workflows/         # 定时抓取 + Pages 发布 + 发版
│   └── ISSUE_TEMPLATE/    # bug 报告模板
├── config.json            # 标的列表与数据目录配置
├── go.mod
└── Makefile
```

## 快速开始（本地）

```bash
# 1. 抓取一次数据
go run ./cmd/astock fetch

# 2. 本地预览前端页面
go run ./cmd/astock serve
# 打开 http://localhost:8080

# 3. 本地定时抓取（默认每小时一次）
go run ./cmd/astock schedule -every 1h
```

## 数据源

使用[腾讯证券](https://gu.qq.com/)公开的历史行情接口（不复权日 K 线）：

```
https://web.ifzq.gtimg.cn/appstock/app/kline/kline?param=sh600519,day,,,2000
```

`secid` 中 `1.` 表示上海（sh）、`0.` 表示深圳（sz），程序会自动推导出 `sh600519` 这类行情代码。

## 配置

编辑根目录 `config.json` 即可增删标的：

```json
{
  "data_dir": "web/data",
  "symbols": [
    { "code": "600519", "name": "贵州茅台", "secid": "1.600519" },
    { "code": "000858", "name": "五粮液",   "secid": "0.000858" }
  ]
}
```

## 定时抓取与 Pages 发布

GitHub Actions 工作流 `.github/workflows/fetch.yml` 会在**每个交易日 15:15（北京时间）**自动：

1. 抓取最新日 K 线数据
2. 将 `web/data/` 的更新提交回仓库
3. 将 `web/` 部署到 GitHub Pages

也可在仓库 Actions 页面手动点击 **fetch-and-publish → Run workflow** 立即触发。

## 迭代流程（bug → issue → 修复 → 发版）

1. **发现 bug**：在仓库 `Issues` 中使用「Bug 报告」模板提交，附上复现步骤与版本号。
2. **修复**：基于 issue 开分支修复并合并，提交信息关联 `#<issue编号>`。
3. **发版**：合并后打版本标签并推送，工作流自动构建多平台二进制并创建 Release：

```bash
git tag v0.1.0
git push origin v0.1.0
```

详见 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 免责声明

本项目仅用于个人学习与技术研究，行情数据来自公开接口，可能存在延迟或缺失；不构成任何投资建议。
