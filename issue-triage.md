# go-sqlcmd Issue Triage Report

**Date:** 2026-04-24 (updated: honest reassessment of "low severity" items, orphaned commit audit)
**Total open issues:** 125 (was 134; 9 closed this session)
**Total open PRs:** 24 (was 23; PR #743 created this session)

---

## Summary

| Category | Count | Notes |
|----------|-------|-------|
| Bugs (actionable) | 11 | |
| Security / CVE | 6 | govulncheck found 0 vulns in built binary |
| Enhancements (has open PR) | 10 | All reviewed |
| Enhancements (no PR, actionable) | 20 | |
| Feature requests (large / speculative) | 30 | |
| Stale / housekeeping / questions | 25 | |
| Closeable (fixed, duplicate, or wontfix) | 27 | 9 already closed this session |

---

## Work Completed This Session

### Issues Closed (9)
- #553 -- Already fixed in go-mssqldb v1.9.8
- #676 -- Duplicate of #486
- #601 -- Question, not issue
- #169 -- Stale question (3+ years)
- #159 -- Stale question (3+ years)
- #171 -- Stale question (3+ years)
- #690 -- Will auto-close on PR #684 merge (added `Fixes`)
- #597 -- Will auto-close on PR #684 merge (added `Fixes`)
- #604 -- Fixed by PR #707 (already merged)

### PRs Created (1)
- **PR #743** -- fix: check wget exit code before using downloaded file (branch: `wget-exit-code`)

### Fixes Pushed to Existing PR Branches (5)

**PR #688 (branch: pr-685)** -- squashed into de898f1. Our 4 fix commits folded in:
- `jsonc_patch.go`: comment-preserving `patchJSONCKey` (replaces `stripJSONC` round-trip)
- `vscode.go`: atomic temp-file-then-rename write with permission preservation
- `vscode_test.go`: test isolation via `t.Setenv` (replaces package-level global)
- `tool.go`/`tool_linux.go`/`tool_windows.go`: `Args[0]` fix, `Launch()` for GUI apps

**PR #628 (branch: regional-settings)** -- 1 commit:
- Fixed French/Nordic thousand separator (was `.`, now `\u00a0` for fr/ru/sv/fi/no/nb/nn)
- Extracted `parseUnixLocale()` to shared `regional_unix.go` (eliminated 3x duplication)
- Removed dead `unsafe` import from `regional_windows.go`
- Fixed `strings.Replace` to `strings.ReplaceAll` in darwin's `getMacOSLocale()`

**PR #631 (branch: print-statistics)** -- 1 commit:
- Parameterized `printStatistics()` to accept `io.Writer` for forward compat with #632's `GetStat()`

**PR #632 (branch: perftrace)** -- 1 commit:
- Added merge integration note on `GetStat()` godoc re: updating `printStatistics` call site

**PR #630 (branch: serverlist-command)** -- 1 commit:
- Fixed MSSQLSERVER dual-entry bug (was emitting both `(local)` and `serverName`)
- Fixed README batch script `%%~z` outside FOR loop
- Fixed `TestHelpCommand` and `TestServerlistCommand` requiring DB connection

### Security Audit
- Ran `govulncheck ./...` -- 0 vulnerabilities in compiled binary
- All CVEs in #733/#568/#563 are in transitively-imported modules not reachable from built code
- PR #728 (go 1.25.9 + security fixes) merged Apr 9. PR #729 (Docker Moby migration) closed without merge

---

## 1. Will Auto-Close on PR Merge

Issues linked to open PRs via `Fixes #NNN`. No manual close needed.

| # | Title | Closed by PR | Notes |
|---|-------|--------------|-------|
| 690 | Add validation/warning for unsupported config file extensions | PR #684 | `Fixes` added this session |
| 597 | Add documentation/help about supported config types | PR #684 | `Fixes` added this session |
| 603 | Remove SQL Edge from sqlcmd create options | PR #680 | |
| 605 | sqlcmd does not handle command timeout (-t) properly on Linux | PR #702 | |
| 621 | Per-statement timing with -p flag | PR #631 | |
| 622 | :TIMING ON/OFF command | PR #632 | |
| 520 | EXIT command not parsed in PowerShell -Q | PR #636 | |
| 566 | Bak restore from local http corrupted | PR #743 | PR #588 closed; PR #743 supersedes |
| 584 | Remove ADS, add VSCode + MSSQL | PR #688 | |

**NOTE:** PR #635 had `Fixes #604` but that linkage was WRONG. **FIXED** this session via `gh pr edit`. #604 was "-? and -h not working" -- now **CLOSED** (fixed by merged PR #707). PR #635 fixes bare `-r` defaulting.

## 2. Close Manually

| # | Title | Status |
|---|-------|--------|
| 553 | Crash selecting NULL for GEOGRAPHY/GEOMETRY | **CLOSED** this session |
| 676 | Add Parquet format support | **CLOSED** this session (duplicate of #486) |
| 601 | Is this a replacement of ODBC sqlcmd.exe? | **CLOSED** this session |
| 169 | How to separate output channels | **CLOSED** this session |
| 159 | Establish rules for code layout | **CLOSED** this session |
| 171 | Coding guidelines | **CLOSED** this session |
| 555 | Inconsistent Float Formatting go-sqlcmd vs ODBC | Keep open. PR #706 closed; needs `big.Float` approach per maintainer |

---

## 3. Security / CVE -- Needs Attention

| # | Title | Age | Notes |
|---|-------|-----|-------|
| 733 | Fix CVEs in Version v1.10.0 | 4d | govulncheck shows 0 vulns in built binary. CVEs are in unreachable transitive deps |
| 568 | CVE-2024-41110, CVE-2024-45337, CVE-2023-45288 | 1y | Docker/crypto CVEs. Not reachable. PR #728 merged security fixes |
| 563 | CVE-2024-45338 | 1.3y | golang.org/x/net -- not reachable in built binary |
| 729 | Migrate from deprecated docker/docker to Moby SDK | 9d | CLOSED (not merged). #728 (merged Apr 9) bumped go + fixed security vulns |
| 429 | Use Microsoft Go tools for official builds | 3y | Security hardening, no activity |
| 11 | Add signing to Mac binaries | 5y | Long-standing, no progress |

**Status:** govulncheck clean. The CVEs exist only in transitive dependencies that are not compiled into the binary. PR #728 (merged Apr 9) bumped go + fixed security vulns. PR #729 (Docker Moby migration) was closed without merge.

---

## 4. Bugs -- Actionable

| # | Title | Has PR | Age | Notes |
|---|-------|--------|-----|-------|
| 699 | Incorrectly assumes not running on Windows | No | 2mo | Needs investigation |
| 596 | Japanese column names garbled | No | 9mo | Encoding/codepage issue; related to #111 |
| #566 | Bak restore from local http corrupted | PR #743 | 1y | PR #588 (Copilot SWE agent) closed without merge. PR #743 supersedes |
| 565 | TestIncludeFileNoExecutions is flaky | No | 1y | Test reliability |
| 520 | EXIT command not parsed in PowerShell -Q | PR #636 | 2y | PR reviewed, ship it |
| 494 | Does not parse ANSI text files correctly | No | 2.3y | Localization/codepage |
| 470 | LocalDB connections | No | 2.5y | Assigned to stuartpa, 11 comments |
| 393 | In --help, REMOVE -h | No | 3y | Help flag conflict |
| 347 | LoginTimeOut default value not correct | No | 3y | Assigned to JyotikaGargg |
| 305 | accept_eula saved into sqlconfig by viper | No | 3y | Viper config leak |
| 249 | Generated connection strings incorrect | No | 3y | 9 comments, assigned stuartpa |
| 45 | Doesn't validate custom batch separator | No | 4y | Original codebase bug |

---

## 5. PR Review Status -- All Reviewed

### Merge Queue (reviewed, ship it)

| PR | Title | Verdict | Notes |
|----|-------|---------|-------|
| #703 | Redundant ToLower | **Ship it** | Clean one-liner |
| #702 | Timeout hang fix | **Ship it** | CI re-triggered |
| #684 | Config extension validation | **Ship it** | Your PR, awaiting merge |
| #743 | wget exit code fix | **Ship it** | Created this session |
| #680 | Remove SQL Edge | **Ship it** | Reviewed this session |
| #633 | Deprecated ioutil/lint | **Ship it** | Reviewed this session |
| #639 | README freshness | **Ship it** | Reviewed this session |

### Feature PRs (reviewed, ship it with notes)

| PR | Title | Branch | Verdict | Fixes Pushed | Key Notes |
|----|-------|--------|---------|--------------|-----------|
| #688 | sqlcmd open vscode/ssms | pr-685 | **Ship it** | Yes (squashed) | All fixes folded into squashed commit de898f1. Verified clean. |
| #637 | --no-bom flag | no-bom | **Ship it** | No | Clean implementation |
| #636 | Multi-line EXIT(query) | exit-query | **Ship it** | No | Clean implementation |
| #631 | -p print statistics | print-statistics | **Ship it** | Yes (1 commit) | Parameterized io.Writer for #632 compat |
| #632 | :perftrace and :help | perftrace | **Ship it** | Yes (1 commit) | Merge note for #631 integration |
| #628 | -R regional settings | regional-settings | **Ship it** | Yes (1 commit) | French separator fix, dedup, dead import cleanup |
| #635 | -r bare flag default | r-default-fix | **Ship it** | No | Wrong `Fixes #604` **FIXED**. Scope verified safe |
| #630 | :serverlist command | serverlist-command | **Ship it** | Yes (1 commit) | Merge AFTER #632 (`:help` conflict). Dual-entry bug, README batch script, test DB dep **FIXED** |
| #609 | ASCII table format | main (fork) | **Ship it with fixes** | No | Missing copyright header, weakened test helper, memory buffering risk |

### Other PRs (automated / needs review)

| PR | Title | Status |
|----|-------|--------|
| #740 | Release 1.11.0 | Release Please, automated |
| #675 | Localization check-in | OneLocBuild, automated |
| #588 | Bak restore corruption fix | **CLOSED** without merge. Superseded by PR #743 |
| #729 | Docker Moby SDK migration | **CLOSED** without merge. PR #728 merged security fixes instead |

### Recommended Merge Order

1. #703, #702, #684, #743 (clean, no dependencies)
2. #680, #633, #639 (clean, no dependencies)
3. #637, #636 (independent features)
4. #632 THEN #630 (`:help` conflict -- #632 must merge first)
5. #631 (after #632 merges, update `printStatistics` call site to use `GetStat()`)
6. #628 (independent)
7. #688 (larger change, JSONC + atomic write)
8. #635 (`Fixes #604` linkage already fixed; #604 now closed via PR #707)
9. #609 (after copyright header + test helper fixes)

---

## 6. High-Value Enhancements (No PR)

Issues with significant community interest or strategic value.

| # | Title | Comments | Notes |
|---|-------|----------|-------|
| 478 | Support ADO.NET connection strings | 11 | High demand |
| 468 | Kerberos ticket cache on Linux | 16 | High demand, auth gap |
| 595 | Integrated security with Kerberos on Linux | 2 | Related to #468 |
| 481 | go-prompt for interactive shell | 10 | Nice UX improvement |
| 574 | Provide Docker image via registry | 3 | Distribution improvement |
| 539 | Output data not easy to parse | 8 | PR #609 partially addresses |
| 620 | Native CSV output format | 0 | Quick win, complements #609 |
| 619 | Markdown table output format | 0 | Quick win, complements #609 |
| 486 | Import/export Parquet files | 1 | Large scope |
| 151 | JSON/YAML/XML query results | 0 | Related to #539, #620, #619 |
| 612 | XDG Base Directory Specification | 0 | Linux standards compliance |
| 660 | ODBC compatibility mode via env var | 0 | Strategic for migration |
| 629 | -D flag for DSN support | 0 | ODBC compat |

---

## 7. Stale / Low Priority

Issues 2+ years old with no recent activity and no clear path forward.

| # | Title | Age |
|---|-------|-----|
| 7 | Evaluate stdout performance | 5y |
| 23 | Basic BCP functionality | 5y |
| 24 | Hierarchical JSON/XML BCP | 5y |
| 28 | Make available in ADO/GH environments | 5y |
| 29 | ADO task for go-sqlcmd | 5y |
| 80 | Custom date/time format strings | 4y |
| 92 | Define a settings file | 4y |
| 108 | Input connection string | 4y |
| 155 | Add to Windows Store | 4y |
| 202 | Man page for sqlcmd | 3y |
| 210-218 | Various stuartpa enhancements | 3y |
| 236 | sqlcmd open ads on WSL | 3y |
| 238 | Expand sqlconfig scope | 3y |
| 245 | update-context/update-endpoint | 3y |
| 252-276 | Various stuartpa enhancements | 3y |
| 291-332 | Various stuartpa enhancements | 3y |
| 410/412 | Telemetry issues | 3y |
| 432 | ADS Login fails after container recreate | 3y |
| 446-453 | Various enhancements | 3y |
| 482 | Secure Enclaves support | 3y |
| 496 | "not yet implemented" error for switches | 2y |

---

## 8. Packaging / Distribution

| # | Title | Age | Notes |
|---|-------|-----|-------|
| 532 | Ubuntu 24.04: sqlcmd not available | 2y | 4 comments |
| 528 | Ubuntu 22.04: v1.6.0 not available | 2y | |
| 524 | No package in RHEL 9 repo | 2y | Assigned stuartpa |
| 523 | RHEL 8 misses v1.6.0 | 2y | Assigned stuartpa |
| 549 | No x64 option after v1.60 | 2y | |
| 316 | Install error on Ubuntu | 3y | |
| 564 | WinGet CI action | 1y | Has open PR #564 |

---

## 9. Hostile Review Findings by PR

Legend: FIXED = addressed in a previous session commit. NEW = found in this review pass, not yet fixed.

> **⚠️ Orphaned commits**: Entries marked `FIXED [hash]` below were local-only commits.
> The following have been pushed: 6fafad5 (origin/regional-settings), 7a47de2 (origin/r-default-fix).
> Still orphaned (can't push): 225fd99 (#609 fork), 9e49b7b/#1be03d3 (#636 no origin branch),
> 3e358b6 (#684 copilot-swe-agent fork), 2b90521 (#639 upstream PR), 51d3fb1 (#702 upstream PR).

---

### PR #703 (redundant ToLower removal) -- Clean
No issues found.

### PR #702 (query timeout test) -- FIXED
- ~~Test can hang CI indefinitely~~: FIXED [51d3fb1]. Changed to `exec.CommandContext` with 30s deadline. Added `context.DeadlineExceeded` check with `t.Fatal`.
- **Informational**: `rows.Close()` could block on a stuck connection (driver-dependent, low risk).

### PR #684 (config file extension validation) -- FIXED
- **Acknowledged: Dotted filenames rejected**: Intentional -- PR requires .yaml/.yml for custom config files.
- **Acknowledged: Extension check is theater**: Viper hardcodes YAML regardless. The check prevents confusing filenames, not content validation.
- ~~Usage string contradicts default~~: FIXED [3e358b6]. Changed to `"configuration file (if named, must use .yaml or .yml extension)"`.
- ~~Error bypasses localization~~: FIXED [3e358b6]. Changed `fmt.Errorf` to `localizer.Errorf`.
- ~~Tests leave temp files~~: FIXED [3e358b6]. Added `t.Cleanup` with `os.Remove` for test config files.

### PR #743 (container exec reliability) -- FIXED
- ~~Test passes for wrong reason~~: FIXED. Moved URL validation before `mkdir` so the `file==""` check is reached before any container operations.
- ~~mkdir exit code discarded~~: FIXED. Exit code now checked with panic on failure.
- **Acknowledged: Docker ExecInspect race**: Known Docker issue (moby#42408). Comment added; race window negligible for short-lived commands.
- **Acknowledged: host.docker.internal**: Documented as intentional -- enables host reachability on Linux Docker Engine (Docker Desktop sets it automatically).

### PR #680 (gotext update) -- Confirmed non-issue
- **Verified**: Ran `go generate ./internal/translations/...` -- gotext confirmed catalog is already up to date. "Stale entries" finding was noise from removed SQL Edge strings, not actual stale references.

### PR #633 (GC safety fix) -- Clean
No issues found. Correctly fixes a latent `unsafe.Pointer` GC safety bug.

### PR #639 (auth mode docs) -- FIXED
- ~~`-p[1]` description unclear~~: FIXED [2b90521]. Expanded from "optional colon format" to explain what the flag does.
- ~~`tokenfilepath` described as "parameter" with no context~~: FIXED [2b90521]. Clarified it is a path to the federated token file.
- ~~`ActiveDirectoryClientAssertion` has no usage guidance~~: FIXED [2b90521]. Added: provide `client_id@tenant_id` as username, signed JWT as password.

### PR #637 (--no-bom) -- Clean
No issues found. Note: will have merge conflict with `redirectWriter` refactoring on main.

### PR #636 (multi-line EXIT) -- FIXED (was CRITICAL)
- ~~Infinite continuation loop~~: FIXED [1be03d3]. `isExitParenBalanced` now returns `true` when `depth < 0`, preventing infinite loop on malformed input like `EXIT ())`.
- ~~Nested block comments mishandled~~: FIXED [1be03d3]. Replaced `inBlockComment` bool with `commentDepth` counter (int) to properly track nested `/* */` blocks.
- ~~No bound on continuation input~~: FIXED [1be03d3]. Added `const maxContinuationLines = 1000` limit in `readExitContinuation`.
- Tests added: negative depth rejection, nested block comments, continuation line limit enforcement.

### PR #632 (perftrace) -- FIXED
- ~~Integration gap~~: FIXED in previous session (merge note in godoc).
- ~~:HELP alias lookup fails~~: Dead code removed. The alias fallback loop (searching by `cmd.name`) never matched because ED/R are resolved by regex, not by map key or name. Removed the loop; `:HELP ED` now correctly returns "not a recognized command."
- ~~:HELP unknown shows listing instead of error~~: FIXED. Returns `"'%s' is not a recognized command"` error.
- ~~Sort by help text not name~~: FIXED. Now sorts by `entries[i].name`.
- ~~BLOCKING: TestHelpCommand asserts wrong behavior~~: FIXED. Test now asserts `Error` for NOSUCHCMD and captures stdout via `os.Pipe`.
- ~~Medium: :HELP writes to :OUT redirect~~: FIXED. `helpCommand` now writes to `os.Stdout` directly, matching ODBC sqlcmd behavior.
- ~~Medium: No test for :HELP ED alias resolution~~: FIXED. Test added asserting `:HELP ED` returns error (ED is not a map key; short forms are resolved by regex during command dispatch, not by `:HELP`).
- ~~Low: BOM comment context lost~~: FIXED. Restored context comment on `outCommand` explaining BOM behavior difference from ODBC sqlcmd.
- **Low: Cross-PR integration with #631**: `printStatistics` uses `GetOutput()` not `GetStat()`. When #631 merges, the call site needs updating.
- ~~Low: stderr/stdout check order swapped~~: FIXED. `redirectWriter` now checks stdout before stderr, matching upstream order.
- ~~Low: No test for stat file close chain~~: FIXED. `TestPerftraceCloseChain` added.

### PR #630 (:serverlist) -- FIXED
- ~~Default instance two entries~~: FIXED.
- ~~README batch script bug~~: FIXED.
- ~~Tests required DB connection~~: FIXED.
- ~~ListLocalServers writes to os.Stderr~~: FIXED. `ListLocalServers` now returns error; callers write to appropriate error stream.
- **False positive: helpCommand rejects :HELP SERVERLIST**: Map key is `"SERVERLIST"` matching the lookup. Issue only affects alias mismatches (fixed in #632).
- **Low: Hardcoded help text drift risk**: Help text is inline strings, not sourced from the localizer.
- **Low: parseInstances behavior change**: Missing `InstanceName` field handling changed silently.
- **Low: Default instance format changed**: Removed `(local)` entry from output.

### PR #631 (print statistics) -- FIXED
- ~~Hardcoded output~~: FIXED (parameterized `io.Writer`).
- ~~GO n prints n separate stat lines~~: FIXED. `runQuery` now returns `(int, int64, error)` with elapsed time. `goCommand` accumulates `totalElapsedMs` and prints once. `exitCommand` prints after query.
- **Low: Timing includes formatting overhead**: `elapsedMs` measures wall time including output formatting, not just query execution.
- **Low: Statistics mix with query output**: Stats go to `GetOutput()`, so they interleave with result data in redirected output.
- **Low: Sub-millisecond queries report 1ms**: `time.Since` truncated to milliseconds; fast queries always show "1 ms".
- **Low: Tests require live SQL Server**: No mock/unit tests for statistics logic.

### PR #628 (regional settings) -- FIXED
- ~~French thousand separator~~: FIXED (previous session).
- ~~Code duplication~~: FIXED (previous session).
- ~~Dead import~~: FIXED (previous session).
- ~~strings.Replace inconsistency~~: FIXED (previous session).
- ~~Division-by-zero on extreme scale~~: FIXED [6fafad5]. `pow10` now clamps n<=0 to 1 and n>18 to 18.
- ~~Locale lookup ignores Swiss region~~: FIXED [6fafad5]. Added `de-CH`/`it-CH` decimal separator overrides and Swiss typographic apostrophe for thousands.
- **Acknowledged: FormatNumber ignores + prefix**: `+123` passed through as literal. Matches ODBC sqlcmd behavior.
- ~~FormatMoney truncates silently~~: FIXED [6fafad5]. Added rounding of 5th+ decimal digit with carry propagation via `incrementIntString` helper.
- **Acknowledged: Windows LCID table incomplete**: Only ~20 LCIDs mapped. Sufficient for common locales; uncommon ones fall through to US defaults (matches ODBC behavior).
- **Acknowledged: Signature conflict with PR #609**: Both PRs change `NewSQLCmdDefaultFormatter`. Merge order matters -- whoever merges second resolves the trivial conflict.

### PR #688 (sqlcmd open vscode/ssms) -- FIXED (verified on squashed de898f1)
- ~~JSONC blocker~~: FIXED. `patchJSONCKey` surgically replaces only the target key value, preserving all comments and formatting. Uses `tidwall/jsonc` for read path only.
- ~~Non-atomic write risk~~: FIXED. `writeSettingsRaw` does temp-file-then-rename with direct-write fallback.
- ~~Atomic rename loses permissions~~: FIXED. Now captures original mode via `os.Stat` and applies with `tmp.Chmod`.
- ~~Test safety~~: FIXED. Uses `t.Setenv(testSettingsEnvVar)` (auto-cleaned) instead of package-level global.
- ~~tool_linux.go stdout conflict~~: FIXED. Both Linux and Windows `generateCommandLine` set `Stdout`/`Stderr` to buffers; `Run()` calls `cmd.Run()` not `Output()`.
- ~~Password persists in clipboard~~: FIXED. Warning message: "paste it when prompted, then clear your clipboard".
- ~~`exec.Cmd.Args` missing program name~~: FIXED. `Args: append([]string{t.exeName}, args...)` on both platforms.
- **New: `Launch()` for GUI apps**: ADS and SSMS use `cmd.Start()` (fire-and-forget). VS Code uses `cmd.Run()` for `--install-extension`. Correct separation.
- **New: `vscode://` URI protocol**: Opens mssql extension connection handler without second window.
- **False positive: tool.Run error not checked**: All `tool.Run()` calls have `c.CheckErr(err)`. Code is correct.
- **False positive: Username escaping incomplete**: Go's `exec.Cmd` passes args directly, no shell interpolation. Escaping is not needed.

### PR #635 (-r bare flag) -- FIXED
- ~~Wrong Fixes #604~~: FIXED (previous session).
- ~~Broad scope~~: VERIFIED SAFE.
- ~~Empty string defaults to 0 silently~~: FIXED [7a47de2]. Clarified comment: the empty-value path only applies to `-r` (errorsToStderr) bare flag usage; other callers always supply a value.

### PR #609 (ASCII table) -- FIXED
- ~~Missing copyright header~~: FIXED [225fd99]. Added Microsoft copyright header to `format_ascii.go` and `format_ascii_test.go`.
- **Acknowledged: Breaking API**: `NewSQLCmdDefaultFormatter` signature changed (public function). Document in changelog.
- ~~rowcount not incremented~~: FIXED [225fd99]. `asciiFormatter.AddRow` now increments `f.rowcount`.
- **Acknowledged: Unbounded memory buffering**: All rows buffered before output. This is by design -- ASCII table formatting requires all rows to calculate column widths.
- **Acknowledged: Hardcoded stdout fd**: `term.GetSize(os.Stdout.Fd())` fallback is `maxWidth = 1000000` (no wrapping), which is reasonable for redirected output.
- ~~Test helper weakened~~: FIXED [225fd99]. Restored `assert.NoError` in `setupSqlCmdWithMemoryOutput` (was `t.Logf` which silently swallowed connection failures).
- **Acknowledged: Signature conflict with PR #628**: Both change the formatter constructor. Merge order matters.

---

### Pass 3 Findings

Third hostile review pass across all 16 PRs. Focus: find anything passes 1-2 missed.

#### PR #636 -- NEW BUG FOUND AND FIXED

- ~~Trailing whitespace after `)` in continuation line causes `InvalidCommandError`~~: FIXED [9e49b7b]. `readExitContinuation` result wasn't trimmed, so `EXIT(\n)   ` failed `HasSuffix(params, ")")`. Added `strings.TrimSpace(params)` after continuation read.
- ~~`TestExitCommandMultiLineInteractive` panicked without DB~~: FIXED [9e49b7b]. Test used `setupSqlCmdWithMemoryOutput` which requires a live SQL connection. Rewrote to use empty `EXIT()` parens so no query execution needed.
- Test added: `TestExitCommandContinuationTrailingWhitespace` verifies trailing spaces don't cause errors.

#### PR #684 -- NEW LOW FINDING

- **Low: Localization key mismatch**: Go code uses `"configuration file (if named, must use .yaml or .yml extension)"` but catalog JSON key is `"YAML configuration file (.yaml or .yml extension)"`. The translation will never match. English fallback still shows the correct text.

#### PR #628 -- Reconfirmed LOW

- **Low: FormatDateTime/FormatTime division-by-zero at scale >= 10**: `pow10(scale)` can exceed `1e9`, causing integer division by zero. Unreachable from SQL Server (max fractional seconds scale = 7). Guard would be trivial (`if scale > 9 { scale = 9 }`) but not worth the churn.

#### PR #688 -- ALL CLEAR (squashed code reviewed)

All previous Low findings resolved in squashed commit de898f1. `patchJSONCKey` preserves comments, `writeSettingsRaw` preserves permissions, `t.Setenv` eliminates global state.

#### All other PRs -- CLEAN on pass 3

PRs #609, #635, #702, #631, #632, #630, #637, #680, #639, #743, #703, #633: No new findings beyond what was already documented in passes 1-2.

---

## 10. Not Fixed and Why

Items deliberately left unfixed, grouped by reason.

### ODBC compatibility debt

Things we intentionally do wrong (or leave suboptimal) to match ODBC sqlcmd behavior. Each is a conscious trade-off: breaking compat would surprise users migrating from ODBC sqlcmd.

| PR | Finding | What's wrong | Why we keep it |
|----|---------|-------------|----------------|
| #631 | `-p` timing includes formatting overhead | Wall time from query submit to last row formatted. Inflated for large result sets. | ODBC sqlcmd `-p` measures the same wall time. Changing it breaks parity. Pure server time is available via `:TIMING ON` (#632). |
| #631 | Statistics interleave with query output | `-p` stats go to `GetOutput()`, mixing with result data in redirected output. | ODBC sqlcmd does the same. Moving stats to stderr would break scripts that parse stdout. |
| #628 | `FormatNumber` ignores `+` prefix | `+123` passed through as literal instead of formatted. | ODBC sqlcmd passes `+` through as literal. |
| #628 | Windows LCID table only covers ~20 locales | Uncommon locales fall through to US defaults silently. | ODBC sqlcmd has the same limited table. Sufficient for all commonly-used locales. |

### By design (not ODBC-related)

| PR | Finding | Rationale |
|----|---------|-----------|
| #609 | Unbounded memory buffering in ASCII formatter | ASCII table rendering requires all rows upfront to calculate column widths. Inherent to the format. |
| #609 | `term.GetSize(os.Stdout.Fd())` hardcoded to stdout | Redirected output gets `maxWidth = 1000000` (no wrapping). Reasonable -- redirected output shouldn't wrap to a terminal width. |
| #684 | Extension check is "theater" (Viper hardcodes YAML) | Prevents confusing filenames (`.json`, `.toml`), not content validation. Intentional UX guardrail. |
| #684 | Dotted filenames like `my-server.config` rejected | Intentional. PR's purpose is to enforce `.yaml`/`.yml` for named config files. |

### Merge-order / cross-PR coordination (not a code fix)

| PR | Finding | Rationale |
|----|---------|-----------|
| #609 + #628 | Signature conflict on `NewSQLCmdDefaultFormatter` | Both PRs change the same function signature. Whoever merges second resolves the trivial conflict. No code fix possible until merge order is decided. |
| #632 + #631 | `printStatistics` uses `GetOutput()` not `GetStat()` | When #631 merges, the call site needs updating. Deferred to merge time -- fixing now would create a different conflict. |
| #637 | Merge conflict with `redirectWriter` refactoring | Noted, no action until main branch state is known. |

### Fix it

Items that should be fixed. "Not worth the churn" was lazy -- most are trivial.

| PR | Finding | Action | Status |
|----|---------|--------|--------|
| ~~#628~~ | ~~`pow10` div-by-zero at scale >= 10~~ | ~~1-line clamp.~~ | ~~PUSHED to origin/regional-settings~~ |
| ~~#631~~ | ~~Timing includes formatting overhead~~ | ~~Moved startTime after BeginBatch.~~ | ~~PUSHED [473ca64] to origin/print-statistics~~ |
| ~~#631~~ | ~~Sub-ms queries report "1 ms"~~ | ~~Shows "< 1" for sub-millisecond.~~ | ~~PUSHED [473ca64] to origin/print-statistics~~ |
| ~~#631~~ | ~~No unit tests for statistics logic~~ | ~~7 subtests added (TestPrintStatisticsUnit).~~ | ~~PUSHED [473ca64] to origin/print-statistics~~ |
| ~~#630~~ | ~~`parseInstances` behavior change undocumented~~ | ~~Comment documenting deliberate change.~~ | ~~PUSHED [b5053d1] to origin/serverlist-command~~ |
| #684 | Localization key mismatch for `--sqlconfig` | Rerun `gotext extract` to sync catalog key. | CAN'T PUSH (copilot-swe-agent's fork). [Comment left on PR](https://github.com/microsoft/go-sqlcmd/pull/684#issuecomment-4316796435). |

### Won't fix (real reasons)

| PR | Finding | Rationale |
|----|---------|-----------|
| #702 | `rows.Close()` blocks on stuck connection | Driver-dependent. Not our PR. No mitigation at sqlcmd layer. |
| #630 | Hardcoded help text (not from localizer) | Every `:command` help string is hardcoded. Localizing one is inconsistent with all the others. |
| #609 | Breaking API (`NewSQLCmdDefaultFormatter` signature) | Not a code fix. Changelog entry at release time. |

### Documentation-only PRs -- all fixed

No remaining unfixed documentation findings.

### False positives (no fix needed)

| PR | Finding | Rationale |
|----|---------|-----------|
| #688 | `tool.Run` error not checked | All callers use `c.CheckErr(err)`. Code is correct. |
| #688 | Username escaping incomplete | Go's `exec.Cmd` passes args directly to the OS, no shell interpolation. Escaping is unnecessary. |
| #630 | `:HELP SERVERLIST` rejected | Map key is `"SERVERLIST"` which matches the lookup. This was an alias issue fixed in #632, not a #630 bug. |
| #680 | Stale `messages.gotext.json` entries | Ran `go generate` -- gotext confirmed catalog is up to date. The "stale" entries were noise from removed SQL Edge strings. |
