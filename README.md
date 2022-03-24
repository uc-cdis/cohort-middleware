# Cohort-middleware

![version](https://img.shields.io/github/release/uc-cdis/cohort-middleware.svg) [![Apache license](http://img.shields.io/badge/license-Apache-blue.svg?style=flat)](LICENSE) [![Travis](https://travis-ci.org/uc-cdis/cohort-middleware.svg?branch=master)](https://travis-ci.org/uc-cdis/cohort-middleware) [![Coverage Status](https://coveralls.io/repos/github/uc-cdis/cohort-middleware/badge.svg?branch=master)](https://coveralls.io/github/uc-cdis/cohort-middleware?branch=master)

Cohort-middleware provides a set of web-services (endpoints) for:

1. providing information about Atlas DB cohorts that a user has authorized access to (as defined in Fence/Arborist?)
2. getting clinical attribute values (CONCEPT values in Atlas/OMOP jargon) for a given cohort
3. providing patient-level clinical attribute values matrix for use in backend workflows, like GWAS workflows (e.g. https://github.com/uc-cdis/vadc-genesis-cwl)

The cohorts and their clinical attribute values are retrieved from a
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

### Config file

See example config file HERE - TODO
