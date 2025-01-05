package services

type TwoFactorService interface {
	GenerateTOTP(userID uint) (string, error)
	ValidateTOTP(userID uint, code string) error
	EnableTwoFactor(userID uint) error
	DisableTwoFactor(userID uint) error
}
