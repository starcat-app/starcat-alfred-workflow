# Starcat Alfred Workflow

<!-- starcat-promo:start -->
<div align="center">
<a href="https://starcat.ink"><img src="https://raw.githubusercontent.com/starcat-app/starcat-pro/main/banner.webp" width="100%" alt="Starcat" /></a>

<p><strong>Official Alfred Workflow for searching Starcat local repositories and GitHub.</strong></p>
<p>Starcat is a native macOS app that turns GitHub Stars into a searchable, organized and AI-assisted knowledge base. It supports README rendering, tags, private notes, release tracking, repository health signals, AI summaries, semantic search, browser plugin workflows and self-hostable support APIs.</p>

<a href="https://github.com/starcat-app/homebrew-starcat"><img src="https://img.shields.io/badge/Install%20with-Homebrew-FBBF24?style=for-the-badge&logo=homebrew&logoColor=white" width="220" alt="Install with Homebrew"/></a>
<br/>
<sub><a href="./README-ZH.md">中文说明</a></sub>
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

**Preferred install method:**

```bash
brew tap starcat-app/starcat
brew trust starcat-app/starcat
brew install --cask starcat
```

**Useful links:**

- Home and downloads: https://starcat.ink
- Public support and release notes: https://github.com/starcat-app/starcat-pro
- Starcat App Homebrew tap: https://github.com/starcat-app/homebrew-starcat
- CLI / MCP: [starcat-cli](https://github.com/starcat-app/starcat-cli) / [Homebrew tap](https://github.com/starcat-app/homebrew-starcat-cli)
- AI Agent Skill: https://github.com/starcat-app/starcat-skill
- Browser plugins: [Chrome](https://github.com/starcat-app/starcat-chrome-plugin) / [Safari](https://github.com/starcat-app/starcat-safari-plugin)
- Documentation: https://github.com/starcat-app/starcat-docs
- Website source: https://github.com/starcat-app/starcat-site
- Localization: https://github.com/starcat-app/starcat-localization

**Self-hostable support APIs:**

- [starcat-sharing-api](https://github.com/starcat-app/starcat-sharing-api)
- [starcat-trending-api](https://github.com/starcat-app/starcat-trending-api)
- [starcat-weekly-api](https://github.com/starcat-app/starcat-weekly-api)
- [starcat-wiki-api](https://github.com/starcat-app/starcat-wiki-api)
- [starcat-recommend-api](https://github.com/starcat-app/starcat-recommend-api)
- [starcat-discovery-api](https://github.com/starcat-app/starcat-discovery-api)
<!-- starcat-promo:end -->

[![CI](https://github.com/starcat-app/starcat-alfred-workflow/actions/workflows/ci.yml/badge.svg)](https://github.com/starcat-app/starcat-alfred-workflow/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

Search Starcat local repositories and GitHub directly from Alfred. Results show
the repository owner avatar, source, language, stars and description.

## Install

Download `Starcat.alfredworkflow` from the [latest GitHub Release](https://github.com/starcat-app/starcat-alfred-workflow/releases/latest), open it, and confirm the import in Alfred.

Release assets include `checksums.txt` and a GitHub artifact attestation. The Workflow still requires Starcat Pro and a paired `starcat` CLI.

## Requirements

- macOS 15 or later
- Alfred with Workflow support
- Starcat Pro with MCP Service enabled
- A paired `starcat` CLI that supports the top-level `search` command

## Usage

1. Install `Starcat.alfredworkflow`.
2. Type `starcat` followed by a query.
3. Press Return on a local result to open it in Starcat.
4. Press Return on a GitHub-only result to open it on GitHub.

The Workflow never reads Starcat SQLite, Keychain or GitHub credentials. Its
only data source is:

```bash
starcat search "$QUERY" --source all --limit 30
```

## Troubleshooting

If `starcat <query>` only shows Alfred's default Web Search results, install
v1.0.0 or newer. Older versions incorrectly enabled Alfred-side result filtering,
so the query was not forwarded to Starcat CLI.

If the list reports `Starcat TLS certificate fingerprint mismatch`, the CLI's
saved pairing no longer matches the active MCP Service. Keep the intended Starcat
instance running, copy a fresh pairing command from Settings → MCP, pair again,
then run:

```bash
starcat doctor
starcat search "starcat" --source all --limit 30
```

## Development

```bash
go test ./...
./scripts/build.sh
plutil -lint info.plist
```

`scripts/package.sh` creates `dist/Starcat.alfredworkflow` for local validation.
Pushing a version tag that matches `info.plist` runs the release workflow; local
build scripts never create tags or publish releases.

Owner avatars are cached under `alfred_workflow_cache/avatars/v1`. Cold-cache
results use a generated high-contrast fallback, then Alfred reruns up to three
times while a detached helper hydrates public GitHub avatars.

[中文说明](./README-ZH.md)
