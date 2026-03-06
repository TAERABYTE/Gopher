package usecase

import (
	"context"

	"go-minimal-backend/internal/domain"
)

type noteUsecase struct {
	noteRepo domain.NoteRepository
}

func NewNoteUsecase(nr domain.NoteRepository) domain.NoteUsecase {
	return &noteUsecase{
		noteRepo: nr,
	}
}

func (u *noteUsecase) Create(ctx context.Context, note *domain.Note) error {
	return u.noteRepo.Create(ctx, note)
}

func (u *noteUsecase) GetByID(ctx context.Context, id int, userID int) (*domain.Note, error) {
	return u.noteRepo.GetByID(ctx, id, userID)
}

func (u *noteUsecase) List(ctx context.Context, userID int) ([]*domain.Note, error) {
	return u.noteRepo.ListByUserID(ctx, userID)
}

func (u *noteUsecase) Update(ctx context.Context, input *domain.Note) error {
	// Let repo handle finding by id and userid
	return u.noteRepo.Update(ctx, input)
}

func (u *noteUsecase) Delete(ctx context.Context, id int, userID int) error {
	return u.noteRepo.Delete(ctx, id, userID)
}
