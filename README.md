# Cohort-middleware

![version](https://img.shields.io/github/release/uc-cdis/cohort-middleware.svg) [![Apache license](http://img.shields.io/badge/license-Apache-blue.svg?style=flat)](LICENSE) [![Travis](https://travis-ci.org/uc-cdis/cohort-middleware.svg?branch=master)](https://travis-ci.org/uc-cdis/cohort-middleware) [![Coverage Status](https://coveralls.io/repos/github/uc-cdis/cohort-middleware/badge.svg?branch=master)](https://coveralls.io/github/uc-cdis/cohort-middleware?branch=master)

Cohort-middleware provides a set of web-services (endpoints) for:

1. providing information about cohorts to which a user has authorized access (Atlas DB cohorts as defined in Fence/Arborist?)
2. getting clinical attribute values for a given cohort (aka CONCEPT values in Atlas/OMOP jargon)
3. providing patient-level clinical attribute values matrix for use in backend workflows, like GWAS workflows (e.g. https://github.com/uc-cdis/vadc-genesis-cwl)

The cohorts and their clinical attribute values are retrieved from a
connected OHDSI/CMD/Atlas database via SQL queries.

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
```sql
-- ========================================================
CREATE SCHEMA atlas;
-- ========================================================

CREATE TABLE atlas.source
CREATE TABLE atlas.source_daimon
CREATE TABLE atlas.cohort_definition
```

**DB Instance2**:

```sql
-- ========================================================
CREATE SCHEMA results;
-- ========================================================

CREATE TABLE results.COHORT
-- can we assume that the cohort_definition will also be present in results? That will make queries easier
CREATE TABLE results.COHORT_DEFINITION


-- ========================================================
CREATE SCHEMA cdm;
-- ========================================================

CREATE TABLE cdm.person
CREATE TABLE cdm.observation
CREATE TABLE cdm.concept
```


#### Setting up databases for local development

Setup the local Atlas DB by running:

```
docker run --name local-atlas-postgres --rm \
-p 5433:5432 \
-e POSTGRES_PASSWORD=mysecretpassword \
-d \
postgres:12.10-bullseye
```

Load ddl and test data:
```
docker exec -i \
local-atlas-postgres \
psql -U postgres -d postgres <$PWD/tests/ddl_atlas.sql

docker exec -i \
local-atlas-postgres \
psql -U postgres -d postgres <$PWD/tests/test_data_atlas.sql
```

Setup the local results and cdm DBs by running:

```
docker run --name local-results-and-cdm-postgres --rm \
-p 5434:5432 \
-e POSTGRES_PASSWORD=mysecretpassword \
-d \
postgres:12.10-bullseye
```

Load ddl and test data:
```
docker exec -i \
local-results-and-cdm-postgres \
psql -U postgres -d postgres <$PWD/tests/ddl_results_and_cdm.sql

docker exec -i \
local-results-and-cdm-postgres \
psql -U postgres -d postgres <$PWD/tests/test_data_results_and_cdm.sql
```
