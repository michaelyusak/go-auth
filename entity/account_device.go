package entity

type AccountDevice struct {
	DeviceId   int64
	AccountId  int64
	DeviceHash string
	UserAgent  string
	DeviceInfo string
	CreatedAt  int64
	UpdatedAt  int64
	DeletedAt  *int64
}
