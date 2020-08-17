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


## Setup

To set up a sandbox project you will need the following:

* Project Owner role on a (preferably blank) Google Cloud project
* Google Cloud SDK
* Linux or MacOS command shell (tested with Ubuntu / Bash)
* GNU Make


## Rules

* Each service has its own service account, these have been already created.
* You may (or need) not create (or reassign) service accounts.
* The service accounts already have (poorly selected) permissions.
* All normal process flows should remain functional:
    * creating new orders
    * creating payments for orders
* You can only work by adding and removing permissions, by any means: the web 
  console, command line (gcloud) or (REST) APIs.
* You may (or need) not redeploy services.
* You should not inspect the code - this decreases learning potential. 
* You may not alter code.


## Interesting Related Reads

* https://cloud.google.com/blog/products/identity-security/preventing-lateral-movement-in-google-compute-engine
