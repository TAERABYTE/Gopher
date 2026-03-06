package postgres

import (
	"context"
	"errors"

	"go-minimal-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type noteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(db *pgxpool.Pool) domain.NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(ctx context.Context, note *domain.Note) error {
	query := `INSERT INTO notes (user_id, title, content) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, note.UserID, note.Title, note.Content).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	return err
}

func (r *noteRepository) GetByID(ctx context.Context, id int, userID int) (*domain.Note, error) {
	query := `SELECT id, user_id, title, content, created_at, updated_at FROM notes WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRow(ctx, query, id, userID)

	var note domain.Note
	err := row.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &note, nil
}

func (r *noteRepository) ListByUserID(ctx context.Context, userID int) ([]*domain.Note, error) {
	query := `SELECT id, user_id, title, content, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		var note domain.Note
		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *noteRepository) Update(ctx context.Context, note *domain.Note) error {
	query := `UPDATE notes SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND user_id = $4 RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, note.Title, note.Content, note.ID, note.UserID).Scan(&note.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *noteRepository) Delete(ctx context.Context, id int, userID int) error {
	query := `DELETE FROM notes WHERE id = $1 AND user_id = $2`
	tag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
