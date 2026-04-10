locals {
  image_url    = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.app.repository_id}/${var.app_name}:${var.image_tag}"
  database_url = "postgres://${google_sql_user.app.name}:${random_password.db_password.result}@/${var.cloud_sql_db_name}?host=/cloudsql/${google_sql_database_instance.main.connection_name}"
}

# -------------------------------------------------
# Cloud Run v2 Service — Main App
# -------------------------------------------------
resource "google_cloud_run_v2_service" "app" {
  name     = var.app_name
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.cloud_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 3
    }

    vpc_access {
      connector = google_vpc_access_connector.main.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.main.connection_name]
      }
    }

    containers {
      image = local.image_url

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      # --- Environment Variables ---
      env {
        name  = "DATABASE_URL"
        value = local.database_url
      }

      env {
        name  = "GCS_BUCKET"
        value = google_storage_bucket.media.name
      }

      env {
        name  = "ALLOWED_USERS"
        value = var.allowed_users
      }

      # --- Secrets ---
      env {
        name = "GOOGLE_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.gemini_api_key.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "AUTH_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.auth_password.secret_id
            version = "latest"
          }
        }
      }

      # Mount Cloud SQL socket
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      # Startup probe
      startup_probe {
        http_get {
          path = "/"
          port = 8080
        }
        initial_delay_seconds = 5
        period_seconds        = 10
        failure_threshold     = 3
      }
    }
  }

  depends_on = [
    google_project_service.apis["run.googleapis.com"],
    google_secret_manager_secret_version.gemini_api_key,
    google_secret_manager_secret_version.auth_password,
  ]
}

# Allow unauthenticated access (the app has its own basic auth)
resource "google_cloud_run_v2_service_iam_member" "public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.app.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# -------------------------------------------------
# Cloud Run v2 Job — Database Migration
# -------------------------------------------------
resource "google_cloud_run_v2_job" "migrate" {
  name     = "${var.app_name}-migrate"
  location = var.region

  template {
    template {
      service_account = google_service_account.cloud_run.email

      vpc_access {
        connector = google_vpc_access_connector.main.id
        egress    = "PRIVATE_RANGES_ONLY"
      }

      volumes {
        name = "cloudsql"
        cloud_sql_instance {
          instances = [google_sql_database_instance.main.connection_name]
        }
      }

      containers {
        image   = local.image_url
        command = ["migrate"]

        resources {
          limits = {
            cpu    = "1"
            memory = "512Mi"
          }
        }

        env {
          name  = "DATABASE_URL"
          value = local.database_url
        }

        volume_mounts {
          name       = "cloudsql"
          mount_path = "/cloudsql"
        }
      }
    }
  }

  depends_on = [
    google_project_service.apis["run.googleapis.com"],
  ]
}
