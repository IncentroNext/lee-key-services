#!/usr/bin/env bash

PROJECT=$1
if [ -z "${PROJECT}" ]; then
	echo "Missing project id parameter."
	exit 1
fi

RESP=$(gcloud run services list \
	--project="$PROJECT" \
	--platform=managed \
	--region=europe-west1)

echo "${RESP}" |\
	tail -n 1 |\
	awk '{ print $4 }' |\
	sed -E 's/.*-(\w+)-ew\.a\.run\.app/\1/'
