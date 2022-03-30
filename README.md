# Cohort-middleware

![version](https://img.shields.io/github/release/uc-cdis/cohort-middleware.svg) [![Apache license](http://img.shields.io/badge/license-Apache-blue.svg?style=flat)](LICENSE) [![Travis](https://travis-ci.org/uc-cdis/cohort-middleware.svg?branch=master)](https://travis-ci.org/uc-cdis/cohort-middleware) [![Coverage Status](https://coveralls.io/repos/github/uc-cdis/cohort-middleware/badge.svg?branch=master)](https://coveralls.io/github/uc-cdis/cohort-middleware?branch=master)

Cohort-middleware provides a set of web-services (endpoints) for:

1. providing information about cohorts to which a user has authorized access (Atlas DB cohorts as defined in Fence/Arborist?)
2. getting clinical attribute values for a given cohort (aka CONCEPT values in Atlas/OMOP jargon)
3. providing patient-level clinical attribute values matrix for use in backend workflows, like GWAS workflows (e.g. https://github.com/uc-cdis/vadc-genesis-cwl)

The cohorts and their clinical attribute values are retrieved from
connected OHDSI/CMD/Atlas databases via SQL queries.

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
TABLE omop.domain
```


#### Setting up databases for local development

Setup the local Atlas DB by running the `init_db.sh` script in the `./tests` folder:

```
cd tests
./init_db.sh
```

**Test this setup by trying the following curl commands**:
JSON summary data endpoints:
- curl http://localhost:8080/sources | python -m json.tool
- curl http://localhost:8080/cohortdefinition-stats/by-source-id/1 | python -m json.tool
- curl http://localhost:8080/concept/by-source-id/1 | python -m json.tool
- curl -d '{"ConceptIds":[2000000324,2000006885]}' -H "Content-Type: application/json" -X POST http://localhost:8080/concept-stats/by-source-id/1/by-cohort-definition-id/3 | python -m json.tool

TSV full data endpoint:
- curl -d '{"ConceptIds":[2000000324,2000006885]}' -H "Content-Type: application/json" -X POST http://localhost:8080/cohort-data/by-source-id/1/by-cohort-definition-id/3


Deprecated endpoints (TODO - remove from code):
- http://localhost:8080/cohortdefinitions
- http://localhost:8080/cohort/by-name/Test%20cohort1/source/by-name/results_and_cdm_DATABASE


# Deployment steps

## Deployment to QA

- PRs to `master` get the docker image built on quay (via github webhook). See https://quay.io/repository/cdis/cohort-middleware?tab=tags
- Once the image is built, it can be pulled. E.g. for branch `branch1`: `docker pull quay.io/cdis/cohort-middleware:branch1`
- If testing on QA:
   - ssh to QA machine
   - edit `/home/<qa-machine-name>/cdis-manifest/<qa-machine-name>.planx-pla.net/manifest.json` to set the desired image name and tag
     for cohort-middleware
   - run `run gen3 roll {service_name}`, e.g. `gen3 roll cohort-middleware`. See also https://github.com/uc-cdis/cloud-automation/blob/master/kube/services/cohort-middleware/cohort-middleware-deploy.yaml, which is used directly by the `gen3 roll` command (see https://github.com/uc-cdis/cloud-automation/blob/master/gen3/bin/roll.sh).
