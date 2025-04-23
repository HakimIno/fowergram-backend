package errors

type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

var (
	ErrInvalidCredentials = &AuthError{
		Code:    "AUTH001",
		Message: "Invalid username/email or password",
	}
	ErrAccountLocked = &AuthError{
		Code:    "AUTH002",
		Message: "Account is locked due to too many failed attempts",
	}
	ErrInvalidToken = &AuthError{
		Code:    "AUTH003",
		Message: "Invalid or expired token",
	}
	ErrInvalidRefreshToken = &AuthError{
		Code:    "AUTH004",
		Message: "Invalid or expired refresh token",
	}
	ErrUserNotFound = &AuthError{
		Code:    "AUTH005",
		Message: "User not found",
	}
)
