package models

type User struct {
	ID           string `json:"id,omitempty" gorm:"primaryKey"`
	PW           string `json:"pw"`
	Email        string `json:"email,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	ProfileImage string `json:"profileImage,omitempty"`
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
