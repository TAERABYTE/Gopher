package domain

import (
	"context"
	"time"
)

type Note struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByID(ctx context.Context, id int, userID int) (*Note, error)
	ListByUserID(ctx context.Context, userID int) ([]*Note, error)
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id int, userID int) error
}

type NoteUsecase interface {
	Create(ctx context.Context, input *Note) error
	GetByID(ctx context.Context, id int, userID int) (*Note, error)
	List(ctx context.Context, userID int) ([]*Note, error)
	Update(ctx context.Context, input *Note) error
	Delete(ctx context.Context, id int, userID int) error
}
