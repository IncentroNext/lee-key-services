#!/usr/bin/env bash

set -euo pipefail

SCRIPTS="../scripts"
SECRETS="../secrets"
MAIN="../main"


function usage() {
	echo "${0} -p PROJECT"
	echo "Bootstraps a project, snaring the elusive Cloud Run project hash."
	echo
	echo "-p PROJECT	Google Cloud Project ID"
}


PROJECT=

while getopts "p:" o; do
	case "${o}" in
	p)
		PROJECT="${OPTARG}"
		;;
	*)
		usage
		exit 1
	esac
done

if [ -z "${PROJECT}" ]; then
	usage
	echo "missing -p <project id>"
	exit 1
fi

SERVICE_ACCOUNT="terraform@${PROJECT}.iam.gserviceaccount.com"

# Create Terraform service account
gcloud iam service-accounts create terraform \
	--project="${PROJECT}"

# Set authorisations on Terraform service account
mkdir -p tmp
POLICY_FILE=tmp/policy-file
gcloud projects get-iam-policy hayovanloon-terraform-7199127a \
	--format json > "${POLICY_FILE}"
"${SCRIPTS}"/manage-policy-file.py -i "${POLICY_FILE}" \
	-e "${SERVICE_ACCOUNT}" -t serviceAccount \
	-a add -r "roles/editor"
gcloud projects set-iam-policy "${PROJECT}" "${POLICY_FILE}"
rm -rf tmp/

# Get credentials file for Terraform service account
mkdir -p "${SECRETS}"
gcloud iam service-accounts keys create "${SECRETS}/terraform.json" \
	--iam-account="${SERVICE_ACCOUNT}" \
	--project="${PROJECT}"

# Initialise project
terraform init
terraform apply -auto-approve \
	-var project="${PROJECT}" \
	-var secrets_file="${SECRETS}/terraform.json"

# Store variables from bootstrap
mkdir -p "${MAIN}/etc"
terraform output > "${MAIN}/etc/bootstrapped.tfvars"

# Clean up dummy service
# This will leave this module in a transitive state. That's fine.
terraform destroy -auto-approve \
	-var project="${PROJECT}" \
	-var secrets_file="${SECRETS}/terraform.json" \
	-target="google_service_account.dummy_service" \
	-target="google_cloud_run_service.dummy_service"

echo
