# Starcat Alfred Workflow

<!-- starcat-promo:start -->
<div align="center">
<a href="https://starcat.ink"><img src="https://raw.githubusercontent.com/starcat-app/starcat-pro/main/banner.webp" width="100%" alt="Starcat" /></a>

<p><strong>这是在 Alfred 中搜索 Starcat 本地仓库与 GitHub 的官方 Workflow。</strong></p>
<p>Starcat 是一款原生 macOS 应用，可以把 GitHub Stars 变成可搜索、可整理、可用 AI 追问的本地知识库，并通过桌面客户端、插件、CLI 与可自部署服务组成完整生态。</p>

<a href="https://github.com/starcat-app/homebrew-starcat"><img src="https://img.shields.io/badge/Install%20with-Homebrew-FBBF24?style=for-the-badge&logo=homebrew&logoColor=white" width="220" alt="Install with Homebrew"/></a>
<br/>
<sub><a href="./README.md">English</a></sub>
</div>

<div align="center">
<a href="https://starcat.ink"><img src="https://img.shields.io/badge/website-starcat.ink-38BDF8?style=flat&color=blue" alt="website"/></a>
<a href="https://github.com/starcat-app/starcat-pro"><img src="https://img.shields.io/badge/support-starcat--pro-lightgrey.svg?style=flat&color=blue" alt="support"/></a>
<a href="https://github.com/starcat-app/homebrew-starcat"><img src="https://img.shields.io/badge/install-homebrew-lightgrey.svg?style=flat&color=blue" alt="homebrew"/></a>
<a href="https://github.com/starcat-app/starcat-localization"><img src="https://img.shields.io/badge/localization-open-lightgrey.svg?style=flat&color=blue" alt="localization"/></a>
</div>

<div align="center">
<img width="900" src="https://raw.githubusercontent.com/starcat-app/starcat-pro/main/main.webp" alt="Starcat main window"/>
</div>

**首选 Homebrew 安装：**

```bash
brew tap starcat-app/starcat
brew trust starcat-app/starcat
brew install --cask starcat
```

**相关链接：**

- 官网与下载: https://starcat.ink
- Mac App Store: 搜索 Starcat for GitHub
- 公开支持与发布说明: https://github.com/starcat-app/starcat-pro
- Starcat App Homebrew tap: https://github.com/starcat-app/homebrew-starcat
- CLI / MCP: [starcat-cli](https://github.com/starcat-app/starcat-cli) / [Homebrew tap](https://github.com/starcat-app/homebrew-starcat-cli)
- AI Agent Skill: https://github.com/starcat-app/starcat-skill
- 浏览器插件: [Chrome](https://github.com/starcat-app/starcat-chrome-plugin) / [Safari](https://github.com/starcat-app/starcat-safari-plugin)
- 启动器集成: [Alfred](https://github.com/starcat-app/starcat-alfred-workflow) / [uTools](https://github.com/starcat-app/starcat-utools-plugin) / [Raycast](https://github.com/starcat-app/starcat-raycast-extension)
- 官方文档: https://github.com/starcat-app/starcat-docs
- 官网源码: https://github.com/starcat-app/starcat-site
- 本地化: https://github.com/starcat-app/starcat-localization

**可自部署支撑 API：**

- [starcat-sharing-api](https://github.com/starcat-app/starcat-sharing-api)
- [starcat-trending-api](https://github.com/starcat-app/starcat-trending-api)
- [starcat-weekly-api](https://github.com/starcat-app/starcat-weekly-api)
- [starcat-wiki-api](https://github.com/starcat-app/starcat-wiki-api)
- [starcat-recommend-api](https://github.com/starcat-app/starcat-recommend-api)
- [starcat-discovery-api](https://github.com/starcat-app/starcat-discovery-api)
<!-- starcat-promo:end -->

[![CI](https://github.com/starcat-app/starcat-alfred-workflow/actions/workflows/ci.yml/badge.svg)](https://github.com/starcat-app/starcat-alfred-workflow/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

在 Alfred 中直接搜索 Starcat 本地仓库和 GitHub。列表展示仓库 owner 头像、
数据来源、语言、Star 数和描述。

## 安装

从 [GitHub 最新 Release](https://github.com/starcat-app/starcat-alfred-workflow/releases/latest)
下载 `Starcat.alfredworkflow`，打开文件并在 Alfred 中确认导入。

Release 同时提供 `checksums.txt` 和 GitHub artifact attestation。Workflow 仍需
Starcat Pro，以及已经完成配对的 `starcat` CLI。

## 使用条件

- macOS 15 或更高版本
- 支持 Workflow 的 Alfred
- Starcat Pro，并已开启 MCP Service
- 已配对且支持顶层 `search` 命令的 `starcat` CLI

## 使用方式

1. 安装 `Starcat.alfredworkflow`。
2. 输入 `starcat` 和仓库关键词。
3. 本地结果按 Return 后在 Starcat 中打开。
4. 仅 GitHub 命中的结果按 Return 后在 GitHub 中打开。

Workflow 不读取 Starcat SQLite、Keychain 或 GitHub 凭据，数据只来自：

```bash
starcat search "$QUERY" --source all --limit 30
```

## 故障排查

如果输入 `starcat <关键词>` 后只看到 Alfred 默认 Web Search，请先确认安装的是
v1.0.0 或更高版本。旧版错误地开启了 Alfred 本地结果过滤，查询词不会传给
Starcat CLI。

如果列表显示 `Starcat TLS certificate fingerprint mismatch`，说明 CLI 保存的
配对资料与当前 MCP Service 不一致。请保持当前 Starcat 实例运行，在「设置 →
MCP」中复制新的配对命令，重新配对后执行：

```bash
starcat doctor
starcat search "starcat" --source all --limit 30
```

## 开发

```bash
go test ./...
./scripts/build.sh
plutil -lint info.plist
```

`scripts/package.sh` 只在本地生成 `dist/Starcat.alfredworkflow`。推送与
`info.plist` 版本一致的 tag 后才会运行 Release Workflow；本地构建脚本不会创建
tag 或发布 Release。

owner 头像缓存于 `alfred_workflow_cache/avatars/v1`。冷缓存先展示高对比度
fallback，后台 helper 下载公开 GitHub 头像；Alfred 最多自动 rerun 三次，
不会因为头像阻塞搜索结果。

[English](./README.md)
