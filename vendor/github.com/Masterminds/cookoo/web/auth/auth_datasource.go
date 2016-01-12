package auth

// Authenticate a username/password pair.
//
// This expects a username and a password as strings.
// A boolean `true` indicates that the user has been authenticated. A `false`
// indicates that the user/password combo has failed to auth. This is not
// necessarily an error. An error should only be returned when an unexpected
// condition has obtained during authentication.
type UserDatasource interface {
	AuthUser(username, password string) (bool, error)
}
