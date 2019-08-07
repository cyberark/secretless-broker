# Secretless automated end-to-end tests

An R&amp;D Technical Design and Delivery Plan

# References

- Aha! Ticket: _Add link_
- PRD Document: _Add link_
- Feature Proposal: _Add link_

# Overview

The Secretless project needs to add automated end-to-end (e2e) tests for Conjur OSS. The current e2e automation system which relies on [the kubernetes-conjur-demo repo](https://github.com/conjurdemos/kubernetes-conjur-demo) is bloated and likely running redundant tests.

Broadly speaking the optimisation needed are on the WHAT and the HOW of Secretless end-to-end tests.
1. For the HOW, we focus on process and seek to reduce the complexity of orchestrating tests, to simplify maintenance and to make the testing pipeline developer-friendly.
2. For the WHAT, we seek to optimise the scope of the tests to avoid doing more testing than is necessary, ideally the bulk of the tests will reside in tests at a lower level of granularity i.e. unit tests.

The goal of this particular effort is to improve the process of creating and maintaining Secretless end-to-end tests. It is inspired by the immediate need to add a test case for Conjur OSS. Our current test cases assume success implicitly as running through some number of steps without encountering failure. WHAT we test requires some thought and improvement, however we deffer this to a later effort while keeping in mind that this addressing this technical debt will become simpler as a result of an improved process.

The improvements in process are expected to allow developers on the Secretless project to focus efforts on writing integration test cases with the assurance that some robust other thing will handle
1. dependency service lifecycle management
2. running of the tests
3. measuring and reporting the outcomes of the tests

# Technical Objectives

_At a high-level, what are the technical requirements that this design needs to satisfy?  How can it meet the acceptance criteria presented in the Feature Proposal?_

- [ ] Tests run in a realistic environment with DAP/OSS as the source of truth for secret data
- [ ] Tests run on developer machine ?
- [ ] Tests run in Kubernetes and OpenShift
- [ ] Tests run against relevant versions of DAP (v10+), and Conjur OSS
- [ ] Tests run against PostgreSQL and MySQL
- [ ] Tests validate that a demo application is able to run as expected while making database connections via Secretless
- [ ] The test suite does not need to run with every PR of Secretless, but should run daily and we should be able to run it against a local build of the Secretless image to prep for a release

## Out of Scope

1. Creating sophisticated implementations of service lifecycle management will not be part of this effort e.g. pooling pre-existing dependency resources to streamline load times
2. Being exhaustive in testing all the possible versions of services like Conjur. Though this would be thorough this would be time consuming and costly. In the future we might consider having test scenarios that are less costly and less representative of production but close enough to it to be of value. An example is running a branch follower against a stable master.
3. juxtaposer will eventually need to be automated. This is not addressed as part of this effort
|

# Experience

## Assumptions

We assume that the experience of running automated tests given configured dependency resources (such as Conjur and target services) is decent enough at this point in time to not warrant consideration in the improvements suggested by this document. This means  a user (a developer or Jenkins) is able to run integration tests in an automated fashion and get meaningful feedback on the test outcome.

Moving the lifecycle management of dependency resources behind an opaque API is a task that can be achieved with minimal effort.

## Overview

There is an idea for a test.
1. A developer writes the test case in as close to declarative a fashion as possible, leveraging preexisting steps and assertions.
2. A developer specifies the service dependencies for their test case.
3. A developer runs a simple command to run the tests, this results in either success or failure, in both cases the outcome is reported and any resources used in the test are cleaned up (unless requested to stay). In the case of failure breadcrumbs are provided fo where in the test the failure occurred.
4. A developer adds the test to CI and pushes the branch got Github
5. A repetition of 3 occurs except this case it in CI

## Walkthroughs

TBD

# Technical Design

## Architecture

[ Service ] is any software that tests rely on e.g. Conjur, Postgres etc.

[ Service Config ] specifies details about a service e.g. for Conjur `CONJUR_ACCOUNT`, `version` etc. Conjur will have different configuration options to Postgres. These options will need to be documented. It's important for users of the Service Engine to know what they're able to configure on a service. Config also specifies the lifecyle method of the interface (start, stop, clean)

[ Test Runner ] is software that provides a conventional way for defining and running test cases. They typically come with capabilities to measure test metrics, test outcomes and code coverage. This will also allow historical tracking. An example of a test-runner is Cucumber. Test cases can be defined in Gherkin

[ Service Engine ] uses [ Service Config ] to create an instance of a [ Service ] 

[ Test Case ] specifies a test case and contains some annotations detailing required [ Service ]s 

[ Test Runner ] gathers annotations from [ Service ]s from [ Test Case ]s and issues requests for services to the [ Service Engine ] before running all [ Test Case ].

[ Service Engine ] uses [ Service Config ] to create an instance of a [ Service ].

Here is an example of the components in this architecture working together.
0. A service definition exists for conjur and mysql
```bash
# Conjur service definition.  This is supplied to the start_services function and defines the service to be started
declare -A conjur_service_config
conjur_service_config=(
["service_name"]="conjur_abc",
["config"]="conjur.yaml",
["start"]="conjur_start.sh",
["stop"]="conjur_stop.sh"
)

declare -A mysql_service_config
mysql_service_config=(
["service_name"]="mysql_abc",
["config"]="mysql.yaml",
["start"]="mysql_start.sh",
["stop"]="mysql_stop.sh"
)
``` 
1. There is a test case written in Gherkin for validating using the MySQL service connector with Conjur.
```
Feature: Guess the word
  
  @conjur @mysql @mysql-conjur
  Scenario: Connection to MySQL
    Given Secretless is running as a sidecar in Kubernetes
    Then the app is able to connect to MYSQL on "localhost:332"

```
2. Cucumber will register that there's a test case that requires Conjur, MySQL and MySQL credentials in Conjur.
3. The developer calls `cucumber` to start the testss
4. As a result of the annotations Cucumber will first call the engine to create the services.
5. The configuration values
 
TODO: provide examples of terms above
TODO: split out components of the architecture by owner (infra, project developer etc.)

## Interfaces

TBD

### Opaque API for lifecycle management of dependency resources

The API makes it possible to implement service lifecycle methods in whatever way makes sense. That could be literal bash, docker-compose, calling an HTTP API to grab the service from a pool etc. This makes it possible to have similar experience between different environments (Jenkins, developer machine etc.). The same exact commands would be called on Jenkins as on the developer machine. The action of the commands would depend on the environment. For example, the developer machine can use a less resource intensive implementation of the service and Jenkins can run something that more closely approximates customer environments.

Steps towards implementation

1. In the short term we can change `kubernetes-conjur-demo` so that the proposed interface masks calls to `kubernetes-conjur-deploy`
2. Later the infrastructure team or whoever can modify `kubernetes-conjur-deploy` jumping into cleaning up and improving implementation details without adversely affecting existing tests.
3. From then on efforts to optimise for startup time, performance, customer-environment accuracy can continue. Once again, without adversely affecting existing tests.

## Testing

TBD

## Infrastructure

The orchestration of dependencies (fixtures, services etc.) will need to be run somewhere. Some dependencies such as Conjur followers need to be run on the same infrastructure as the tests. Some dependencies such as target services can and perhaps should be run on separate infrastructure from the tests.

Q: It's not clear if this requires new infrastructure

## Security

TBD

This section is still an open question. It's not clear at the moment if there are any security implications to the work under consideration.

## Documentation

1. Seperate `secretless-e2e-tests` project shall document how to go about creating new tests, running them in a variety of environments (locally, and on jenkins after commit or pushe to master)
2. API for service lifecycle management will be located in bash-lib.  There will be documentation of usage. It remains to be determined where service implementations will reside.
    1. Usage: how to request dependencies, how dependencies are exposed, procedures for cleanup
    2. General information: Assumptions, possibilities and limitations
    3. Configuration

## Considerations and Alternatives

### Dependency Resources Lifecycle Management
1. A fully fledged opaque API for managing the lifecycle of dependency resources would be ideal.
2. 1 might might be cost-prohibitive so an alternative is to place the current mess of an implementation behind an API that allows us to carry out the clean up while maintaining a consistent interface

### What & How to test

1. Versions of Conjur to test
2. Scenarios to test (happy vs sad)
3. Well defined mechanism for reporting test outcome (e.g. junit)

# Future Work

# Delivery Plan

## Effort Estimates

TBD

### Original Planning Guestimate

TBD

### Detailed Estimate

TBD

| **Best Case ** | **Most Likely Case  ** | **Worst Case ** |
| --- | --- | --- |
| _n_ engineer weeks  | _n_ engineer weeks  | _n_ engineer weeks  |

## Components and Milestones

1. Opaque API for requesting Conjur exists
1. Opaque API for requesting Postgres, MySQL exists with the ability to specify version and SSL configurations
1. Documentation exists for new APIs
1. A draft exists for E2E test criteria
1. Tests cases are added and removed in response to criteria
1. Tests cases exists exercise the agreed-upon variations of dependency resources

## Story Breakdown

TBD

# Solution Sign-Off

- **Meeting Date: YYYY/MM/DD**
- **Attendees:**

## Open Questions

Q: How ambitious do we want to be in the short-term ?

Q: What's a good test runner ?

Q: How do we test a test runner ?

Q: How do we test the opaque API for creating dependency resources ?

Q: How will we ensure that improvements we'd like to have in the mid-term aren't put in the icebox forever ?

Q: Where will service lifecycle management implementations reside ?

Q: Do all Secretless test cases need a full cluster including master, standbys and followers ?

A: Some test cases do not need all the components of the cluster. For example, in testing Conjur authn-k8s and service connection all that is necessary is an API that is able to return secrets to an authenticated client. However, performance tests require a high fidelity service to accurately measure performance. This consideration makes it possible to optimise on startup times for the former test cases and have features like pooling.

## Meeting Notes

TBD

Parts of kubernetes-conjur-demo that are relevant to e2e testing for secretless need to be split out into a seperate project dedicated to secretless e2e testing.


Idea: Fork `kubernetes-conjur-deploy`, `kubernetes-conjur-demo` to avoid modifying it and dealing with the implications. For now we implement the opaque API using the fork. This can be followed later by deprecation of kubernetes-conjur-deploy and adoption of the opaque API.
Pros:
  + Makes the first pass at implementation much simpler.
  + Move the repo to `cyberark` org.
  + Ignore version 4.
  + Allows us to abstract the testing steps of setting up a demo application so that they can be reused for testing (not specific) to demo. We can hide this behind an API. It doesnt't have to be an example anymore.
Note: Might not want to fork `deploy` to avoid digging into DAP deployment details. `demo` on the other hand is more pressing because we need to overhaul the process of creating and running test cases.
Note: Strongly consider Cucumber as test runner for non-unit tests. Great for composition etc. Ability to use previously defined steps to construct new test cases

TODO: Ask Andy for an example test matrix

TODO: Ask Matt to prioritise definition of the interface. This allows us to de-risk him being away. This would allow Secretless to create the `secretless-e2e-tests` repo so that we can use the interfaces while using mocks or `kubernetes-conjur-deploy`. We can create a throw away implementation to validate the interface for

For the architecture section we can talk about how cucumber and the API are used and interact.

Infrastructure section: Requires new tools from infra but not new infrastructure

Security section: Open question

# Addendums

TBD
