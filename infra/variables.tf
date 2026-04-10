variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region for all resources"
  type        = string
  default     = "us-central1"
}

variable "app_name" {
  description = "Application name used for resource naming"
  type        = string
  default     = "vibe-composer"
}

variable "gemini_api_key" {
  description = "Google Gemini API key for Lyria and Flash"
  type        = string
  sensitive   = true
}

variable "auth_password" {
  description = "Shared password for basic auth"
  type        = string
  sensitive   = true
  default     = "vibecheck"
}

variable "allowed_users" {
  description = "Comma-separated list of allowed usernames"
  type        = string
  default     = "sunjin"
}

variable "cloud_sql_tier" {
  description = "Cloud SQL machine tier"
  type        = string
  default     = "db-f1-micro"
}

variable "cloud_sql_db_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "vibecomposer"
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "latest"
}
