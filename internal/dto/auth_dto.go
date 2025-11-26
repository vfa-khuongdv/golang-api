package dto

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	AccessToken  string `json:"access_token" binding:"required"`
}

type JwtResult struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type LoginResponse struct {
	AccessToken  JwtResult `json:"access_token"`
	RefreshToken JwtResult `json:"refresh_token"`
}
