package domain

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
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
