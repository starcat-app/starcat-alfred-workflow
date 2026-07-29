# Contributing

Thank you for improving the Starcat Alfred Workflow.

## Before opening a pull request

1. Open an issue for user-visible behavior or Starcat CLI contract changes.
2. Keep search, ranking, permissions, and entitlement logic in Starcat. This repository is only an Alfred adapter.
3. Never read Starcat SQLite, Keychain, GitHub tokens, or pairing credentials directly.
4. Preserve argv-based CLI invocation and the allowlisted repository URL policy.
5. Add tests for changed behavior.

## Local checks

```bash
go mod verify
go test ./...
go test -race ./...
go vet ./...
bash -n scripts/*.sh
plutil -lint info.plist
./scripts/build.sh
lipo bin/starcat-alfred -verify_arch arm64 x86_64
```

Do not include generated `bin/`, `dist/`, or `.alfredworkflow` artifacts in commits.

## Commit and pull request scope

Keep changes focused. A pull request should describe the Alfred-visible outcome, security or compatibility implications, tests run, and any minimum Starcat or CLI version requirement.
