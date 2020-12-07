
resource "time_sleep" "service_enable_delay" {
  depends_on = [google_project_service.iam, google_project_service.run]

  create_duration = "60s"
}

resource "google_service_account" "dummy_service" {
  depends_on = [time_sleep.service_enable_delay]

  account_id   = "dummy-service"
}

resource "google_cloud_run_service" "dummy_service" {
  name = "dummy-service"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello"
      }
      service_account_name = google_service_account.dummy_service.email
    }
  }

  traffic {
    percent = 100
    latest_revision = true
  }
}
