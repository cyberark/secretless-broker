# Secretless automated end-to-end tests

An R&amp;D Technical Design and Delivery Plan

# References

- [Aha! Ticket](https://cyberark.aha.io/features/AAM-75)
- PRD Document: n/a
- [Feature Proposal](https://cyberark365.sharepoint.com/:w:/s/Conjur/EZLGsx4HM79OhZA-jU87XoIBx_idED-wdWf7lKbnWH8rqw?e=4G6Rzh) (part of the technical debt cleanup work in the "Secretless supports Conjur OSS" epic)

# Overview

The Secretless project needs to add automated end-to-end (e2e) tests for Conjur OSS. The current e2e automation system which relies on [the kubernetes-conjur-demo repo](https://github.com/conjurdemos/kubernetes-conjur-demo) is bloated and likely running redundant tests. As well as Secretless, this repo exercises summon and the autheticator sidecar. It is unsophisticated in that it implicitly validates functionality by measuring success as running from start to finish. It doesn't offer particularly rich information regarding test runs aside from unstructured logs and so can be difficult to debug.

Broadly speaking the optimisation needed are on the WHAT and the HOW of Secretless end-to-end tests.
1. For the HOW, we focus on process and seek to reduce the complexity of orchestrating tests, to simplify maintenance and to make the testing pipeline developer-friendly.
2. For the WHAT, we seek to optimise the scope of the tests to avoid doing more testing than is necessary, ideally the bulk of the tests will reside in tests at a lower level of granularity i.e. unit tests.

The goal of this particular effort is to improve the process of creating and maintaining Secretless end-to-end tests. It is inspired by the immediate need to add a test case for Conjur OSS. As mentionted before, current test cases measure success implicitly as running through some number of steps without encountering failure. WHAT we test requires some thought and improvement, however we deffer this to a later effort while keeping in mind that addressing this technical debt will become simpler as a result of an improved process.

The improvements in process are expected to allow developers on the Secretless project to focus efforts on writing integration test cases with the assurance that some, other robust system/tool will handle the following:
1. dependency service lifecycle management
2. running of the tests
3. measuring and reporting the outcomes of the tests

# Technical Objectives

_At a high-level, what are the technical requirements that this design needs to satisfy?  How can it meet the acceptance criteria presented in the Feature Proposal?_

- [ ] Tests run in a realistic environment with DAP/OSS as the source of truth for secret data
- [ ] Tests can be run by a developer. This means kicked off from the developer's machine. The tests need not run on the developer's machine as some parts won't make sense to run on the developer's machine.
- [ ] Tests run in Kubernetes and OpenShift
- [ ] Tests run against relevant versions of DAP (v10+), and Conjur OSS
- [ ] Tests run against PostgreSQL and MySQL
- [ ] Tests validate the critical scenario of a test app running as expected while making database connections via Secretless
- [ ] The test suite does not need to run with every PR of Secretless, but should run daily and we should be able to run it against a local build of the Secretless image to prep for a release

## Out of Scope

1. Creating sophisticated implementations of service lifecycle management will not be part of this effort e.g. pooling pre-existing dependency resources to streamline load times
2. Being exhaustive in testing all the possible versions of services like Conjur. Though this would be thorough this would be time consuming and costly. In the future we might consider having test scenarios that are less costly and less representative of production but close enough to it to be of value. An example is running a branch follower against a stable master.
3. [`juxtaposer`](https://github.com/cyberark/secretless-broker/tree/master/bin/juxtaposer), the Secretless perfomance testing framework, will eventually need to be automated. This is not addressed as part of this effort

# Experience

## Assumptions

1. Moving the lifecycle management of dependency services behind an opaque API is a task that can be achieved with minimal effort. This is supported by the fact that the test run currently works while poluting the test source with service provisioning logic. The act of moving this logic behind an opaque API could be reduced to housing it in a seperate repository, that would still be an improvement. In this effort we go several steps further and take the opportunity to create a robust abstract with additional benefits. If at some point that becomes unwiedly we are always at liberty to descope the work and be less ambitious with the API.
2. Adopting a test runner for writing and runing automated tests, given configured dependency resources (such as Conjur and target services), can be achieved with minimal effort. This is supported by the fact that there is currently only one path/test excercised as part of `kubernetes-conjur-demo` and that we only intend to add one more for now.

## Overview

We begin with an idea for a test.
1. A developer writes the test case in as close to declarative a fashion as possible, leveraging preexisting steps and assertions; defining new steps and assertions if necessary.
2. A developer specifies the service dependencies for their test case using the predefined test cases that come with the Service Engine (defined in the Architecture section).
3. A developer runs a simple command to run the tests, this results in either success or failure, in both cases the outcome is reported and any resources used in the test are cleaned up (unless requested to stay). In the case of failure breadcrumbs are provided fo where in the test the failure occurred.
4. A developer adds the test to CI via Jenkinsfile and pushes the branch to Github
5. The tests run automatically as in Step #3, this time on CI infrastructure. 

## Walkthroughs

TBD

# Technical Design

## Architecture

A `Service` is any external software that end to end tests rely on e.g. Conjur, Postgres etc.

A `Service Config` is a JSON object that specifies information required for a service to be created e.g. for Conjur `CONJUR_ACCOUNT`, `version` etc. Conjur will have different configuration options to Postgres. These options must be documented for each Service. Ideally this would be in a centralized repository. It's important for all users of a Service to know what they're able to configure on a service.
```
{
  "conjur_account": "test",
  "version": "10.9",
  "admin_user": "kumbi",
  "admin_password": "definetelyasecret"
}
```

A `Service Definition` is a JSON object that specifies the lifecyle methods of a Service (start, stop). 
    1. `start_command`. This command sets up the service and outputs any meaningful information as part of the JSON output. The value is an array to avoid CLI ambiguity.
    2. `stop_command`. This command carries out any required regardless of test outcome. It is important for the test runner to ensure that this is run on both success and failure. The value is an array to avoid CLI ambiguity.
    3. `service_type`. This is optional and for the future. The `service_type` specifies a internal implementation of a Service. The internal implementation come from a standard library provided by and maintened by the Infrastructure team.
```
{
  "service_type": "mysql",
  "start_command": [ "mysql_start.sh"
  "stop_command": [ "mysql_stop.sh" ]
}
```

A `Test Runner`  is software that provides a conventional way for defining and running test cases. They typically come with capabilities to measure test metrics, test outcomes and code coverage. This will also allow historical tracking. An example of a test-runner is Cucumber. Test cases can be defined as a sequence of steps in the Gherkin language. The step definitions are written in the language of the test runner e.g. Python, Ruby, Go. In this case we're strongly considering adopting `Cucumber` in Ruby as the test-runner. Ruby because of the ubiquity of working knowledge of that language in the team and the fact that there are Ruby Cucumber test suites that exist within some of our repos.

The `Service Engine` is a combination of steps defined in the `Test Runner`. The Engine orchestrates the lifecycle of a Service. It provides step definitions for starting Services, stopping services regardless of the test outcome, creating named pipes for the Service output and writing that to the test context.

A `Test Case` is a sequence of steps that goes through a scenario to validate some given functianlity. It should really include at least one step making an assertion as part of the validation.

A `Context` is a temporary store within a test case. Steps can read and write from the store.

Here is an example of the components in this architecture working together.

1. Service definitions exist for Conjur and MySQL
    1. In this example it is MySQL and Conjur. An example for MySQL is provided below.
    ```
    {
      "service_type": "mysql",
      "start_command": [ "mysql_start.sh" ], # imagine that "mysql_start.sh" calls out to "helm install"
      "stop_command": [ "mysql_stop.sh" ]
    }
    ```
1. Service Configs exist for Conjur and MySQL. These files would exist as `./mysql_config.json` and `./conjur_config.json`. An example for Conjur is provided below
    ```
    {
      "conjur_account": "test",
      "version": "10.9",
      "docker_image": "registry.tld/conjur-appliance:10.9-stable"
      "admin_user": "kumbi",
      "admin_password": "definetelyasecret"
    }
    ```
1. A service instances file exists for the test case under consideration. This file contains an array of Service Definition and Service Config pairs. This files would be called `services.json`
    ```
    [
      {
          "config": "./mysql_config.json",
          "service_definition": "mysql_service",
      },
      {
          "config": "./conjur_config.json",
          "service_definition": "conjur_service"
      }
    ]
    ```
1. There is a test case written in Gherkin for validating using the MySQL service connector with Conjur. See below that there is an expliit step specifying the services that are expected to be running. 
    ```
    Feature: Secretless happy path with database connectors
    
      Scenario: Connection to MySQL with MySQL credentials stored in Conjur
        Given running services in ./services.json
        When I store secrets ./conjur-secrets.json to Conjur
        And I create an app identity ./application_identity.json
        And I deploy a test app ./test-app.yml
        Then the app can connect to MYSQL on "localhost:332"
    ```
1. The command `cucumber` is called to run test cases. The command is configurable and allows the runner to select a subset of all tests to run.

## Interfaces

### Opaque API for lifecycle management of dependency resources

The API makes it possible to implement service lifecycle methods in whatever way makes sense. That could be literal bash, docker-compose, calling an HTTP API to grab the service from a pool etc. This makes it possible to have similar experience between different environments (Jenkins, developer machine etc.). The same exact commands would be called on Jenkins as on the developer machine. The action of the commands would depend on the environment. For example, the developer machine can use a less resource intensive implementation of the service and Jenkins can run something that more closely approximates customer environments.

### Steps towards implementation

1. In the short term we can change `kubernetes-conjur-demo` so that the Service Definition for Conjur masks calls to `kubernetes-conjur-deploy`
2. Later the infrastructure team or whoever can fork and modify `kubernetes-conjur-deploy` (or start from scratch). This means improvements (e.g. optimise startup time, performance, customer-environment accuracy continue) could be made to the service implementation without affecting existing tests. As long as the interface contract is enforced all things remain equal.

## Testing

TBD

## Infrastructure

The orchestration of dependencies (fixtures, services etc.) will need to be run somewhere. Some dependencies such as Conjur followers need to be run on the same infrastructure as the tests. Some dependencies such as target services can and perhaps should be run on separate infrastructure from the tests.

Q: It's not clear if this requires new infrastructure
A: No, for now everything can run on preexisting infrastructure. There will need to be a mechanism to map service requests and infrastructure. 

## Security

TBD

This section is still an open question. It's not clear at the moment if there are any security implications to the work under consideration.

## Documentation

Documentation will need to be created as part of this effort. The improvement and tools under discussion are directed towards the developer. Documentation will be written to assist the developer.

1. Seperate `secretless-e2e-tests` project shall document how to go about creating new tests, running them in a variety of environments (locally, and on jenkins after commit or pushes to master)
2. API for service lifecycle management will be located in `bash-lib`.  There will be documentation of usage. It remains to be determined where service implementations will reside.
    1. Usage: how to request dependencies, how dependencies are exposed, procedures for cleanup
    2. General information: Assumptions, possibilities and limitations
    3. Configuration

## Considerations and Alternatives

### Dependency Resources Lifecycle Management
1. A fully fledged opaque API for managing the lifecycle of dependency resources would be ideal. This will be cost-prohibitive so an alternative is to place the current implementation behind an API that allows us to carry out any improvements under a consistent interface.

The plan is to leverage existing work and take incremental steps torwards the intended improvements. For now, the Conjur Service Definition will use a fork of `kubernetes-conjur-deploy`. This will avoid modifying the original repo since it is user-facing. This can be followed later by deprecation of `kubernetes-conjur-deploy` and adoption of the opaque API and Conjur Service Definition.
Pros:
  + Makes the first pass at implementation much simpler.
  + Move the repo to `cyberark` org.
  + Ignore version 4.
  + Allows us to abstract the testing steps of setting up a demo application so that they can be reused for testing (not specific) to demo. We can hide this behind an API. It doesnt't have to be an example anymore.
We will not want to do too much work on the fork of `kubernetes-conjur-deploy` to avoid digging into DAP deployment details at this point in time.

### What & How to test

1. Versions of Conjur to test
2. Scenarios to test (happy vs sad)
3. Well defined mechanism for reporting test outcome (e.g. junit)

 We will extract the parts of `kubernetes-conjur-demo` relevant to the end to end tests. These parts will be moved into `secretless-e2e-tests` as part of overhauling the process of creating and running test cases. The results will be test cases written in the Cucumber Gherkin language.
 
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

1. Create secretless-e2e-tests repo with all the plumbing and documentation for the Test runner in place.
1. Create the Service Engine within secretless-e2e-tests
1. Document the Service Definition interface in secretless-e2e-tests
1. Service Definition using a fork of `kubernetes-conjur-deploy` exists for Conjur with the ability to specify version, Kubernetes connection details etc. and retrieve connection details
1. Service Definition for Postgres and MySQL exists with the ability to specify version and SSL configurations
1. Documentation exists for new APIs
2. A mechanism exists for specifying 

## Story Breakdown

- [ ] Create repository to hold secretless-e2e-tests framework (**_TODO: Figure out if a separate repo is needed_**)

- [ ] **Epic**: Implement service engine (**_Note: Service definition should be driven by implementation_**)
  - [ ] Implement a _tested_ TBD-language dummy service engine (simple webserver)
  - [ ] Implement a TBD-language CLI service runner for Conjur OSS
  - [ ] Document service runner for Conjur OSS
  - [ ] Implement a TBD-language CLI service runner for Postgres
  - [ ] Document service runner for Postgres
  - [ ] Implement a TBD-language CLI service runner for MySQL
  - [ ] Document service runner for MySQL

- [ ] **Epic**: Implement Cucumber runner
  - [ ] (Spike/Throwaway) Find workable/appropriate combination of gherkin and backend languages
  - [ ] Implement Cucumber runner based on spike learning (**_Note: This should have only a single sample test and Jenkins-parsed output_**)
  - [ ] Implement invocation of dummy service runner by Cucumber
  - [ ] Document Cucumber API for generic service runner invocation
  - [ ] Implement invocation of Conjur OSS service runner by Cucumber
  - [ ] Document Cucumber API for Conjur OSS service runner invocation
  - [ ] Implement invocation of Postgres service runner by Cucumber
  - [ ] Document Cucumber API for Postgres service runner invocation
  - [ ] Implement invocation of MySQL service runner by Cucumber
  - [ ] Document Cucumber API for MySQL service runner invocation

- [ ] **Epic**: Implement E2E tests
  - [ ] Secretless broker is tested against Conjur OSS using the new test runner
  - [ ] TBD

# Solution Sign-Off

- **Meeting Date: YYYY/MM/DD**
- **Attendees:**

## Open Questions

Q: What's a good test runner ?

A: Cucumber. It is used in multiple places within the org.

Q: How do we test a test runner ?

A: We don't. However we can test individual step definitions.

Q: How do we test the opaque API for creating dependency resources ?

Q: How will we ensure that improvements we'd like to have in the mid-term aren't put in the icebox forever ?

Q: Where will service lifecycle management implementations reside ?

Q:

Q: Do all Secretless test cases need a full cluster including master, standbys and followers ?

A: Some test cases do not need all the components of the cluster. For example, in testing Conjur authn-k8s and service connection all that is necessary is an API that is able to return secrets to an authenticated client. However, performance tests require a high fidelity service to accurately measure performance. This consideration makes it possible to optimise on startup times for the former test cases and have features like pooling.

## Meeting Notes

TBD

# Addendums

TBD
