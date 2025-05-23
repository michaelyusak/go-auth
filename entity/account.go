package entity

type Account struct {
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password,omitempty" binding:"required"`
	CreatedAt   int64  `json:"-"`
	UpdatedAt   int64  `json:"-"`
	DeletedAt   *int64 `json:"-"`
}

type LoginReq struct {
	Name     string `json:"name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required"`
}

type Token struct {
	Token     string `json:"token"`
	ExpiredAt int64  `json:"expired_at"`
}

type TokenData struct {
	AccessToken  Token `json:"access_token"`
	RefreshToken Token `json:"refresh_token"`
}
