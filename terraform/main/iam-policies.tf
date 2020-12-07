
data "google_iam_policy" "allUsers" {

  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}
