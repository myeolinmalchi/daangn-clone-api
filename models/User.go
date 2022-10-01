package models

type User struct {
	ID           string   `json:"id,omitempty" gorm:"primaryKey"`
	PW           string   `json:"pw"`
	Email        string   `json:"email,omitempty"`
	Nickname     string   `json:"nickname,omitempty"`
	ProfileImage string   `json:"profileImage,omitempty"`
	Devices      []Device `json:"deivce" gorm:"foreignKey:UserID"`
}

type DeviceType string

const (
	IOS     DeviceType = "IOS"
	ANDROID DeviceType = "ANDROID"
)

type Device struct {
	UserID     string     `json:"userId"`
	ID         int        `json:"id" gorm:"primaryKey"`
	Token      string     `json:"token"`
	DeviceType DeviceType `json:"deviceType"`
}

type UserValidationResult struct {
	PW       *string `json:"pw,omitempty"`
	Email    *string `json:"email,omitempty"`
	Nickname *string `json:"nickname,omitempty"`
}

func (r *UserValidationResult) GetOrNil() *UserValidationResult {
	if r.PW == nil && r.Email == nil && r.Nickname == nil {
		return nil
	}
	return r
}

func (r *UserValidationResult) Test() bool {
	return r.PW == nil && r.Email == nil && r.Nickname == nil
}
