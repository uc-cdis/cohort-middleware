#/bin/bash

# Run this script to setup new test databases

#### Setting up databases for local development

# Stop (and auto-remove) previous instance of DB containers:
docker stop local-atlas-postgres
docker stop local-results-and-cdm-postgres

# Setup the local Atlas DB:
docker run --name local-atlas-postgres --rm \
-p 5433:5432 \
-e POSTGRES_PASSWORD=mysecretpassword \
-d \
postgres:12.10-bullseye

sleep 3

# Load ddl and test data:
docker exec -i \
local-atlas-postgres \
psql -U postgres -d postgres <$PWD/ddl_atlas.sql

docker exec -i \
local-atlas-postgres \
psql -U postgres -d postgres <$PWD/test_data_atlas.sql

# Setup the local results and cdm DBs:
docker run --name local-results-and-cdm-postgres --rm \
-p 5434:5432 \
-e POSTGRES_PASSWORD=mysecretpassword \
-d \
postgres:12.10-bullseye

sleep 3

# Load ddl and test data:
docker exec -i \
local-results-and-cdm-postgres \
psql -U postgres -d postgres <$PWD/ddl_results_and_cdm.sql

docker exec -i \
local-results-and-cdm-postgres \
psql -U postgres -d postgres <$PWD/test_data_results_and_cdm.sql
