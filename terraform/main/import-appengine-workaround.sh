#!/usr/bin/env bash

set -euo pipefail

SCRIPTS="../scripts"
SECRETS="../secrets"
MAIN="../main"


function usage() {
	echo "${0} -p PROJECT"
	echo "Workaround to pull previously 'destroyed' AppEngine into current state."
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


terraform import \
	-var-file=etc/bootstrapped.tfvars \
	google_app_engine_application.app "${PROJECT}"

echo
