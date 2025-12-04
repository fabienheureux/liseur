# Liseur, a Miniflux reader

A minimal RSS reader client for Miniflux with server-side rendering.

## Features

- **Three-column layout**: Categories/Feeds, Unread Entries, Entry Content
- **Semantic HTML**: Uses proper HTML5 semantic elements (`<aside>`, `<article>`, `<nav>`, etc.)
- **Collapsible categories**: Uses `<details>`/`<summary>` elements for native collapsing
- **Server-side rendering**: All HTML is rendered on the Go backend using the Miniflux Go client
- **Minimal JavaScript**: Navigation uses standard `<a href>` links - no JavaScript required

## Prerequisites

- A running Miniflux instance
- Miniflux API key

## Quick Start with Docker

The easiest way to run Miniflux Reader is with Docker:

```bash
docker run -d \
  -p 8080:8080 \
  -e MINIFLUX_API_URL=https://your-miniflux-instance.com \
  -e MINIFLUX_API_KEY=your-api-key-here \
  -e PORT=8080 \
  ghcr.io/fabienheureux/liseur:main
```

**Note:** The PORT environment variable sets the internal port the app listens on. The `-p` flag maps the host port to the container port (format: `host:container`).

Or using docker-compose:

```yaml
version: '3.8'

services:
  liseur:
    image: ghcr.io/fabienheureux/liseur:main
    ports:
      - "8080:8080"
    environment:
      - MINIFLUX_API_URL=https://your-miniflux-instance.com
      - MINIFLUX_API_KEY=your-api-key-here
      - PORT=8080
    restart: unless-stopped
```

To use a different port, change both the host port mapping and the PORT environment variable:

```yaml
    ports:
      - "3000:3000"  # Map host port 3000 to container port 3000
    environment:
      - PORT=3000    # App listens on port 3000 inside container
```

Then visit `http://localhost:8080`

## Development Setup

### Local Development (without Docker)

**Prerequisites:**
- Go 1.21 or higher

### Setup

1. **Install dependencies**:
   ```bash
   cd web
   go mod download
   ```

2. **Configure environment variables**:
   
   Copy the example environment file:
   ```bash
   cp web/.env.example web/.env
   ```
   
   Or if you have [just](https://github.com/casey/just) installed:
   ```bash
   just setup
   ```

   Edit `web/.env` and set your Miniflux credentials:
   ```
   MINIFLUX_API_URL=https://your-miniflux-instance.com
   MINIFLUX_API_KEY=your-api-key-here
   PORT=8080
   ```

   You can get your API key from your Miniflux instance:
   - Log in to Miniflux
   - Go to Settings → API Keys
   - Create a new API key

3. **Run the application**:
   ```bash
   cd web
   go run main.go
   ```
   
   Or with just:
   ```bash
   just web
   ```

   The server will start on `http://localhost:8080`

## Using Just Commands

If you have [just](https://github.com/casey/just) installed, you can use these shortcuts:

```bash
just web          # Run the web app
just build-web    # Build web binary
just docker-build # Build Docker image
just docker-run   # Run in Docker
just setup        # Create .env from example
just --list       # Show all available commands
```

## Usage

- **View all unread entries**: Click "All Entries" in the header or visit `/`
- **Filter by category**: Click on a category name in the sidebar
- **Filter by feed**: Click on a feed name under a category
- **Read an entry**: Click on an entry title in the middle column
- **Mark as read**: Click the "Mark as Read" button when viewing an entry
- **Open original**: Click "Open Original" to view the entry on its source website
- **Close entry**: Click the "×" button to close the entry view

## Project Structure

```
liseur/
├── web/                 # Web application
│   ├── main.go          # Go server with Miniflux client integration
│   ├── go.mod           # Go module definition
│   ├── templates/
│   │   └── index.html   # Main HTML template
│   └── static/
│       ├── styles.css   # CSS styling
│       └── app.js       # Client-side JS utilities
├── ios/                 # iOS app (future)
├── android/             # Android app (future)
├── Dockerfile           # Docker build configuration
├── justfile             # Just command runner recipes
└── README.md
```

## Architecture

The application uses:
- **Go standard library** for HTTP server and HTML templating
- **Miniflux Go client** (`miniflux.app/v2/client`) for API interactions
- **Server-side rendering** via Go templates
- **Pure CSS** for styling (no JavaScript frameworks)
- **Semantic HTML5** for accessibility and structure

## Building

### Build Docker Image Locally

```bash
docker build -t liseur .
```

### Build Binary

```bash
cd web
go build -o liseur main.go
```

Or with just:

```bash
just build-web
```

Then run with:

```bash
./web/liseur
```

## CI/CD

The project includes a GitHub Actions workflow that automatically:
- Builds Docker images on every push to `main`
- Publishes images to GitHub Container Registry (ghcr.io)
- Tags images with version tags when you create a release

To use the automated builds:
1. Push to your GitHub repository
2. The image will be available at `ghcr.io/fabienheureux/liseur:main`
3. Create a tag like `v1.0.0` to publish versioned images

## Environment Variables

All configuration is done through environment variables:

- `MINIFLUX_API_URL`: URL of your Miniflux instance (required)
  - Example: `https://miniflux.example.com`
- `MINIFLUX_API_KEY`: Your Miniflux API key (required)
  - Get this from Settings → API Keys in your Miniflux instance
- `PORT`: Server port (optional, defaults to 8080)
  - The port the application listens on inside the container
  - When using Docker, make sure to match this with your port mapping
