package models

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	PassHash  []byte    `db:"pass_hash"`
	Name      string    `db:"name"`
	Surname   string    `db:"surname"`
	AvatarKey string    `db:"avatar_key"`
}
