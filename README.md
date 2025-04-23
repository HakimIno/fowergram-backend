# Fowergram Backend API

A robust and scalable backend API for the Fowergram social media platform, built with Go and deployed on Google Kubernetes Engine (GKE).

## Features

- User authentication and authorization
- Rate limiting and security measures
- Health monitoring endpoints
- Kubernetes-ready deployment
- Comprehensive API documentation
- Environment-based configuration

## Tech Stack

- **Language**: Go 1.21
- **Framework**: Fiber
- **Database**: PostgreSQL
- **Cache**: Redis
- **Container**: Docker
- **Orchestration**: Kubernetes (GKE)
- **CI/CD**: GitHub Actions

## Documentation

- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Environment Variables](docs/environment-variables.md)

## Prerequisites

- Go 1.21 or later
- Docker
- kubectl
- Access to GCP and GKE
- PostgreSQL
- Redis

## Local Development

1. Clone the repository:
```bash
git clone https://github.com/yourusername/fowergram-be.git
cd fowergram-be
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the application:
```bash
go run cmd/api/main.go
```

5. Run with hot reload (recommended for development):
```bash
# Install Air (if not already installed)
go install github.com/cosmtrek/air@latest

# Run with hot reload using make
make dev

# Or run Air directly
air
```

6. Run tests:
```bash
go test -v ./...
```

## Deployment

See [Deployment Documentation](docs/deployment.md) for detailed instructions.

Quick start:
```bash
# Set up GCP credentials
gcloud auth login
gcloud config set project fowergram-backend

# Deploy to GKE
kubectl apply -f k8s/base/
```

## Testing

### Unit Tests
```bash
go test -v ./...
```

### Integration Tests
```bash
go test -v ./tests/integration/...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

Project Link: [https://github.com/yourusername/fowergram-be](https://github.com/yourusername/fowergram-be)
