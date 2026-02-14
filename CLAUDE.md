# CLAUDE.md

You are Claude Code, helping me build and maintain a self‑hostable group scheduling app written in Go (Gin) with Tailwind for the UI.

Your job is to:
- Understand the existing architecture before generating code.
- Propose small, iterative changes rather than big rewrites.
- Keep the project easy to self‑host (simple Docker, minimal dependencies).
- Respect Go idioms, Gin conventions, and Tailwind best practices.
- **Use the available skills** in `.claude/skills/` when they match the task (e.g., Golang work → golang skill; Tailwind → tailwindcss-development skill).

---

## Project overview

- Language: Go
- Backend framework: Gin
- Frontend: Server‑rendered HTML templates + Tailwind CSS + PWA (installable)
- Database: SQLite (file-based, embedded) — **not yet implemented**
- Purpose: Create polls for scheduling group events/meetups; share links for participants to vote.

Goals:
- Anonymous, URL-based access: separate paths for poll creators (admin) and voters (participants).
- No authentication, logins, or user accounts.
- Simple, fast, easy to deploy in a home‑lab or small VPS.
- Clean, readable code that’s friendly to AI‑assisted development.
- Minimal magic, explicit behavior, and predictable side effects.

---

## Repository structure

**Current layout** (update this as the project evolves):

```
meetkat/
├── main.go                  # Application entrypoint
├── go.mod / go.sum          # Go 1.25, Gin dependency
├── package.json             # Node deps (@tailwindcss/cli)
├── internal/                # Domain code (handlers, services, repositories)
│   ├── config/
│   ├── handler/
│   ├── i18n/
│   ├── middleware/
│   ├── poll/
│   ├── sqlite/
│   └── view/
├── web/                     # Web assets (templates + static files)
│   ├── templates/           # Go html/template files
│   │   ├── layouts/
│   │   │   └── base.html    # Base layout (head, theme switcher, content block)
│   │   ├── index.html       # Home / hero page
│   │   ├── new.html         # Create poll form
│   │   ├── poll.html        # Vote on a poll
│   │   ├── admin.html       # Admin view
│   │   └── 404.html         # Not found
│   └── static/              # Static assets served at /static
│       ├── css/
│       │   ├── input.css    # Tailwind source (theme, variables, dark mode)
│       │   └── style.css    # Compiled output (gitignored)
│       ├── js/
│       │   ├── app.js
│       │   └── sw.js        # Service worker (network-first caching)
│       └── manifest.json    # PWA web app manifest
├── data/                    # meetkat.db (runtime, gitignored)
├── .air.toml                # Air live-reload config
├── Dockerfile               # Multi-stage build (CSS → Go → Alpine); copies icons + manifest
└── docker-compose.yml       # Docker Compose for deployment
```

For generic Go project-structure patterns, see `.claude/skills/golang/references/project-structure.md`.

---

## Available skills

Use these project-specific skills automatically when relevant. They are loaded from `.claude/skills/`:

- **Golang skill** (`.claude/skills/golang/skill.md`): For all Go/Gin code generation, refactoring, testing, and Docker setup. Use it for backend tasks like handlers, services, repositories, or migrations.
- **TailwindCSS development skill** (`.claude/skills/tailwindcss-development/skill.md`): For frontend work—templates, Tailwind classes, CSS builds, responsive design, and component creation.

**How to invoke:**
- Claude Code auto-triggers them based on context/slash commands.
- Explicitly reference: "Use the golang skill to implement X" or "Apply tailwindcss-development skill for Y."

Consult their supporting files (e.g., references/project-structure.md in golang skill) as needed.

---

## How to run the project

- **Live-reload dev server** (recommended):
    - `air` — watches `.go` and `.html` files, rebuilds and proxies (app on `:8080`, proxy on `:8090`).
- **Start dev server manually:**
    - `go run .`
- **Run tests:**
    - `go test ./...`
- **Build binary:**
    - `go build -o meetkat .`
- **Tailwind CSS** (use tailwindcss-development skill for details):
    - `npm run dev:css` — watch mode, recompiles on template changes.
    - `npm run build:css` — one-shot minified build for production.
- **Docker:**
    - `docker compose up --build`

When you propose changes that affect these commands, update this section as well.

---

## Go & Gin coding guidelines

**Consult the golang skill first** for detailed patterns.

Follow idiomatic Go and common Gin practices:

- Organize code by domain (polls, scheduling) rather than by technical layer only.
- Keep handlers thin: they should parse/validate input, call domain services, and return responses.
- Domain services should not depend on Gin; accept/return plain Go types.
- Prefer constructor functions for services (e.g. `NewPollService(repo PollRepository)`).
- Use context (`context.Context`) for cancellation/deadlines where appropriate.
- Avoid global state; use dependency injection via structs and constructors.
- Prefer returning `(T, error)` rather than panicking; keep panics for truly exceptional situations.
- Add tests for new public functions and for bug fixes.

PWA / Service worker:
- The service worker (`web/static/js/sw.js`) is served at `/sw.js` (root scope) via a dedicated Gin route in `main.go`.
- The manifest (`web/static/manifest.json`) and theme-color meta tag are linked in `base.html`.
- The SW uses a network-first strategy: tries the network, caches successful responses, falls back to cache when offline.
- When updating cached assets, bump the `CACHE_NAME` version string in `sw.js`.

Routing:
- Use URL-based access only:
    - `/polls` or `/new` – create new poll (admin).
    - `/polls/:simpleId` – view/vote (participants).
    - `/polls/:complexId` – edit/close poll (admin).
- Group routes logically (e.g. `/polls`, `/polls/:id`, `/polls/:id/votes`).
- Use middlewares for cross‑cutting concerns (logging, recovery).

Error handling & responses:
- For HTTP handlers, use consistent error responses (templates for HTML).
- Avoid leaking internal details in error messages returned to users.

---

## Tailwind & frontend guidelines

**Consult the tailwindcss-development skill first** for setup, utilities, and best practices.

- Use Tailwind utility classes for styling; avoid writing large custom CSS unless necessary.
- Prefer semantic HTML structure and accessible components.
- Keep templates simple and composable:
    - Layout templates (base layout, nav, footer).
    - Reusable partials for poll rows, inputs, etc.
- Distinguish admin vs. participant views clearly via URL paths.
- When generating HTML, do **not** mix heavy business logic into templates. Keep logic in Go and pass prepared data.
- The base layout includes PWA meta tags (`manifest`, `theme-color`) and registers the service worker. New pages inheriting from `base.html` get PWA support automatically.

---

## Database & persistence (planned — not yet implemented)

**Follow golang skill for SQLite integration when this is added.**

- SQLite only: use `database/sql` + `modernc.org/sqlite` or `github.com/mattn/go-sqlite3`.
- Use a repository or data‑access layer per aggregate (e.g. `PollRepository`).
- Avoid embedding SQL directly in handlers; put queries in repositories.
- For schema changes:
    - Propose an init script or embedded migration in Go.
    - Keep it simple; SQLite file auto‑creates if missing.
- When designing tables:
    - Use clear, explicit column names.
    - Normalize where practical for poll data (e.g. polls, options, votes).

---

## Testing

**Use golang skill for test patterns.**

- For new features:
    - Add or update tests for core domain logic (poll creation, voting, closing polls, etc.).
    - Prefer table‑driven tests in Go.
- For bug fixes:
    - First reproduce the bug with a failing test.
    - Then fix the code.
    - Ensure the test passes and avoids regressions.

Do not introduce heavy testing frameworks unless already in use; stick to Go’s standard testing library unless otherwise specified.

---

## Git & change workflow

When performing non‑trivial changes:

1. Explain your plan briefly before modifying files.
2. Make changes in small, coherent steps.
3. Keep commit messages clear and descriptive:
    - `feat(polls): add multi‑day poll support`
    - `fix(polls): handle duplicate responses`
    - `refactor(http): extract poll loading middleware`

If you add new files:
- Clearly indicate the intended path and update the repository structure section above.
- Keep filenames and package names consistent with the rest of the project.

---

## How I want you to work

When I ask for help, follow this pattern unless I explicitly ask for something else:

1. **Clarify and plan**
    - Restate the goal in your own words.
    - List the files or areas you plan to touch (reference project-structure.md).
    - Propose a short step‑by‑step plan.
    - Note which skill(s) you'll use (golang/tailwindcss-development).

2. **Implement**
    - Apply changes stepwise, explaining each step briefly.
    - Prefer minimal, focused changes over broad refactors.
    - Respect existing patterns and conventions in this repo and skills.

3. **Validate**
    - Tell me which commands to run to verify the change (tests, build, manual checks).
    - Mention potential edge cases or follow‑up improvements.

4. **Document**
    - If needed, update relevant documentation (README, this CLAUDE.md, comments) when behavior changes.
    - Keep documentation concise and accurate.

If you’re unsure about a design choice or there are multiple good options, pause and present the trade‑offs instead of guessing.

---

## Project‑specific rules

- No authentication, logins, or user accounts – everything URL‑driven.
- SQLite for simplicity; no external DB servers.
- This is a self‑hostable app first; avoid SaaS‑only features or vendor lock‑in.
- Keep configuration simple: environment variables and a small config struct are preferred.
- Favor performance and simplicity over premature abstraction.
- Avoid introducing new major dependencies without a clear benefit.
- Maintain a clean public surface for the main package; internal complexity is fine as long as it’s well organized.

---

## Things to avoid

- Large, speculative refactors without prior discussion.
- Introducing patterns that conflict with idiomatic Go or Gin conventions (check golang skill).
- Adding JavaScript frameworks or heavy client‑side complexity unless explicitly requested.
- Generating code that won’t compile or clearly breaks the existing build/test setup.
- Writing overly clever code at the cost of readability.
- Any auth/login features – stick to URL‑based access.

---

## Behaviour

Expected runtime behaviours to preserve. Reference this section when modifying related code.

- **AJAX partial updates**: Vote operations (submit, edit, remove) use `fetch()` with `X-Requested-With: fetch` to POST, and the server returns only the `vote_table` HTML fragment. JS swaps `#vote-table-wrapper.innerHTML` and re-initializes event listeners via `initTable()`. Full-page redirects remain as the no-JS fallback.
- **Confirm-incomplete vote (two-click pattern)**: When a user submits a vote with unanswered options (hidden inputs still `""`), the submit button changes to an amber "are you sure?" state and the submission is **blocked**. Only on the **second** click does the vote actually send. Clicking any vote button while armed resets the button to its original state. This logic lives in the AJAX submit handler — not in a separate submit listener — to guarantee execution order.
- **Scroll position preservation**: After an AJAX table swap, `wrapper.scrollLeft` is saved before and restored after the innerHTML replacement so horizontal scroll position is not lost.

---

## Open questions / TODOs for Claude

(Use this as a living list you edit over time.)

- Refine the domain model for polls (time‑slot vs. generic option polls).
- Add basic test coverage for core poll flows.
- Optimize page load and bundle size once the core flows are stable.
