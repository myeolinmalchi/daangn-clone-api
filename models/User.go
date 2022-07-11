package models

type User struct {
    ID              string      `json:"id,omitempty" gorm:"primaryKey"` 
    PW              string      `json:"pw"`
    Email           string      `json:"email,omitempty"`
    Phone           string      `json:"phone"`
    Name            string      `json:"name,omitempty"`
    Nickname        string      `json:"nickname,omitempty"`
    ProfileImage    string      `json:"profileImage,omitempty"`
}

type UserValidationResult struct {
    PW              *string     `json:"pw"`
    Email           *string     `json:"email"`
    Phone           *string     `json:"phone"`
    Name            *string     `json:"name"`
    Nickname        *string     `json:"nickname"`
}

func (r *UserValidationResult) GetOrNil() *UserValidationResult {
    if r.PW == nil && r.Email == nil && r.Phone == nil && r.Name == nil && r.Nickname == nil {
        return nil
    }
    return r
}
