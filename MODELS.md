# MODELS.md

Which AI model to use for which kind of work in this repo. [AGENTS.md](AGENTS.md) is the main
guide; this file only covers model choice.

This file is guidance, not configuration. Nothing switches models automatically:
- You pick the session's model with `/model`.
- Claude can apply this guidance only when it starts a subagent.

## Claude models

Models are named by **alias**. An alias always resolves to the newest version your plan includes, so
this file doesn't go stale or name a model you can't use. Run `/model` to see what your plan offers.
If a model isn't available, use the next one up.

| Alias | Use it for |
|---|---|
| `haiku` | Quick, mechanical work: renames, typo and doc fixes, formatting, small UI copy changes, codebase searches |
| `sonnet` | Default for day-to-day work: features, bug fixes, tests, dependency bumps, Docker/CI edits |
| `opus` | Hard problems: tile/projection math, anything that changes the number of billable Google API calls, concurrency, security review |
| `fable` | Writing work: README, docs, release notes. If your plan doesn't include it, use `sonnet` |

## Pick by area

| Area | Model | Why |
|---|---|---|
| `pkg/earth/` (lat/lng to tile math) | Opus | Off-by-one and projection errors are subtle, so tiles end up misaligned or duplicated |
| `internal/service/print.go` (tile fan-out, merge, temp files) | Opus | Mistakes either leak temp files or multiply Google API costs |
| `internal/service/ratelimit.go` | Opus | Concurrency, and it's the main control on API costs |
| `pkg/file/`, `cmd/` | Sonnet | Plain Go I/O and wiring |
| `ui/` (React components, fetch call) | Sonnet | Standard React work |
| Security-sensitive code: authentication, sessions, cookies/CORS/CSRF, payment webhooks, IAM and infrastructure policies | Opus | Bugs here leak accounts or money, and rarely show up in tests |
| Map rendering and print layout: styles, DPI and scale math, PDF page grids, attribution placement | Opus for the math, Sonnet for layout code | Scale and overlap errors only show up once printed |
| Infrastructure as code, deploy scripts | Sonnet, with Opus reviewing anything that grants permissions | Routine to write, risky to get wrong |
| Tests, CI workflows, Dockerfiles, compose files | Sonnet | Routine work that follows existing patterns |
| Licensing, terms-of-service and pricing research | Opus | Needs careful reading; always quote and link the live source |
| Docs, comments, schema descriptions | Haiku or Fable | Low risk; use Fable for longer writing |

## Rules of thumb

- Start with Sonnet. Move up to Opus if the first attempt misses something or the change touches
  how many Google tiles get requested.
- Use Haiku for read-only exploration and bulk mechanical edits, where you want speed over depth.
- For anything that changes the rules in `validateRequest` or `maxRadius`, have Opus review the
  change, even if another model wrote it. The same goes for security-sensitive code.
- Don't trust any model's memory for prices, quotas or license terms. Check the provider's current
  page, and note the date you checked.
- Changing models doesn't change the rules: every model follows [AGENTS.md](AGENTS.md).

## Claude Code

- Switch models in a session with `/model` (for example `/model opus`).
- Subagents can be given a different model (`haiku`, `sonnet`, `opus`, `fable`), so codebase
  searches can run on Haiku while the main session stays on Sonnet or Opus.
- To make a choice stick, define a subagent in `.claude/agents/<name>.md` with `model:` in its
  frontmatter, or set a default `model` in `.claude/settings.json`.
