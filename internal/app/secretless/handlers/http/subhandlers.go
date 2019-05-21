package http

// sub handlers represent different strategy for credential injection
// under a particular protocol
var SubHandlers = map[string]func() HttpSubHandler {
	"aws": func() HttpSubHandler { return AWSHandler{} },
	"conjur": func() HttpSubHandler { return ConjurHandler{} },
	"basic_auth": func() HttpSubHandler { return BasicAuthHandler{} },
}
