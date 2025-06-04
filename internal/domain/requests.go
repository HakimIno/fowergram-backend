package domain

type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=32"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	Password  string `json:"password" validate:"required"`
	BirthDate string `json:"birth_date,omitempty" validate:"omitempty"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Username            string `json:"username" validate:"omitempty,min=3,max=32"`
	Email               string `json:"email" validate:"omitempty,email"`
	CurrentPassword     string `json:"current_password" validate:"required_with=NewPassword"`
	NewPassword         string `json:"new_password" validate:"omitempty,min=8"`
	NotificationEnabled *bool  `json:"notification_enabled"`
}

type CreatePostRequest struct {
	Caption  string `json:"caption" validate:"required"`
	ImageURL string `json:"image_url" validate:"required,url"`
}
