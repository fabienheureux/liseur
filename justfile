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
    cd web && go build -o liseur main.go

# Install web dependencies
deps-web:
    cd web && go mod download

# Run web app with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev-web:
    cd web && air

# Build Docker image
docker-build:
    docker build -t liseur ./web

# Run Docker container
docker-run:
    docker run -d -p 5601:5601 --env-file web/.env --name liseur liseur

# Stop Docker container
docker-stop:
    docker stop liseur && docker rm liseur

# Clean build artifacts
clean:
    rm -f web/liseur
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
