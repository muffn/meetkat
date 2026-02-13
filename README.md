# meetkat

A self-hostable group scheduling app. Create polls, share a link, and let participants vote on the dates that work for them. No accounts, no logins -- just simple, URL-based scheduling.

## Features

- **Create polls** with a title, optional description, and date options
- **Share a link** for participants to vote on which dates work
- **Admin view** with a separate private link to manage the poll and remove votes
- **Dark mode** support with system preference detection
- **Embedded SQLite** database -- no external database server needed
- **Single binary** deployment with Docker support

## How it works

1. Create a poll at `/new` with your proposed dates
2. Share the participant link -- anyone with the link can vote
3. Use the admin link (shown after creation) to manage the poll, view results, and remove votes

There are no user accounts. Access is controlled entirely through URL-based links: one public link for voting, one private link for admin actions.

## Deployment

### Docker Compose (recommended)

Create a `docker-compose.yml`:

```yaml
services:
  meetkat:
    image: ghcr.io/muffn/meetkat:latest
    ports:
      - "8080:8080"
    volumes:
      - meetkat-data:/app/data

volumes:
  meetkat-data:
```

Then run:

```bash
docker compose up -d
```

The app will be available at `http://localhost:8080`. Poll data is persisted in the `meetkat-data` volume.

### Docker (standalone)

```bash
docker run -d \
  -p 8080:8080 \
  -v meetkat-data:/app/data \
  ghcr.io/muffn/meetkat:latest
```

### Build from source

Prerequisites: Go 1.25+, Node.js (for Tailwind CSS)

```bash
# Install CSS dependencies and build
npm install
npm run build:css

# Build the Go binary
go build -o meetkat .

# Run
./meetkat
```

The server starts on port `8080` by default.

### Configuration

| Variable | Default | Description |
|---|---|---|
| `MEETKAT_DB_PATH` | `data/meetkat.db` | Path to the SQLite database file |
| `GIN_MODE` | `debug` | Set to `release` for production |

## Development

### Prerequisites

- Go 1.25+
- Node.js (for Tailwind CSS CLI)
- [Air](https://github.com/air-verse/air) (optional, for live reload)

### Getting started

```bash
# Install dependencies
npm install

# Start Tailwind in watch mode (in one terminal)
npm run dev:css

# Start the Go server with live reload (in another terminal)
air
```

Air watches `.go` and `.html` files, rebuilds on changes, and proxies the app on port `8090` (app itself runs on `8080`).

To run without Air:

```bash
go run .
```

### Running tests

```bash
go test ./...
```

## Tech stack

- **Backend:** Go with [Gin](https://github.com/gin-gonic/gin)
- **Database:** SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)
- **Frontend:** Server-rendered Go templates with [Tailwind CSS](https://tailwindcss.com/) v4

## Project structure

```
meetkat/
├── main.go                  # Application entrypoint
├── internal/
│   ├── handler/             # HTTP handlers (Gin)
│   ├── poll/                # Domain model, service, repository interface
│   └── sqlite/              # SQLite repository implementation + migrations
├── web/                     # Web assets
│   ├── templates/           # Go html/template files
│   │   ├── layouts/
│   │   │   └── base.html    # Base layout
│   │   ├── index.html       # Home page
│   │   ├── new.html         # Create poll form
│   │   ├── poll.html        # Vote on a poll
│   │   ├── admin.html       # Admin view
│   │   └── 404.html         # Not found
│   └── static/              # Tailwind source + compiled output
│       ├── css/
│       └── js/
├── Dockerfile               # Multi-stage build
└── docker-compose.yml       # Docker Compose for deployment
```

## Roadmap

Planned features and improvements (in no particular order):

- [ ] Edit votes inline from the admin page
- [ ] Delete polls from the admin page
- [ ] Closing/archiving polls to prevent new votes
- [ ] "Maybe" option alongside yes/no
- [ ] Time-slot polls (not just dates)
- [ ] CSV/JSON export of poll results
- [ ] Optional poll expiration dates
- [ ] QR code generation for easy sharing
- [ ] i18n / multi-language support
