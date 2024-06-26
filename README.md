# Cohort-middleware

![version](https://img.shields.io/github/release/uc-cdis/cohort-middleware.svg)
[![Apache license](http://img.shields.io/badge/license-Apache-blue.svg?style=flat)](LICENSE)
[![GitHub Actions](https://github.com/uc-cdis/cohort-middleware/workflows/Build%20Image%20and%20Push%20to%20Quay/badge.svg)](https://github.com/uc-cdis/cohort-middleware/actions)
[![Coverage Status](https://coveralls.io/repos/github/uc-cdis/cohort-middleware/badge.svg?branch=master)](https://coveralls.io/github/uc-cdis/cohort-middleware?branch=master)

Cohort-middleware provides a set of web-services (endpoints) for:

1. providing information about cohorts to which a user has authorized access (Atlas DB cohorts as defined in Fence/Arborist?)
2. getting clinical attribute values for a given cohort (aka CONCEPT values in Atlas/OMOP jargon)
3. providing patient-level clinical attribute values matrix for use in backend workflows, like GWAS workflows (e.g. https://github.com/uc-cdis/vadc-genesis-cwl)

The cohorts and their clinical attribute values are retrieved from
connected OHDSI/CMD/Atlas databases via SQL queries.

## Table of Content

- [Cohort-middleware](#cohort-middleware)
  - [Table of Content](#table-of-content)
  - [API Documentation](#api-documentation)
  - [Overview diagram](#overview-diagram)
  - [Running](#running)
    - [Config file](#config-file)
    - [DB schemas](#db-schemas)
      - [Setting up databases for local development](#setting-up-databases-for-local-development)
- [Deployment steps](#deployment-steps)
  - [Deployment to Gen3](#deployment-to-gen3)
  - [Test the endpoints](#test-the-endpoints)
    - [Troubleshooting](#troubleshooting)
      - [How to make curl with Auth](#how-to-make-curl-with-auth)
      - [How to see the logs](#how-to-see-the-logs)
      - [In case of infra / network issues:](#in-case-of-infra--network-issues)
      - [Updating the Docker base image](#updating-the-docker-base-image)

## API Documentation

[OpenAPI documentation available here.](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/uc-cdis/cohort-middleware/master/openapis/swagger.yaml)

YAML file for the OpenAPI documentation is found in the `openapis` folder.

## Overview diagram

Overview of cohort-middleware and its connected systems:

<div align="center">
<img src="./docs/cohort-middleware-overview.png" alt="Cohort-middleware and connected systems overview" height="400" hspace="10"/>
</div>


## Running

Execute the following command to get help:

```
go run main.go -h
```

To just start with the default "development" settings:
```
go run main.go
```


### Config file

See example config file in `./config/` folder.

### DB schemas

The data which our code queries is currently assuming 2 separate databases.
The "atlas" schema on one database, and the "results" and "cdm" schemas
together on another DB. In practice, the databases could even be a mix from
different vendors/engines (e.g. one a "sql server" and one a "postgres").
Therefore, the code does not have queries that do a direct join between
tables in "atlas" and "results" or "atlas" and "cdm".

Below is an overview of the schemas and respective tables.

**DB Instance1**:
```
===============================
       SCHEMA atlas
===============================
TABLE atlas.source
TABLE atlas.source_daimon
TABLE atlas.cohort_definition
```

**DB Instance2**:

```
===============================
      SCHEMA results
===============================
TABLE results.COHORT

===============================
      SCHEMA omop
===============================
TABLE omop.person
TABLE omop.observation
TABLE omop.concept
VIEW omop.observation_continuous
```


#### Setting up databases for local development

Setup the local Atlas DB by running the `init_db.sh` script in the `./tests` folder:

```
cd tests/setup_local_db/
./init_db.sh
```

**Test this setup by trying the following curl commands**:
JSON summary data endpoints:
```bash
curl http://localhost:8080/sources | python -m json.tool
curl "http://localhost:8080/cohortdefinition-stats/by-source-id/1/by-team-project?team-project=test" | python -m json.tool
curl http://localhost:8080/concept/by-source-id/1 | python -m json.tool
curl -d '{"ConceptIds":[2000000324,2000006885]}' -H "Content-Type: application/json" -X POST http://localhost:8080/concept/by-source-id/1 | python -m json.tool
curl -d '{"ConceptTypes":["Measurement","Person"]}' -H "Content-Type: application/json" -X POST http://localhost:8080/concept/by-source-id/1/by-type | python -m json.tool

curl http://localhost:8080/concept-stats/by-source-id/1/by-cohort-definition-id/3/breakdown-by-concept-id/2000007027 | python3 -m json.tool

curl -d '{"variables": [{"variable_type": "concept", "concept_id": 2000006885}]}' -H "Content-Type: application/json" -X POST http://localhost:8080/concept-stats/by-source-id/1/by-cohort-definition-id/3/breakdown-by-concept-id/2000007027 | python3 -m json.tool
```

CSV data endpoints:
```bash
curl -d '{"variables":[{"variable_type": "concept", "concept_id": 2000000324},{"variable_type": "concept", "concept_id": 2000006885},{"variable_type": "concept", "concept_id": 2000007027},{"variable_type": "custom_dichotomous", "cohort_ids": [1, 2]}]}' -H "Content-Type: application/json" -X POST http://localhost:8080/cohort-data/by-source-id/1/by-cohort-definition-id/3

curl -d '{"variables":[{"variable_type": "concept", "concept_id": 2000000324},{"variable_type": "concept", "concept_id": 2000006885},{"variable_type": "concept", "concept_id": 2000007027},{"variable_type": "custom_dichotomous", "provided_name": "test123", "cohort_ids": [1, 99]}]}' -H "Content-Type: application/json" -X POST http://localhost:8080/concept-stats/by-source-id/1/by-cohort-definition-id/3/breakdown-by-concept-id/2000007027/csv

```

Histogram endpoint:
```bash
curl -d '{"variables":[{"variable_type": "custom_dichotomous", "cohort_ids": [1, 4]}]}' -H "Content-Type: application/json" -X POST http://localhost:8080/histogram/by-source-id/1/by-cohort-definition-id/4/by-histogram-concept-id/2000006885
```

# Deployment steps

## Deployment to Gen3

For deployment in Gen3 simply use [`kube-setup-cohort-middleware`](https://github.com/uc-cdis/cloud-automation/blob/master/gen3/bin/kube-setup-cohort-middleware.sh) script:

```
gen3 kube-setup-cohort-middleware
```

The script will use `ohdsi` database credentials and will result in `cohort-middleware-g3auto` Kubernetes secret.

If any changes need to be made to the settings, find the .yaml config file:

```
find ~ -type f -path '*/g3auto/cohort-middleware/*.yaml'
```
and **remove** that first before running the `gen3 kube-setup` command above.

### Roll cohort-middleware

To roll cohort-middleware (in case of version update), full `kube-setup-cohort-middleware` is not required:

```
gen3 roll cohort-middleware
```

This will take care of all the secrets via [`g3auto`](https://github.com/uc-cdis/cloud-automation/blob/master/doc/secrets.md#overview).

## Test the endpoints

Example:
```
curl -H "Content-Type: application/json" -H "$(cat auth.txt)" https://<server-url-here>/sources | python -m json.tool
```

**Note that** the `<server-url-here>` in these examples above needs to be replaced, and the ids used (`by-source-id/2`, `by-cohort-definition-id/3`) need
to be replaced with real values from your environment. The main addition in these `curl` commands is the presence of `https` and the
extra `-H "$(cat auth.txt)"`. More explained in the subsections below.

### Troubleshooting

Troubleshooting steps when using manifest.json based deployment:

   - check `/home/<your-machine-name>/cdis-manifest/<your-machine-name>/manifest.json` to make sure the desired image name and tag for cohort-middleware
   are present. Do _not_ edit this file directly on the server, but make a PR with changes if needed.
   - regarding `gen3 roll cohort-middleware`, see also https://github.com/uc-cdis/cloud-automation/blob/master/kube/services/cohort-middleware/cohort-middleware-deploy.yaml,
   which is used directly by the `gen3 roll` command (see https://github.com/uc-cdis/cloud-automation/blob/master/gen3/bin/roll.sh).

#### How to make curl with Auth

Go to https://<server-url-here> and then to "Login"->"Profile"->"Create API key". Download the JSON to your local computer.

Run (e.g. if the downloaded JSON file is called `credentials.json`):
```bash
export SERVER_NAME=<your-server-name-here>
curl -d "$(cat credentials.json)" -X POST -H "Content-Type: application/json" https://${SERVER_NAME}/user/credentials/api/access_token
```
Save the contents of token in a file, e.g. `auth.txt`. Then try for example:
```bash
curl -H "Content-Type: application/json" -H "Authorization: bearer $(cat auth.txt)" https://${SERVER_NAME}/cohort-middleware/sources | python -m json.tool
```

#### How to see the logs

Find the pod(s):

```
kubectl get pods --all-namespaces | grep cohort-middleware
```

or:
```
kubectl get pods -l app=cohort-middleware
```

Then run:

```
kubectl logs <pod-name-here>

or

kubectl logs -f -l app=cohort-middleware
```

See also https://kubernetes.io/docs/reference/kubectl/cheatsheet/#interacting-with-running-pods


#### In case of infra / network issues:

Get help from **"PE team"**:
- PE team = Platform Engineering team = [GPE project Jira ticket](https://ctds-planx.atlassian.net/browse/GPE) = #gen3-devops-oncall (slack channel)

If networking changes are necessary:
- see https://github.com/uc-cdis/cloud-automation/blob/master/gen3/bin/kube-setup-networkpolicy.sh

If proxy changes are necessary:
- see https://github.com/uc-cdis/cloud-automation/tree/master/kube/services/revproxy/gen3.nginx.conf

Other config related to network policies:
- https://github.com/uc-cdis/cloud-automation/blob/master/kube/services/cohort-middleware/cohort-middleware-deploy.yaml

#### Updating the Docker base image

To push a new **generic** dockerhub image to Quay (like a specific version of Golang), use something like this in slack:

```
@qa-bot run-jenkins-job gen3-self-service-push-dockerhub-img-to-quay jenkins {"SOURCE":"python:3.10-alpine","TARGET":"quay.io/cdis/python:3.10-alpine-master"}
```

Or use the self-service page:
- https://jenkins.planx-pla.net/job/gen3-self-service-push-dockerhub-img-to-quay/build

The result will be a new image pushed to quay.io that we can start using in our Dockerfile, like:

```
FROM quay.io/cdis/golang:1.18-bullseye
```
