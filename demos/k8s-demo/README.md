# Advanced Introduction to Secretless in Kubernetes

This tutorial will get Secretless running quickly in Kubernetes.  

It's intended for advanced users and is light on explanations.

For a friendlier version of this tutorial, with breakdowns of every step, try:

[Our Detailed Introduction to Secretless in Kubernetes](https://secretless.io/docs/get_started/kubernetes_tutorial.html)

## Overview

Here's what we'll do:

1. Deploy a PostgreSQL database
2. Store its credentials in Kubernetes secrets
3. Setup Secretless Broker to proxy connections to it 
4. Deploy a sample application that connects to the database **without knowing
   its password**

You'll play two roles in this tutorial:

1. A **Security Admin** who handles secrets, and has sole access to those secrets
2. An **Application Developer** with no access to secrets.

## Prerequisites

+ A running Kubernetes cluster (you can use
  [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a
  cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured
  to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

## Steps for Security Admin

As security admin, you will:

1. Create a PostgreSQL database
1. Create a DB user for the application
1. Add that user's credentials to Kubernetes Secrets
1. Configure Secretless to connect to PostgreSQL using those credentials

To perform all these steps in one go, run:

```bash
./01_security_admin_steps
```

<p></p>
<details>
  <summary>View expected output</summary>
  <pre>
>>--- Create a new namespace
Cleaning up old namespace............Done
namespace "quick-start-backend-ns" created
>>--- Add certificates to Kubernetes Secrets
secret "quick-start-backend-certs" created
>>--- Create StatefulSet for Database
statefulset "pg" created
service "quick-start-backend" created
Waiting for quick-start-backend to be ready........Done
>>--- Create Application Database
CREATE DATABASE
>>--- Create Database Table and Permissions
CREATE ROLE
CREATE TABLE
GRANT
GRANT
>>--- Store DB credentials in Kubernetes Secrets
Cleaning up old namespace....................Done
namespace "quick-start-application-ns" created
secret "quick-start-backend-credentials" created
>>--- Create Application Service Account
serviceaccount "quick-start-application" created
role "quick-start-backend-credentials-reader" created
rolebinding "read-quick-start-backend-credentials" created
>>--- Create and Store Secretless Configuration
configmap "quick-start-application-secretless-config" created
  </pre>
</details>
<p></p>


## Steps for Application Developer

**Important: The application developer never needs any credentials to connect
to the database.**

In particular, as the application developer you do not know any of the secrets
in **security_admin_secrets.sh**.

**As the application developer:**                                                                                                                                       

1. Configure the application to connect to PostgreSQL via Secretless                                                                                                    
1. Deploy the application and the Secretless sidecar                                                                                                                    
1. Test the application

To perform all these steps in one go, run:

```bash
./02_app_developer_steps
```

<p></p>
<details>
  <summary>View expected output</summary>
  <pre>
>>--- Start application
deployment "quick-start-application" created
service "quick-start-application" created
Waiting for application to boot up
(This may take more than 1 minute)
............................Done
>>--- Add a Sample Pet
HTTP/1.1 201 
Location: http://192.168.99.100:30002/pet/1
Content-Length: 0
Date: Thu, 14 Mar 2019 15:35:33 GMT
>>--- Retrieve All Pets
HTTP/1.1 200 
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 14 Mar 2019 15:35:33 GMT
[{"id":1,"name":"Mr. Snuggles"}]
All finished!  Secretless is working!
  </pre>
</details>
<p></p>

That's it!

The application is connecting to a password-protected Postgres database
**without any knowledge of the credentials**.

## Conclusion

Please let us know what you think of Secretless! You can submit [Github
issues](https://github.com/cyberark/secretless-broker/issues) for features
you'd like to see, or send a message to our [mailing
list](https://groups.google.com/forum/#!forum/secretless) with comments or
questions.
