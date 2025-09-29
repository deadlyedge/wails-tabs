# photoTidyGo Development Plan

## Guiding Principles
- Serve a single local user; avoid multi-user sync, cloud services, or background daemons.
- Prefer configuration files (`settings.toml`) over complex UI forms; keep UI limited to monitoring and confirmations.
- Reuse existing OS capabilities (file copy/move, recycle bin) instead of rebuilding them in Go.
- Commit to a "one feature in, one feature justified" rule to control scope creep.

## MVP Scope (Iteration 0)
1. Basic settings loader: read `settings.toml` for source/destination paths and scan options.
2. Media scanner: walk configured directories, record file path, size, hash (MD5) , EXIF info into SQLite.
3. Duplicate report: list potential duplicates grouped by hash in the UI; no automatic actions yet.
4. Manual tidy flow: execute move files to target structured folder, with processing record in database, for crush resume, show progress bar in ui.
5. Wails desktop shell: minimal React view for selecting folders, showing scan progress, showing settings, recent actions, and duplicate list, and excution progress.

## Iteration 1: Quality & Safety
- Add structured logging/tracing for scans and file actions.
- Include schema versioning in SQLite and migrations for future changes.
- Introduce dry-run mode for action plans and confirmation dialogs before destructive steps.

## Deferred Backlog (only consider after Iteration 1)
- Advanced hashing (BLAKE3) or perceptual image comparison.
- Automated EXIF extraction and filtering.
- Background scheduler, remote APIs, or multi-device sync.
- Installers/signing for multiple OS targets.

## Operational Checklist
- Update README with prerequisites (Go, Wails, Node) and quick-start steps before each release.
- Keep `docs/setup.md` aligned with actual CLI commands and configuration keys.
- Maintain a short testing note: go unit tests for scanner logic + manual end-to-end checklist.

## Scope Guardrails
- Review new feature ideas against MVP goals during planning; defer anything outside duplicate cleanup.
- If a task takes more than two days or introduces third-party services, re-evaluate necessity.
- Prefer polishing existing flows (error states, UX copy, docs) over adding novel capabilities.
