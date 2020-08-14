# Lee-Key Services - Protect the Flag

An exercise in working with service accounts on Google Cloud and applying the 
principle of least privilege.


## The Story So Far ...
Lee-Key Services is experiencing cyber attacks. They handed out project owner 
roles like candy and now suffer the consequences. 

Web services appear to run as normal, but there are strong suspicions that 
malicious code has been injected into them. Unfortunately, the developers are 
out for lunch, barring any and all investigations in that direction.
Nonetheless, management has decided that all services must continue running as 
normal.

As the one are in charge of managing user and service account access privileges, 
you are the last bastion of hope. At least until the devs get back after lunch.


## Rules

* For all services, service accounts have been created.
* You may not create and reassign new service accounts.
* You can only work with adding and removing permissions.
* All normal functionalities should continue running.
* You may not alter code.
* You should not inspect the code (spoilers).


## Further Reading

* https://cloud.google.com/blog/products/identity-security/preventing-lateral-movement-in-google-compute-engine