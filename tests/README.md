# About tests in this folder

Unit tests and integration tests (using real dependencies - e.g. tests on `./model/` code
use a set of real postgres DB containers).

# Setup (on your local machine)

1. For running he integration tests on your laptop, setup the local DB containers by running:

```
cd tests/setup_local_db
./init_db.sh
```

# Run

Run the tests (in root of this project folder) with:

```
./tests/run.sh
```


# Useful resources

- https://blog.alexellis.io/golang-writing-unit-tests/
