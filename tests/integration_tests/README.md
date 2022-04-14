# About tests in this folder

Tests using real dependencies. E.g. tests on `./model/` code
hitting a real DB.


# Setup

1. Setup a local DB by running:

```
../setup_local_db/init_db.sh
```

# Run

Run the tests with

```
go test ./... -v -coverpkg=./...
```


# Useful resources

- https://blog.alexellis.io/golang-writing-unit-tests/
