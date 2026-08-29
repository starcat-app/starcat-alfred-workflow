# AGENTS.md — Starcat Alfred Workflow

本文档是本仓库 AI 协作规则的唯一维护源。

## 独立仓库边界

- 本目录是 `starcat-app/starcat-alfred-workflow` 独立 Git 仓库，拥有自己的
  Workflow 版本、CI 与 GitHub Release，不属于 Starcat App 或 CLI 仓库。
- 修改前必须确认当前分支与工作区状态；未经 dong4j 明确要求，不得切换分支、提交或处理其他仓库。
- 本仓只负责 Alfred 展示和交互适配；搜索能力与 contract 由 `starcat-cli` 提供。

## 用途与技术栈

本项目是 Alfred 的 Starcat 仓库搜索薄适配器。Universal Go helper 调用已配对的
`starcat search`，将 CLI JSON contract 转换为 Alfred Script Filter JSON；
本地结果通过受限 Starcat deep link 打开，GitHub-only 结果在浏览器打开。

- Go module：`github.com/starcat-app/starcat-alfred-workflow`
- Go directive 1.25.0，toolchain Go 1.26.5
- Alfred Workflow `info.plist` / Script Filter
- macOS Universal binary：arm64 + x86_64

## 关键目录

- `cmd/starcat-alfred/`：helper 入口与集成测试。
- `internal/starcatcli/`：CLI 定位、参数调用与 contract 解码。
- `internal/alfredjson/`：Alfred JSON 模型和渲染。
- `internal/avatarcache/`：头像缓存、回退图和异步补全。
- `info.plist`：Workflow 图、keyword、变量、版本和 helper 调用。
- `scripts/build.sh`：构建 Universal helper；`scripts/package.sh`：生成本地 Workflow 包。

## 开发与验证

```bash
go mod verify
go test ./...
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
bash -n scripts/*.sh
plutil -lint info.plist
./scripts/build.sh
lipo bin/starcat-alfred -verify_arch arm64 x86_64
git diff --check
```

## 项目特有约束

- 保持薄适配器：不得读取 Starcat SQLite、Keychain、GitHub token 或 pairing profile，
  不得在本仓实现搜索、排序、去重、鉴权或 Pro entitlement。
- 查询必须继续交给 `starcat search "$QUERY" --source all --limit ...`，并遵守
  `starcat-cli/contracts/global-search` v1；contract 变化必须同步 fixture 和测试。
- `info.plist` 的 Alfred 本地结果过滤必须保持关闭，查询变化应传给 CLI，
  否则 Alfred 会用空查询结果自行过滤并隐藏真实提示或错误。
- CLI 路径参数和查询必须作为独立 argv 传递，禁止拼接 shell 命令。
- 只允许打开 contract 中通过 allowlist 校验的 Starcat/GitHub URL。
- `info.plist` 版本必须与发布 tag 一致；头像缓存位于
  `alfred_workflow_cache/avatars/v1`，不得写入仓库或读取私有凭据。

## 发布边界与禁令

`scripts/package.sh` 只用于生成 `dist/Starcat.alfredworkflow`；`v*` tag 会触发
GitHub Release、checksum 与 attestation。未经 dong4j 在当前任务中明确授权，
禁止执行打包脚本、创建或推送 tag、执行 `git push`、发布 Release、上传
`.alfredworkflow`、触发发布 workflow 或执行任何对外分发操作。
