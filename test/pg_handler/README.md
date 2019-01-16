# Postgresql Handler Development

##TLDR
The following two steps are all you need to know:
1. A single command starts the dev workflow:

    ```sh-session
    $ ./dev
    ```
2. Another one runs the tests:
    ```sh-session
    $ ./test
    ```
So while developing, you'll do a single `./dev` and then many `./test` runs as you iteratively change the code or add new tests.

##Additional Details

`./dev` uses `docker-compose` to start both `pg` containers, the `secretless` container, and the `test` container (where tests are run from).  

The `test` container does not get torn down after tests are run, as it does when invoked in normal (non-dev) mode.  This means the required Go packages only need to be downloaded on the first test run.  And this means that subsequent test runs are very fast.
