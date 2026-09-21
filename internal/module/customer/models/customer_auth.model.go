package models

// ── Customer Auth Request / Response ─────────────────────────────────────────

type CustomerRegisterRequest struct {
	FirstName string  `json:"first_name" binding:"required,min=2,max=65"`
	LastName  string  `json:"last_name"  binding:"required,min=2,max=65"`
	EmailId   string  `json:"email_id"   binding:"required,email"`
	MobileNo  *string `json:"mobile_no"  binding:"omitempty,max=15"`
	Gender    *string `json:"gender"     binding:"omitempty"`
	Password  string  `json:"password"   binding:"required,min=8"`
}

type CustomerLoginRequest struct {
	EmailId  string `json:"email_id" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CustomerRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CustomerAuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Customer     *Customer `json:"customer"`
}

type CustomerRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
