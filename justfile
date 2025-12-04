# Miniflux Reader - Justfile
# Run commands with: just <command>

# Default recipe - show available commands
default:
    @just --list

# Run the web app locally
web:
    cd web && go run main.go

# Build the web app binary
build-web:
    cd web && go build -o miniflux-reader main.go

# Install web dependencies
deps-web:
    cd web && go mod download

# Run web app with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev-web:
    cd web && air

# Build Docker image
docker-build:
    docker build -t miniflux-reader ./web

# Run Docker container
docker-run:
    docker run -d -p 5601:5601 --env-file web/.env --name miniflux-reader miniflux-reader

# Stop Docker container
docker-stop:
    docker stop miniflux-reader && docker rm miniflux-reader

# Clean build artifacts
clean:
    rm -f web/miniflux-reader
    rm -f web/.env

# Setup environment file from example
setup:
    cp web/.env.example web/.env
    @echo "⚠️  Please edit web/.env with your Miniflux credentials"

# Run tests (when tests are added)
test:
    cd web && go test ./...

# Format Go code
fmt:
    cd web && go fmt ./...

# Run Go vet
vet:
    cd web && go vet ./...

# Lint code (requires golangci-lint)
lint:
    cd web && golangci-lint run
