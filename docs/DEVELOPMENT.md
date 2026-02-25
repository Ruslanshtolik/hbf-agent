# Development Guide

This guide will help you set up your development environment and work efficiently with the HBF Agent project.

## Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)
- Docker (for containerized testing)

## Initial Setup

### 1. Clone the Repository

```bash
git clone https://github.com/Ruslanshtolik/hbf-agent.git
cd hbf-agent
```

### 2. Install Dependencies

```bash
make deps
# or
go mod download
go mod tidy
```

### 3. Build the Project

```bash
make build
# or
go build -o build/hbf-agent ./cmd/hbf-agent
```

## Development Workflow

### Quick Push to GitHub

We've set up multiple ways to quickly commit and push your changes:

#### Option 1: Using Make (Recommended)

```bash
make git-push MSG="Your commit message here"
```

#### Option 2: Using Quick Push Script (Bash/Git Bash)

```bash
./scripts/quick-push.sh "Your commit message here"
```

#### Option 3: Using Quick Push Script (Windows CMD)

```cmd
scripts\quick-push.bat "Your commit message here"
```

#### Option 4: Manual Git Commands

```bash
git add .
git commit -m "Your commit message"
git push origin main
```

### Other Useful Make Commands

```bash
# Check git status
make git-status

# Pull latest changes
make git-pull

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Run in development mode
make run-dev

# Build Docker image
make docker-build

# Show all available commands
make help
```

## Git Hooks

The project includes a pre-commit hook that automatically:
- Formats Go code with `go fmt`
- Runs `go vet` to catch common errors
- Adds formatted files to the commit

The hook is located at `.git/hooks/pre-commit` and runs automatically before each commit.

## Project Structure

```
hbf-agent/
├── cmd/
│   └── hbf-agent/          # Main application entry point
├── internal/
│   ├── agent/              # Core agent logic
│   ├── api/                # REST API server
│   ├── config/             # Configuration management
│   ├── firewall/           # Firewall management
│   ├── health/             # Health checking
│   ├── metrics/            # Metrics collection
│   └── servicemesh/        # Service mesh functionality
├── config/                 # Configuration files
├── deploy/                 # Deployment configurations
├── docs/                   # Documentation
├── scripts/                # Utility scripts
└── test-infrastructure/    # Testing infrastructure
```

## Testing

### Unit Tests

```bash
make test
```

### Integration Tests

```bash
# Using Docker Compose
cd test-infrastructure/docker-compose
docker-compose up -d

# Using Vagrant
cd test-infrastructure/vagrant
vagrant up

# Using Terraform
cd test-infrastructure/terraform
terraform init
terraform apply
```

## Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting (automatically done by pre-commit hook)
- Run `go vet` before committing (automatically done by pre-commit hook)
- Write tests for new functionality
- Update documentation when adding features

## Common Tasks

### Adding a New Feature

1. Create a new branch (optional but recommended):
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Test your changes:
   ```bash
   make test
   make run-dev
   ```

4. Commit and push:
   ```bash
   make git-push MSG="Add: your feature description"
   ```

### Fixing a Bug

1. Create a bug fix branch (optional):
   ```bash
   git checkout -b fix/bug-description
   ```

2. Fix the bug

3. Add tests to prevent regression

4. Commit and push:
   ```bash
   make git-push MSG="Fix: bug description"
   ```

### Updating Documentation

1. Edit the relevant documentation files

2. Commit and push:
   ```bash
   make git-push MSG="Docs: update documentation"
   ```

## Commit Message Conventions

Use clear, descriptive commit messages:

- `Add: new feature description`
- `Fix: bug description`
- `Update: what was updated`
- `Refactor: what was refactored`
- `Docs: documentation changes`
- `Test: test-related changes`
- `Chore: maintenance tasks`

## Troubleshooting

### Git Authentication Issues

If you encounter authentication issues when pushing:

1. **HTTPS**: Use a Personal Access Token
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Generate a new token with `repo` scope
   - Use the token as your password

2. **SSH**: Set up SSH keys
   ```bash
   ssh-keygen -t ed25519 -C "your-email@example.com"
   # Add the key to your GitHub account
   ```

### Build Issues

If you encounter build issues:

```bash
# Clean and rebuild
make clean
make deps
make build
```

### Module Issues

If you have Go module issues:

```bash
go clean -modcache
go mod download
go mod tidy
```

## Resources

- [Project README](../README.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Quick Start Guide](QUICKSTART.md)
- [GitHub Repository](https://github.com/Ruslanshtolik/hbf-agent)

## Getting Help

If you need help:
1. Check the documentation in the `docs/` directory
2. Review existing issues on GitHub
3. Create a new issue with detailed information about your problem
