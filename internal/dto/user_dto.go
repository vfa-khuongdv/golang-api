package dto

type CreateUserInput struct {
	Email    string  `json:"email" binding:"required,email"`                     // Email must be valid format
	Password string  `json:"password" binding:"required,min=6,max=255"`          // Password must be between 6-255 chars
	Name     string  `json:"name" binding:"required,min=1,max=45,not_blank"`     // Name must be between 1-45 chars and not blank
	Birthday *string `json:"birthday" binding:"required,valid_birthday"`         // Assumes birthday is valid format: YYYY-MM-DD
	Address  *string `json:"address" binding:"required,min=1,max=255,not_blank"` // Address must be between 1-255 chars and not blank
	Gender   int16   `json:"gender" binding:"required,oneof=1 2 3"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"` // Email must be valid format
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`                      // Token is required
	NewPassword string `json:"new_password" binding:"required,min=6,max=255"` // New password must be between 6-255 chars
}

type ChangePasswordInput struct {
	OldPassword     string `json:"old_password" binding:"required,min=6,max=255"`     // Old password must be between 6-255 chars
	NewPassword     string `json:"new_password" binding:"required,min=6,max=255"`     // New password must be between 6-255 chars
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6,max=255"` // Confirm password must be between 6-255 chars
}

type UpdateUserInput struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=45,not_blank"`     // Name must be between 1-45 chars and not blank
	Birthday *string `json:"birthday" binding:"omitempty,valid_birthday"`         // Assumes birthday is valid format: YYYY-MM-DD
	Address  *string `json:"address" binding:"omitempty,min=1,max=255,not_blank"` // Address must be between 1-255 chars and not blank
	Gender   *int16  `json:"gender" binding:"omitempty,oneof=1 2 3"`              // Gender must be one of [1 2 3]
}

type UpdateProfileInput struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=45,not_blank"`     // Name must be between 1 and 45 characters and not blank if provided
	Birthday *string `json:"birthday" binding:"omitempty,valid_birthday"`         // Birthday must be a valid date (YYYY-MM-DD) if provided
	Address  *string `json:"address" binding:"omitempty,min=1,max=255,not_blank"` // Address must be between 1 and 255 characters and not blank if provided
	Gender   *int16  `json:"gender" binding:"omitempty,oneof=1 2 3"`              // Gender must be 1, 2, or 3 if provided
}
