#!/usr/bin/env bash

set -euo pipefail

NAME="terraform"
SCRIPTS="./"
CREDENTIALS="var/terraform.json"
OUTPUT="var/bootstrapped.tfvars"

function usage() {
	# TODO(hvl): defaults not correct if printed after parameter parsing
	echo "${0} -p PROJECT [-n NAME] [-s SCRIPTS] [-c CREDENTIALS] [-o OUTPUT]"
	echo "Bootstraps a project, snaring the elusive Cloud Run project hash."
	echo
	echo "-p PROJECT		Google Cloud Project ID"
	echo "-n NAME			Service account name (default: ${NAME})"
	echo "-s SCRIPTS		directory where manage-policy-file.py is located (default: ${SCRIPTS})"
	echo "-c CREDENTIALS	file to store Terraform service account key in (default: ${CREDENTIALS})"
	echo "-o OUTPUT			file to write output variables (default: ${OUTPUT})"
}

function prepare_dir() {
	_DIR=$(dirname "${1}")
	mkdir -p "${_DIR}"
}

function setup_service_account() {
	_PROJECT="$1"
	_NAME="$2"
	_SCRIPTS="$3"

	# Create Terraform service account
	gcloud iam service-accounts create "${_NAME}" \
		--project="${_PROJECT}"

	# Set authorisations on Terraform service account
	_TMP_DIR="tmp-$(uuidgen)"
	_POLICY_FILE="${_TMP_DIR}/policy-file"

	mkdir -p "${_TMP_DIR}"

	gcloud projects get-iam-policy "${_PROJECT}" \
		--format json > "${_POLICY_FILE}"
	"${_SCRIPTS}/manage-policy-file.py" -i "${_POLICY_FILE}" \
		-e "${_NAME}@${_PROJECT}.iam.gserviceaccount.com" -t serviceAccount \
		-a add -r "roles/editor"
	gcloud projects set-iam-policy "${_PROJECT}" "${_POLICY_FILE}"

	rm -rf "${_TMP_DIR}"
}

PROJECT=

while getopts "p:n:c:s:o:" o; do
	case "${o}" in
	p)
		PROJECT="${OPTARG}"
		;;
	n)
		NAME="${OPTARG}"
		;;
	c)
		CREDENTIALS="${OPTARG}"
		;;
	s)
		SCRIPTS="${OPTARG}"
		;;
	o)
		OUTPUT="${OPTARG}"
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

if [ -f "${CREDENTIALS}" ]; then
	usage
	echo "Credentials file already exists, will not overwrite ${CREDENTIALS}"
	exit 1
fi

if [ ! -f "${SCRIPTS}/manage-policy-file.py" ]; then
	usage
	echo "Could not find manage-policy-file.py in ${SCRIPTS}"
	exit 1
fi

if [ -f "${OUTPUT}" ]; then
	usage
	echo "Output file already exists, will not overwrite ${OUTPUT}"
	exit 1
fi

exit

# Create Terraform service account
setup_service_account "${PROJECT}" "${NAME}" "${SCRIPTS}"
SERVICE_ACCOUNT="${NAME}@${PROJECT}.iam.gserviceaccount.com"

# Get credentials file for Terraform service account
prepare_dir "${CREDENTIALS}"
gcloud iam service-accounts keys create "${CREDENTIALS}" \
	--iam-account="${SERVICE_ACCOUNT}" \
	--project="${PROJECT}"

# Initialise project
terraform init
terraform apply -auto-approve \
	-var project="${PROJECT}" \
	-var secrets_file="${CREDENTIALS}"

# Store variables from bootstrap
prepare_dir "${OUTPUT}"
terraform output > "${OUTPUT}"

# Clean up dummy service
# This will leave this module in a transitive state. That's fine.
terraform destroy -auto-approve \
	-var project="${PROJECT}" \
	-var secrets_file="${CREDENTIALS}" \
	-target="google_service_account.dummy_service" \
	-target="google_cloud_run_service.dummy_service"

echo "Bootstrap done; variables have been written to ${OUTPUT}."
