# -------------------------------------------------
# Secret Manager — Gemini API Key
# -------------------------------------------------
resource "google_secret_manager_secret" "gemini_api_key" {
  secret_id = "${var.app_name}-gemini-api-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis["secretmanager.googleapis.com"]]
}

resource "google_secret_manager_secret_version" "gemini_api_key" {
  secret      = google_secret_manager_secret.gemini_api_key.id
  secret_data = var.gemini_api_key
}

# -------------------------------------------------
# Secret Manager — Auth Password
# -------------------------------------------------
resource "google_secret_manager_secret" "auth_password" {
  secret_id = "${var.app_name}-auth-password"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis["secretmanager.googleapis.com"]]
}

resource "google_secret_manager_secret_version" "auth_password" {
  secret      = google_secret_manager_secret.auth_password.id
  secret_data = var.auth_password
}
