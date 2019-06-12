# Perf Test Agent

## Description

This tool can be used to compare timing data between arbitrary number of similar
services to evaluate differences in speed between a specified baseline backend
and the other backends.

Specifically, this tool is used here as a performance test agent deployed
alongside Secretless to compare the following scenarios:

| Backend       | Baseline          | Compare With                       | Connection Type |
| ---           |---                | ---                                |---              |
| MySQL         | Direct connection | Secretless (persistent connection) | Unix socket     |
| MySQL         | Direct connection | Secretless (persistent connection) | TCP port        |
| Postgres      | Direct connection | Secretless (persistent connection) | Unix socket     |
| Postgres      | Direct connection | Secretless (persistent connection) | TCP port        |

We compare the following results:

- Returned values versus expected values
- Number of rounds (single-shot tests runs) completed
- Average/Min/Max single-shot test duration
- Error count, error messages, and percentage of errors
- Differences in timing (percentage-based) between runs temporaly close
- Percentage of single-shot runs that are above the specified baseline threshold
- 90% confidence interval (percentage-based) of test runs as compared to the
baseline backend.

Note: More comparison types may be added in the future.

---

### **Status**: Alpha

#### **Warning: Naming and APIs are still subject to breaking changes!**

---

## CLI flags

### `-c`: Continue running after end of tests

Leaves the process running after the tests are complete and the results are
printed. This protects the data from container log reaping.

### `-f`: Path to configuration file

Overrides the default configuration file path, which is `./juxtaposer.yml`.

### `-t`: Run tests for this duration

Run tests for a specified time rather than a number of loops.  Uses Golang's
[ParseDuration](https://golang.org/pkg/time/#ParseDuration) format to specify
time (e.g. `10h5m3s`).

## Configuration

Configuration is specified either with a default `juxtaposer.yml` in the current
directory or overridden by `-f` CLI flag.

## Format

A configuration file looks like this:

```yaml
driver: <DRIVER_NAME>

comparison:
  baselineBackend: backend1

formatters:
  stdout:

backends:

  backend1:
    host: <HOST>
    port: <PORT>
    username: <USERNAME>
    password: <USERNAME>
    sslmode: disable

  backend2:
    host: <PATH_TO_SOCKET>
```

## Supported Drivers:

- `mysql-5.7`
- `postgres`

## `comparison` options

1. `baselineBackend`, `string`
This setting is linked to a named backend to indicate the baseline backend
against which all calculation will be compared.

1. (optional) `type`, `string`, default: `sql`
This setting decides what type of comparison will be run. Only `sql` is
currently supported.

1. (optional) `style`, `string`, default: `select`
This setting decides what style (subtype) of comparison will be run. Currently
only `select` is supported.

1. (optional) `rounds`, `int`, default: `1000`
This setting decides how many loops the main test run will iterate over all
the defined backends. This setting is ignored if time-based CLI flag is used.
This field also supports a special keyword `infinity` that lets the tests run
forever or until the user sends an interrupt signal.

1. (optional) `baselineMaxThresholdPercent`, `int`, default: `120`
This setting decides how slow (percentage-wise) can a non-baseline backend be
before it is counted as "exceeding threshold".

1. (optional) `silent`, `bool`, default: `false`
When this setting is turned on, only minimal test run messages are displayed during
runs.

## `formatters` options

This key-value map indicates what formatters are run at the end of tests to handle
aggregated results. Currently only `json` and `stdout` are supported. `json` formatter
currently also supports `outputFile` options which can be used to send the output data
from it to a file instead of stdout.

## `backends` options

This key-value map lists all backends that should be tested as part of the agent's
running.

Options are as follows:
```
host: Hostname (or socket address) that the driver will use. Note that for PG driver and sockets, this is the *folder* instead of the path to the socket file.
port: Port to use when connecting to the backend
username: Username to use for connecting to the backend. Some drivers do not need this setting.
password: Password to use for connecting to the backend. Some drivers do not need this setting.
sslmode: SSL mode to use for connecting to the backend. Some drivers do not need this setting.
debug: Print out verbose information about this backend test runs
ignore: If set to true, the agent will ignore this backend
```
