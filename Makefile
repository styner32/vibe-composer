.PHONY: run build dev db-up db-down migrate tidy docker-build docker-push tf-plan tf-apply deploy

# Run the server
run:
	go run ./cmd/server

# Build the binary
build:
	go build -o bin/server ./cmd/server

# Run database migrations
migrate:
	go run ./cmd/migrate

# Start local PostgreSQL
db-up:
	docker compose up -d db

# Stop local PostgreSQL
db-down:
	docker compose down

# Run with auto-reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Tidy modules
tidy:
	go mod tidy

# --- Deployment ---

REGISTRY = asia-northeast3-docker.pkg.dev/gen-lang-client-0826771503/vibe-composer
IMAGE    = $(REGISTRY)/vibe-composer:latest

# Build and push Docker image
docker-deploy:
	docker build -t $(IMAGE) .
	docker push $(IMAGE)

REGION   = asia-northeast3

# Terraform plan
tf-plan:
	cd infra && terraform plan

# Terraform apply
tf-apply:
	cd infra && terraform apply

# Run migration job on Cloud Run
cloud-migrate:
	gcloud run jobs execute vibe-composer-migrate --region=$(REGION) --wait

# Full deploy: build, push, apply, migrate
deploy: docker-deploy tf-apply cloud-migrate
