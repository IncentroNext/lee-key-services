// provider variables and settings

variable "project" {
}

variable "secrets_file" {
}

variable "region" {

  default = "europe-west1"
}

variable "zone" {

  default = "europe-west1-b"
}

provider "google" {

  credentials = file(var.secrets_file)

  project = var.project
  region  = var.region
  zone    = var.zone
}
