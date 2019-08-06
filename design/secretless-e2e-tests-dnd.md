# Secretless automated end-to-end tests

An R&amp;D Technical Design and Delivery Plan

# References

- Aha! Ticket: _Add link_
- PRD Document: _Add link_
- Feature Proposal: _Add link_

# Overview

The Secretless project needs to add automated end-to-end (e2e) tests for Conjur OSS. The current e2e automation system which relies on [the kubernetes-conjur-demo repo](https://github.com/conjurdemos/kubernetes-conjur-demo) is bloated and likely running many redundant tests.

The goal here is to determine the WHAT and the HOW of Secretless end-to-end tests. 
1. For the WHAT, we seek to optimise the scope of the tests to avoid doing more testing than is necessary, ideally the bulk of the tests will reside in tests at a lower level of granularity i.e. unit tests.
2. For the HOW, we seek to reduce the complexity of orchestrating tests, to simplify maintenance and to make the testing pipeline developer-friendly.

Described more succinctly, the goal is to allow developers on the Secretless project to focus efforts on writing integration test cases with the assurance that some robust other thing will handle 
1. the orchestration of dependency resources (fixtures, services etc.) for the tests
2. the running of the tests
3. the reporting of the outcome of the tests

# Technical Objectives

_At a high-level, what are the technical requirements that this design needs to satisfy?  How can it meet the acceptance criteria presented in the Feature Proposal?_

- [ ] Tests run in a realistic environment with DAP/OSS as the source of truth for secret data
- [ ] Tests run in Kubernetes and OpenShift
- [ ] Tests run against relevant versions of DAP (v10+), and Conjur OSS
- [ ] Tests run against PostgreSQL and MySQL
- [ ] Tests validate that a demo application is able to run as expected while making database connections via Secretless
- [ ] The test suite does not need to run with every PR of Secretless, but should run daily and we should be able to run it against a local build of the Secretless image to prep for a release

- [ ] Criteria exists for what is allowed to be an end to end test case
- [ ] E2E test validate specific behavior and justification for the test is documented

## Out of Scope

1. Creating sophisticated implementations of service lifecycle management will not be part of this effort e.g. pooling pre-existing dependency resources to streamline load times
2. Being exhaustive in testing all the possible versions of services like Conjur. Though this would be thorough this would be time consuming and costly. In the future we might consider having test scenarios that are less costly and less representative of production but close enough to it to be of value. An example is running a branch follower against a stable master.
3. juxtaposer will eventually need to be automated. This is not addressed as part of this effort

# Stakeholder Solution Sign-Off

| **Stakeholder Role** | **Name** | **Date** |
| --- | --- | --- |
| Product Manager |   |   |
| Product Owner |   |   |
| Feature Lead | Kumbirai Tanekha  |   |

# Experience

## Assumptions

We assume that the experience of running automated tests given configured dependency resources (such as Conjur and target services) is decent enough at this point in time to not warrant consideration in the improvements suggested by this document. This means  a user (a developer or Jenkins) is able to run integration tests in an automated fashion and get meaningful feedback on the test outcome.

Moving the lifecycle management of dependency resources behind an opaque API is a task that can be achieved with minimal effort.

## Overview

1. A developer makes changes to the Secretless source. The developer runs a script to run integration tests. The part of the script responsible for setting up external dependencies, such as Conjur and target service databases is opaque and delegated to some robust backend which is well-tested and reliable. The API merely requires configuration information for setting up the external dependencies.
 
2. A well-defined workflow with minimal boilerplate exists to add and modify integration tests. The experience of running integration tests is identical on Jenkins and on a developer's machine.

3. There are clear criteria for what ought to be an integration test vs a unit test such that deciding to embark on 2 is easy.

## Walkthroughs

TBD

# Technical Design

## Architecture

TBD


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

## Documentation

1. The API for requesting dependencies has documentation. The documentation covers 
    1. Usage: how to request dependencies, how dependencies are exposed, procedures for cleanup
    2. General information: Assumptions, possibilities and limitations
    3. Configuration
 
1. The mechanism for running has documentation. The documentation covers how to run, gathering 

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

Q: How do we test a test runner ?

Q: How do we test the opaque API for creating dependency resources ?

Q: How will we ensure that improvements we'd like to have in the mid-term aren't put in the icebox forever ?

Q: Do all Secretless test cases need a full cluster including master, standbys and followers ?

A: Some test cases do not need all the components of the cluster. For example, in testing Conjur authn-k8s and service connection all that is necessary is an API that is able to return secrets to an authenticated client. However, performance tests require a high fidelity service to accurately measure performance. This consideration makes it possible to optimise on startup times for the former test cases and have features like pooling.

## Meeting Notes

TBD

Parts of kubernetes-conjur-demo that are relevant to e2e testing for secretless need to be split out into a seperate project dedicated to secretless e2e testing.
A test runner should be leveraged 
test outcomes
code coverage,
testing metrics,
historical tracking

Idea: Fork `kubernetes-conjur-deploy`, `kubernetes-conjur-demo` to avoid modifying it and dealing with the implications. For now we implement the opaque API using the fork. This can be followed later by deprecation of kubernetes-conjur-deploy and adoption of the opaque API. 
Pros:
+ Makes the first pass at implementation much simpler.
+ Move the repo to `cyberark` org.
+ Ignore version 4.
+ Allows us to abstract the testing steps of setting up a demo application so that they can be reused for testing (not specific) to demo. We can hide this behind an API. It doesnt't have to be an example anymore.
Note: Might not want to fork `deploy` to avoid digging into DAP deployment details. `demo` on the other hand is more pressing because we need to overhaul the process of creating and running test cases.
Note: Strongly consider Cucumber as test runner for non-unit tests. Great for composition etc. Ability to use previously defined steps to construct new test cases

TODO: Ask Andy for an example test matrix 

NOTE: the goal of this is to improve the process. add OSS test case. the actual test cases can be technical debt which will become simpler to address given the improved process.

 
For the architecture section we can talk about how cucumber and the API are used and interact. 

Infrastructure section: Requires new tools from infra but not new infrastrcture

Security section: Open question

Docuemntation: 
1. secretless-e2e-tests project documents how to go about creating new tests, running them locally and running them on jenkins after commit or pushed to master
2. interfaces lives in bash-lib with documentation of usage. still not sure where our modules and all the implementation bits will live. 

# Addendums

TBD
