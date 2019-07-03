package testutil

// StringPointer simply returns a pointer to the given string.
// TODO: move this to a util folder somewhere
func StringPointer(s string) *string {
	return &s
}
