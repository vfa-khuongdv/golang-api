package dto

type CreateUserInput struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6,max=255"`
	Name     string  `json:"name" binding:"required,min=1,max=45,not_blank"`     // Name must be between 1-45 chars and not blank
	Birthday *string `json:"birthday" binding:"required,valid_birthday"`         // Assumes birthday is valid format: YYYY-MM-DD
	Address  *string `json:"address" binding:"required,min=1,max=255,not_blank"` // Address must be between 1-255 chars and not blank
	Gender   int16   `json:"gender" binding:"required,oneof=1 2 3"`
	RoleIds  []uint  `json:"role_ids" binding:"required,min=1,dive,required"` // RoleIds must be a non-empty array of uints
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	Password    string `json:"password" binding:"required,min=6,max=255"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=255"`
}

type ChangePasswordInput struct {
	OldPassword     string `json:"old_password" binding:"required,min=6,max=255"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=255"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6,max=255"`
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
