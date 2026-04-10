DATABASE_URL ?= postgresql://sunjinlee@localhost:5432/vibecomposer_dev?sslmode=disable

.PHONY: run build dev db-up db-down migrate tidy docker-build docker-push tf-plan tf-apply deploy

# Run the server
run:
	go run ./cmd/server

# Build the binary
build:
	go build -o bin/server ./cmd/server

migrate-up:
	env DATABASE_URL=$(DATABASE_URL) go run ./cmd/migrate

# Run with auto-reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Tidy modules
tidy:
	go mod tidy

# --- Deployment ---

REGISTRY = asia-northeast3-docker.pkg.dev/gen-lang-client-0826771503/vibe-composer
IMAGE    = $(REGISTRY)/vibe-composer:latest
REGION   = asia-northeast3
SERVICE  = vibe-composer

# Build and push Docker image
docker-deploy:
	docker build --platform linux/amd64 -t $(IMAGE) .
	docker push $(IMAGE)

# Deploy new revision to Cloud Run (forces pull of latest image)
cloud-deploy:
	gcloud run services update $(SERVICE) --region=$(REGION) --image=$(IMAGE)

# Run migration job on Cloud Run
cloud-migrate:
	gcloud run jobs execute vibe-composer-migrate --region=$(REGION) --wait

# Terraform plan
tf-plan:
	cd infra && terraform plan

# Terraform apply (use for infra changes, not routine deploys)
tf-apply:
	cd infra && terraform apply

# Full deploy: build, push, update service, migrate
deploy: docker-deploy cloud-deploy cloud-migrate

