package models

import (
    "time"
)

type Product struct {
    ID          int             `json:"id,omitempty" gorm:"primaryKey"`
    Title       string          `json:"title"`
    Content     string          `json:"content"`
    Price       int             `json:"price,omitempty"`
    CategoryID  int             `json:"categoryId"`
    UserID      string          `json:"userId"`
    Regdate     time.Time       `json:"regdate,omitempty" gorm:"->"`
    Views       int             `json:"views,omitempty"`

    Wishes      int             `json:"wishes,omitempty" gorm:"->"`

    Images      []ProductImage  `json:"images" gorm:"foreignKey:ProductID"`
}

type ProductImage struct {
    ProductID   int             `json:"productId,omitempty"`
    ID          int             `json:"id,omitempty"`
    URL         string          `json:"url"`
    Sequence    int             `json:"sequence"`
}

type ProductValidationResult struct {
    Title       *string         `json:"title,omitempty"`
    Content     *string         `json:"content,omitempty"`
    Price       *string         `json:"price,omitempty"`
}

func (r *ProductValidationResult) GetOrNil() *ProductValidationResult {
    if r.Title == nil && r.Content == nil && r.Price == nil {
        return nil
    }
    return r
}
