# Lee-Key Services - Protect the Flag

An exercise in working with service accounts on Google Cloud and applying the 
principle of least privilege.


## The Story So Far ...
Lee-Key Services is experiencing cyber attacks. They handed out project owner 
roles like candy and now suffer the consequences. 

Web services appear to run as normal, but there are strong suspicions that 
malicious code has been injected into them. Unfortunately, all the developers 
are out for lunch, so the code can neither be investigated nor patched. Against
better judgment, management has decided that there can be no downtime.

As the one are in charge of managing user and service account access privileges, 
you are the last bastion of hope. At least until the devs get back from lunch.


## Rules

The goal of the games is to score as little points as possible via the Attacker
Portal without disrupting normal functions.

* All normal process flows should remain functional:
    * creating new orders
    * creating payments for orders
* Each service has its own service account, these have been already created.
* The service accounts already have (poorly selected) permissions.
* You can only work by adding and removing permissions, by any means: the web 
  console, command line (gcloud) or (REST) APIs.
* You may (or need) not create (or reassign) service accounts.
* You may (or need) not redeploy services.
* You should not inspect the code - this decreases learning potential. 
* You may not alter code.


## Setup

### Pre-requisites

* Project Owner role on a (preferably blank) Google Cloud project
* Google Cloud SDK (gcloud, gsutil, ...)
* Linux or MacOS command shell (tested with Ubuntu / Bash)
* GNU Make
* Docker

### Project Setup

At some point this project might use Cloud Build, or it might not. For the 
foreseeable future, everything is handled by locally run scripts. 

If you have never used Docker in combination with Google Container Registry, you
first need to run the following command (as per https://cloud.google.com/container-registry/docs/pushing-and-pulling)
```shell script
gcloud auth configure-docker
```

The application code and project setup scripts are located in the `services`
directory. From there, execute and follow the instructions provided the 
following commands:

1. `make services-build` (builds docker images) (only once or after code change)
2. `make init-project PROJECT=<project id>` (lists links of Google APIs to activate)
3. `make project-setup PROJECT=<project id>` (deploys services, sets up data)



## Interesting Related Reads

* https://cloud.google.com/blog/products/identity-security/preventing-lateral-movement-in-google-compute-engine


## License
Copyright 2020 Incentro

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
