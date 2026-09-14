# SmileX-Deep-Study · Personal Learning Cockpit

**English** | [简体中文](README_CN.md)

> **Files are the database · AGENTS.md is the contract · your agent harness is the tutor · a single Go binary is the cockpit. The system itself has zero LLM integration.**

The learning loop is split in half: **understanding work** (import & extraction, Socratic tutoring, Feynman-style probing, quiz grading, weakness diagnosis) is done by the agent harness you already use (ZCode / Trae / Kimi CLI / WorkBuddy); **scheduling work** (FSRS spaced repetition, statistics, the management UI) is handled by this system's single Go binary. All data is plain Markdown + JSON files — any tool can read and write them directly.

See [docs/00-theory-gap-analysis.md](docs/00-theory-gap-analysis.md) for the theoretical foundation and design-gap analysis, and [docs/01-architecture.md](docs/01-architecture.md) for the architecture and data contract.

## Quick Start

```bash
# 1. Build the frontend (first time, or after frontend changes)
cd web && pnpm install && pnpm build && cd ..

# 2. Build the single binary (web/dist is embedded via go:embed)
go build -o deep-study ./server/cmd/server

# 3. Run from the repo root (the Workflow page's prompt feature depends on prompts/ at the root)
./deep-study
# → http://127.0.0.1:8788
```

Dev mode: `./deep-study` (port 8788) + `cd web && pnpm dev` (port 5173, `/api` proxy pre-configured).

Customization: `./deep-study -addr 0.0.0.0:8788 -data /path/to/data`

## Six Workflows

| Workflow                   | Executed by         | Usage                                                                                                |
| -------------------------- | ------------------- | ---------------------------------------------------------------------------------------------------- |
| W1 Import                  | UI upload + harness | Drag a file into the "Library" page → run `/study:import <filename> [topic]` in your harness         |
| W2 Reading Tutor           | harness             | `/study:tutor <topic>` — Socratic dialogue: questions first, tiered hints, never the answer up front |
| W3 Feynman Internalization | harness             | `/study:feynman <topic\|note>` — you explain, it probes; gaps are written back into the note         |
| W4 Spaced Review           | Web UI (zero LLM)   | "Review" page: recall first, then reveal; 4-level grading; FSRS scheduling                           |
| W5 Retrieval Quiz          | harness             | `/study:quiz <topic> [n]` — brand-new questions + grading + mastery write-back                       |
| W6 Diagnosis               | harness             | `/study:diagnose [topic]` — weakness report + targeted drills (a deliberate-practice loop)           |

## Harness Integrations

| Tool                                      | How it connects                                                                                                                          |
| ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **Any agent following the common format** | `.agents/skills/` (SKILL.md standard) + `.agents/commands/` + `AGENTS.md` — works out of the box                                         |
| **ZCode**                                 | Natively discovers `.agents/` (the generic fallback path of `.zcode/`); `/study:*` commands ready                                        |
| **Kimi CLI**                              | Natively reads `AGENTS.md`; open the repo root and paste the universal templates from `prompts/`                                         |
| **WorkBuddy**                             | Full adaptation: `CODEBUDDY.md` loaded by default + `.codebuddy/{rules,skills,commands}/` (with `/study:*` commands and skills in place) |
| **Trae**                                  | Settings → Rules → enable "Include AGENTS.md"; `.trae/rules/deep-study.mdc` is ready                                                     |

The "Workflow" page provides **one-click copyable commands and universal prompts** for every workflow, usable by any tool.

## Where the Data Lives

```
data/
├── inbox/     Raw uploaded materials       ├── notes/     Atomic notes (with a ## Gaps section)
├── library/   Archived materials + index   ├── cards/     Cards (FSRS scheduling state embedded)
├── topics/    Topic manifests              ├── sessions/  Tutor/Feynman/quiz/diagnose session logs
└── progress/  mastery.json + append-only review/recall logs
```

Red-line rules (see `AGENTS.md` for details): the `fsrs:` block in `cards/*.md` and the two log files are writable **only by the Go server**; logs are append-only; every mastery change must carry evidence; **any file-writing workflow must finish by running** **`curl -s http://127.0.0.1:8788/api/validate`** **and reducing** **`errors`** **to zero** (hard validation: schema, fsrs block, date formats, evidence chain). The agent's behavioral boundaries are constrained by role definitions under `roles/` (tutor / examiner / curious novice / coach / librarian), while skills define only the procedure — soft constraints plus hard validation together guarantee output quality.

## Tech Stack

- **Backend**: Go + gin + [go-fsrs](https://github.com/open-spaced-repetition/go-fsrs) (official FSRS implementation, target retention 0.90) + yaml.Node-based targeted frontmatter rewriting (preserves any extra fields the agent wrote)
- **Frontend**: Vite + React 19 + TypeScript + Tailwind 4 + daisyUI 5, embedded into the binary via `go:embed`
- **Deployment**: one static binary + a `data/` directory, no runtime dependencies

## Repository Layout

```
├── AGENTS.md                  # The authoritative contract (required reading for agents)
├── CLAUDE.md / CODEBUDDY.md / .trae/ / .codebuddy/   # Per-tool adapters (.codebuddy copies are synced from .agents/ by scripts/sync-adapters.sh)
├── .agents/                   # Generic agent assets: skills (SKILL.md) + slash commands
├── roles/                     # Role definitions (identity + discipline): tutor / examiner / novice / coach / librarian
├── prompts/                   # Universal prompt templates
├── server/                    # Go backend (cmd/ + internal/{api,store,fsrsx})
├── web/                       # Frontend source
├── docs/                      # Theory gap analysis + architecture docs
└── gui-test-screenshots/      # GUI test evidence screenshots
```

