# Feature Title
An MSSQL handler that allows client applications to use Secretless to get authenticated connections (via SQL server authentication) to MSSQL instances


- [Current Challenges](#current-challenges)
- [Open Questions](#open-questions)
- [Benefits](#benefits)
- [Risks](#risks)
- [Costs](#costs)
- [Experience (Summary)](#experience-summary)
- [Experience (Detailed)](#experience-detailed)
- [Technical Details](#technical-details)
- [Testing](#testing)
- [Stories](#stories)
- [Future Work](#future-work)

### Current Challenges
- A MSSQL handler with SQL server authentication capabilities
  - A MSSQL Protocol encoder and decoder exists
  - A spec exists for the credentials necessary to authenticate via SQL server (for secretless.yml)

### Open Questions
- What versions of MSSQL should we support and therefore test against ?
- What is the test sample of applications to validate it works
### Experience (Summary)

Given an instance of MSSQL is running and credentials to connect to the target service are stored in a vault supported by Secretless

1. Configure Secretless to proxy an authenticated connection to the MSSQL instance
2. Use client application to consume MSSQL instance via Secretless

### Experience (Detailed)
##### Assumptions
- The use case is for client applications that connect to MSSQL instances using SQL server credentialed-authentication
- Client application will consume from Unix and Windows environments
- The MSSQL target service is already running
- Credentials to connect to the target service are stored in a vault supported by Secretless
- Client application has retry logic
- Client application connects to Secretless without TLS

##### Workflow

1. Create a `secretless.yml` with a listener-handler  combination using the `mssql` protocol. This configures a credential injection proxy to a MSSQL target service
2. Run client application directing DB traffic to the listener in 1

### Technical Details
`<technical details required for developing the above workflow>`

### Testing

1. MSQL Protocol Unit tests to validate the encoding and decoding
2. Handler unit tests to validate credential injection and error propagation
3. Integration tests with a matrix of client applications and, target service versions and deployments to validate client can transparently use Secretless
    + [MSSQL CLI](https://docs.microsoft.com/en-us/sql/tools/mssql-cli?view=sql-server-2017)
    + [C# Application](https://www.codeproject.com/Articles/4416/Beginners-guide-to-accessing-SQL-Server-through-C)

### Benefits
- `<list of benefits>`

### Risks
- `<list of risk>`

### Costs
- `<list of costs>`

### Stories
- `<list of story titles and links to corresponding GH issue>`

### Future Work
- `<list of potential future work>`
