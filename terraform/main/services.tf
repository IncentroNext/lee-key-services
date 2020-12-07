
resource "google_project_service" "iam" {
  project = var.project
  service = "iam.googleapis.com"
}

resource "google_project_service" "cloudresourcemanager" {
  project = var.project
  service = "cloudresourcemanager.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy = true
}

resource "google_project_service" "run" {
  project = var.project
  service = "run.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy = true
}

resource "google_project_service" "containerregistry" {
  project = var.project
  service = "containerregistry.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy = true
}

resource "google_project_service" "cloudbuild" {
  project = var.project
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy = true
}

resource "time_sleep" "iam_service_enable_delay" {
  depends_on = [google_project_service.iam]

  create_duration = "60s"
}
