package entity

type Account struct {
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	Email       string `json:"mail"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password,omitempty" binding:"required"`
	CreatedAt   int64  `json:"-"`
	UpdatedAt   int64  `json:"-"`
	DeletedAt   *int64 `json:"-"`
}
