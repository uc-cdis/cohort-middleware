# About

This folder contains logic to populate a test DB based on a configuration file.
The configuration file can be used to indicate how many and what type of records
should be added to the test DB. See the `.yaml` files in this folder for examples.

# How to run

## Setup local test DB

To setup a local test DB, run:

```bash
(cd ../setup_local_db; ./init_db.sh)
```

## Simple command w/ default config files

To run a simple test on your local DB, run the command below. It will read the DB
config from the default [../../config/development.yaml](development.yaml) file in this project and
generate test data based on [models_tests_data_config.yaml](models_tests_data_config.yaml) in this folder:

```bash
go run datagenerator.go
```

## Specify your own DB connection and test data config files

To use this command to populate another DB, you can add your own DB `.yaml` config file
and your own test data `.yaml` file to the current folder and run the command like this:

```bash
go run datagenerator.go -e example_custom_db -d example_test_data_config -s 1
```
where:
 - `example_custom_db` is the prefix of the DB config .yaml file
 - `example_test_data_config` is the prefix of the test data .yaml file
 - `1` is the source id for the Omop DB (see your Atlas table `source` to find out what is the
    sourceId in your case)

After a successful run you should get something like:

```
============================= DONE! =================================

2023/03/15 20:18:21 Added this to your DB:
 - Persons: 300
 - Concepts: 20
 - Observations: 750
 ```

## Running in a container

For running in the container, `exec -it` the container to run the command:

```bash
/data-generator -e /config/development.yaml -d  /example_dataset.yaml -s 2
```
