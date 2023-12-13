module github.com/cyberark/secretless-broker/bin/juxtaposer

go 1.19

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/go-sql-driver/mysql v1.7.1
	github.com/lib/pq v1.10.6
	github.com/stretchr/testify v1.7.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
)

// Security fixes to ensure we don't have old vulnerable packages in our
// dependency tree. We're often not vulnerable, but removing them to ensure
// we never end up selecting them when other dependencies change.

// Only put specific versions on the left side of the =>
// so we don't downgrade future versions unintentionally.

replace golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2 => golang.org/x/crypto v0.2.0

replace golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c => golang.org/x/crypto v0.2.0

replace golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 => golang.org/x/crypto v0.2.0

replace golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 => golang.org/x/crypto v0.2.0

replace golang.org/x/crypto v0.0.0-20220314234659-1baeb1ce4c0b => golang.org/x/crypto v0.2.0

replace golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d => golang.org/x/crypto v0.2.0

replace golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.0.0-20190620200207-3b0461eec859 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.0.0-20210610132358-84b48f89b13b => golang.org/x/net v0.19.0

replace golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.0.0-20220722155237-a158d28d115b => golang.org/x/net v0.19.0

replace golang.org/x/net v0.0.0-20211209124913-491a49abca63 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.2.0 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.6.0 => golang.org/x/net v0.19.0

replace golang.org/x/net v0.10.0 => golang.org/x/net v0.19.0

replace golang.org/x/text v0.3.0 => golang.org/x/text v0.4.0

replace golang.org/x/text v0.3.3 => golang.org/x/text v0.4.0

replace golang.org/x/text v0.3.6 => golang.org/x/text v0.4.0

replace golang.org/x/text v0.3.7 => golang.org/x/text v0.4.0

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1

replace gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b => gopkg.in/yaml.v3 v3.0.1
