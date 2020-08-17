
REGION := europe-west1

COLLECTIONS := orders,secrets
BUCKET1 := $(PROJECT)-leekeyservices-invoices
BUCKET2 := $(PROJECT)-leekeyservices-payments


.PHONY: init-project

check:
ifndef PROJECT
	$(error PROJECT not set)
endif

# Setup Step 0: Build once
services-build:
	make -j -C printservice/v1 build
	make -j -C numberservice/v1 build
	make -j -C orderservice/v1 build
	make -j -C paymentservice/v1 build
	make -j -C website/v1 build
	make -j -C intruder/v1 build

# Setup Step 1: Repeat for each project
init-project: check
	$(info Enable Container Registry:)
	$(info https://console.cloud.google.com/apis/library/containerregistry.googleapis.com?project=$(PROJECT))
	$(info Enable Cloud Run:)
	$(info https://console.cloud.google.com/apis/library/run.googleapis.com?project=$(PROJECT))
	$(info Firestore:)
	$(info https://console.cloud.google.com/firestore/data?project=$(PROJECT))

# Setup Step 2: Repeat for each project
project-setup: services-no-build create-buckets import-firestore

# The project hash only becomes known once the first service ahs been deployed.
# So first we deploy a service that does not need URLs of any other service.
services-no-build: check
	make -j -C printservice/v1 no-build
	make -j -C numberservice/v1 no-build
	make -j -C orderservice/v1 no-build
	make -j -C paymentservice/v1 no-build
	make -j -C website/v1 no-build
	make -j -C intruder/v1 no-build

create-buckets: check
	gsutil mb \
		-p $(PROJECT) \
		-l $(REGION) \
		-b on \
		gs://$(BUCKET1)
	gsutil mb \
		-p $(PROJECT) \
		-l $(REGION) \
		-b on \
		gs://$(BUCKET2)
	mkdir -p tmp
	echo "i am a happy little flag" >> tmp/happy-little-file.txt
	gsutil cp tmp/happy-little-file.txt gs://$(BUCKET2)
	rm tmp/happy-little-file.txt

import-firestore: check
	-gsutil -m rm -rf gs://$(BUCKET2)/orders-export
	gsutil -m cp -r orders-export gs://$(BUCKET2)
	gcloud firestore import gs://$(BUCKET2)/orders-export \
		--project=$(PROJECT) \
		--collection-ids=$(COLLECTIONS)

# See comment on services-no-build
services: check
	make -C printservice/v1 all
	make -C numberservice/v1 all
	make -C orderservice/v1 all
	make -C paymentservice/v1 all
	make -C website/v1 all
	make -C intruder/v1 all

# Or just delete the project altogether (by hand).
destroy: check
	make -C intruder/v1 destroy
	make -C website/v1 destroy
	make -C orderservice/v1 destroy
	make -C paymentservice/v1 destroy
	make -C numberservice/v1 destroy
	make -C printservice/v1 destroy
	gsutil rb gs://$(BUCKET1)
	gsutil rb gs://$(BUCKET2)
	$(info Remove $(COLLECTIONS) manually from Firestore)

# For development
export-orders: check
	-rm -rf ./orders-export
	-gsutil -m rm -rf gs://$(BUCKET2)/orders-export
	gcloud firestore export \
		gs://$(BUCKET2)/orders-export \
		--project=$(PROJECT) \
		--collection-ids='$(COLLECTIONS)'
	gsutil -m cp -r gs://$(BUCKET2)/orders-export .
	gsutil -m rm -rf gs://$(BUCKET2)/orders-export
