
variable "app_region" {

  default = "europe-west"
}

resource "google_project_service" "appengine" {
  project = var.project
  service = "appengine.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy = true
}

resource "time_sleep" "service_enable_delay" {
  depends_on = [google_project_service.appengine]

  create_duration = "60s"
}

resource "google_app_engine_application" "app" {
  depends_on = [time_sleep.service_enable_delay]

  project     = var.project
  location_id = var.app_region
  database_type = "CLOUD_FIRESTORE"
}
