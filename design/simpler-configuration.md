# Secretless configuration is intuitive and simple
The Secretless configuration has been the same since the start of the project, but since the start of the project
we have learned more about the internal concepts, how they link together, and how to name them so that they are clear
to end-users and people who are learning about the project for the first time.

In this design proposal, we suggest an alternate Secretless configuration syntax that will make it simpler to configure
a Secretless Broker sidecar. We propose making these updates in a way that we will continue to support the old configuration
syntax at a low cost, while improving the test coverage for the configuration package and clarifying / simplifying the
project documentation.

Aha Card: https://cyberark.aha.io/features/AAM-<>

- [Objective](#objective)
- [Experience (Summary)](#experience-summary)
- [Experience (Detailed)](#experience-detailed)
- [Technical Details](#technical-details)
- [Testing](#testing)
- [Documentation](#documentation)
- [Open Questions](#open-questions)
- [Stories](#stories)
- [Future Work](#future-work)

### Objective
- Make it simpler to configure Secretless Broker, including using language in the configuration spec that is more intuitive
- Improve the Secretless documentation by simplifying the concepts one needs to understand in order to make sense of how
  the project works
- Stabilize the configuration package so that end users can expect no near-term changes will need to be made to the
  configuration beyond the next stable release

### Experience 
#### Current Experience


#### (Summary)
1. `<high level summary of steps in workflow>`

### Experience (Detailed)
##### Assumptions
- `<list of assumptions being made>`

##### Workflow
1. `<step by step flow of configuring and using this feature>`

### Technical Details
`<technical details that enhance an understanding of the above workflow>`

#### Dependent Components
1. [Sidecar Injector](https://github.com/cyberark/sidecar-injector)
1. [Configuration CRD](https://github.com/cyberark/secretless-broker/tree/master/internal/app/secretless/configurationmanagers/kubernetes/crd)
1. [File Configuration Watcher](https://github.com/cyberark/secretless-broker/blob/master/internal/app/secretless/configurationmanagers/configfile/fs_watcher.go)
1. [Demos](https://github.com/cyberark/secretless-broker/tree/master/demos/)
1. Documentation and tutorials
1. Architecture diagrams
1. All existing integration and end-to-end tests

### Testing
`<overview of unit and integration tests to be added>`
`<Do the tests ensure privileged operations are verified, there is no leakage of secrets or privileged data, and all privileged operations on the component leave an audit trail?>`

### Open Questions
- How should Secretless fail on invalid configuration when watching?

### Stories
#### Development
- #708 - Configuration parsing in config package handles new config format
- #710 - New configuration object is proxied to existing listeners and handlers
- #714 - Old secretless.yml are still supported to ease transition to new format. 

#### Testing
- #711 - Tests still pass after secretless.yml config changes are done
- #712 - All Secretless test cases use new yml config format

#### Documentation
- #713 - All Secretless configuration code samples are updated to reflect new yml config format
- cyberark/secretless-docs#135 - Listener / Handler references are updated to refer to single handler construct

### Future Work
- We have planned future work to further simplify and improve the plugin interface. This will be done in a follow-on effort.
- Update the configuration CRD to follow the new model (#715)
