package ctxtypes

// ContextKey is, as the name implied, a type reserved
// for keys when passing values into the context
type ContextKey string

// PreLoginResponseKey is used to obtain PreLogin Response fields
const PreLoginResponseKey ContextKey = "preLoginResponse"
