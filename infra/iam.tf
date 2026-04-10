# -------------------------------------------------
# Service Account for Cloud Run
# -------------------------------------------------
resource "google_service_account" "cloud_run" {
  account_id   = "${var.app_name}-run"
  display_name = "Vibe Composer Cloud Run SA"
  project      = var.project_id
}

# Cloud SQL client access
resource "google_project_iam_member" "cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# GCS read/write access
resource "google_storage_bucket_iam_member" "media_user" {
  bucket = google_storage_bucket.media.name
  role   = "roles/storage.objectUser"
  member = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Secret Manager access
resource "google_secret_manager_secret_iam_member" "gemini_key_access" {
  secret_id = google_secret_manager_secret.gemini_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_secret_manager_secret_iam_member" "auth_password_access" {
  secret_id = google_secret_manager_secret.auth_password.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_secret_manager_secret_iam_member" "db_password_access" {
  secret_id = google_secret_manager_secret.db_password.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

# -------------------------------------------------
# Artifact Registry Repository
# -------------------------------------------------
resource "google_artifact_registry_repository" "app" {
  location      = var.region
  repository_id = var.app_name
  format        = "DOCKER"
  description   = "Docker images for Vibe Composer"

  depends_on = [google_project_service.apis["artifactregistry.googleapis.com"]]
}

# Artifact Registry read access (pull images)
resource "google_artifact_registry_repository_iam_member" "cloud_run_reader" {
  project    = var.project_id
  location   = var.region
  repository = google_artifact_registry_repository.app.repository_id
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.cloud_run.email}"
}
