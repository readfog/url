package model

import "time"

type Url struct {
    ID         uint       `json:"-" gorm:"primaryKey"`
    ShortCode  string     `json:"short_code" gorm:"size:12;uniqueIndex;not null"`
    OriginUrl  string     `json:"origin_url" gorm:"size:2048;index;not null"`
    Hits       uint       `json:"hits" gorm:"default:0;not null"`
    Deleted    bool       `json:"is_deleted" gorm:"default:false;not null"`
    CreatedAt  time.Time  `json:"-" gorm:"not null"`
    UpdatedAt  time.Time  `json:"-" gorm:"not null"`
    ExpiresOn  time.Time  `json:"expires_on"`
    Keywords   []Keyword  `json:"-" gorm:"many2many:url_keywords"`
}

func (urlModel Url) IsActive() bool {
    if urlModel.Deleted {
        return false
    }

    return urlModel.ExpiresOn.In(time.UTC).After(time.Now().In(time.UTC))
}