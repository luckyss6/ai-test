# 贡献与迭代流程

本项目遵循“发现 bug → 提交 issue → 修复 → 发布版本”的迭代闭环。

## 1. 发现 bug

在仓库 `Issues` 中使用 **Bug 报告** 模板提交，尽量包含：

- 清晰的描述
- 复现步骤
- 预期行为 vs 实际行为
- 版本号与运行环境
- 相关日志或截图

## 2. 修复

1. 从 `main` 新建分支，例如 `fix/issue-12`。
2. 修改代码，本地验证：

   ```bash
   go vet ./...
   go build ./...
   go run ./cmd/astock fetch
   ```

3. 提交时在信息中引用 issue，例如 `fix: 修复 K 线解析 (#12)`。
4. 推送并创建 Pull Request，合并到 `main`。

## 3. 发布版本

每次迭代完成后，打语义化版本标签并推送，`release.yml` 会自动构建
Linux / Windows / macOS 二进制并创建 GitHub Release：

```bash
git tag v0.2.0
git push origin v0.2.0
```

## 4. 提交规范

- `feat:` 新功能
- `fix:` 修复 bug（关联 issue）
- `data:` 数据更新（由工作流自动提交）
- `docs:` 文档
- `chore:` 杂项
