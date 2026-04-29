package model

import "time"

type User struct {
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	Username  string    `gorm:"column:username" json:"username"`
	Password  string    `gorm:"column:password" json:"-"`
	AvatarURL string    `gorm:"column:avatar_url" json:"avatar_url"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

type Album struct {
	ID          uint      `gorm:"column:id;primaryKey" json:"id"`
	UserID      uint      `gorm:"column:user_id" json:"user_id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	IsPublic    bool      `gorm:"column:is_public" json:"is_public"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	User        User      `gorm:"foreignKey:UserID;references:ID" json:"-"` //
}

func (Album) TableName() string {
	return "albums"
}

type Photo struct {
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	AlbumID   uint      `gorm:"column:album_id" json:"album_id"`
	FilePath  string    `gorm:"column:file_path" json:"file_path"`
	FileSize  int64     `gorm:"column:file_size" json:"file_size"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Photo) TableName() string {
	return "photos"
}
