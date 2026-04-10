# -------------------------------------------------
# GCS Bucket for audio inputs and generated music
# -------------------------------------------------
resource "google_storage_bucket" "media" {
  name     = "${var.project_id}-${var.app_name}-media"
  location = var.region

  uniform_bucket_level_access = true
  force_destroy               = true # Set to false for production

  versioning {
    enabled = false
  }

  lifecycle_rule {
    condition {
      age = 90 # Clean up old files after 90 days
    }
    action {
      type = "Delete"
    }
  }

  cors {
    origin          = ["*"]
    method          = ["GET", "PUT", "POST"]
    response_header = ["Content-Type"]
    max_age_seconds = 3600
  }
}
