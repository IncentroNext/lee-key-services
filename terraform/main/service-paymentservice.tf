
resource "google_storage_bucket" "invoices-bucket" {

  name = "${var.project}-leekeyservices-invoices"
  location = var.region
  force_destroy = true

  uniform_bucket_level_access = true
}

resource "google_service_account" "paymentservice_v1" {
  depends_on = [time_sleep.iam_service_enable_delay]

  account_id   = "paymentservice-v1"
}

resource "google_cloud_run_service" "paymentservice_v1" {

  name = "paymentservice-v1"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project}/paymentservice_v1"

        dynamic "env" {
          for_each = local.envs
          content {
            name = env.key
            value = env.value
          }
        }
      }
      service_account_name = google_service_account.paymentservice_v1.email
    }
  }

  traffic {
    percent = 100
    latest_revision = true
  }
}

resource "google_cloud_run_service_iam_policy" "paymentservice_v1" {
  
  location = google_cloud_run_service.paymentservice_v1.location
  project = google_cloud_run_service.paymentservice_v1.project
  service = google_cloud_run_service.paymentservice_v1.name

  policy_data = data.google_iam_policy.allUsers.policy_data
}
