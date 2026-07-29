# Security Policy

## Reporting a vulnerability

Please report suspected vulnerabilities through [GitHub Security Advisories](https://github.com/starcat-app/starcat-alfred-workflow/security/advisories/new). Do not publish pairing URIs, device tokens, private Starcat data, local paths, or exploit details in a public issue.

Include the affected Workflow version, macOS and Alfred versions, reproduction steps, and the expected security impact. You should receive an acknowledgement within seven days.

## Supported versions

Security fixes are provided for the latest published stable release. Install the newest `Starcat.alfredworkflow` from GitHub Releases before reporting a problem.

## Security boundaries

- The Workflow obtains repository data only by invoking the paired `starcat` CLI.
- The Workflow never reads Starcat SQLite, Keychain, GitHub tokens, or pairing credentials.
- Only versioned `starcat://repo/...` links and canonical `https://github.com/{owner}/{repo}` links can be opened.
- Avatar downloads are limited to public HTTPS URLs returned by Starcat, with bounded size, timeout, and cache lifetime.
- Release assets include a SHA-256 checksum and GitHub artifact attestation.

Local malware, a compromised operating-system account, a compromised Alfred installation, a compromised GitHub account, and a compromised Starcat or CLI installation are outside this Workflow's protection boundary.
