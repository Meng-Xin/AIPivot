# Project Memory

## 2026-04-26: Observability Foundation Execution Lessons

### What went wrong

- I over-applied the `superpowers` workflow to a small bootstrap task.
- I treated process completion as the goal instead of fast local delivery.
- I used `.worktrees` and multiple subagent review loops for work that should have stayed in the main workspace.
- I spent too much effort on exact-shape spec review details before proving the code could compile and run locally.
- I updated specs, plans, and review loops in lockstep when a direct code/config correction would have been enough.

### Rules for future small tasks in this repo

- For small bootstrap or single-service infra work, default to the current workspace, not `.worktrees`.
- First priority is local evidence: `go test ./...`, `go build ./...`, and `docker compose config` before heavy documentation or review loops.
- Use specs and plans only when they materially reduce risk. Do not generate long process artifacts for straightforward setup work.
- Do not use multi-agent review chains for low-complexity tasks unless the user explicitly asks for that workflow.
- Review should focus on runtime correctness, maintainability, and integration risk, not cosmetic or exact-text differences.
- Distinguish code issues from machine-environment issues early. Work around local cache and permission problems quickly instead of turning them into process delays.
- When a dependency path is clearly outdated, fix the code path directly and record the reason briefly. Do not trigger large process loops unless the change affects user-visible scope.

### Practical execution order

1. Create the minimum runnable files in the main workspace.
2. Resolve module and cache issues locally.
3. Run compile/test verification.
4. Start local dependencies if the machine environment allows it.
5. Only then tighten docs, polish config, or add deeper review.
