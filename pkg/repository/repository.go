package repository

import (
	"AuthService/pkg"
	"context"
	"database/sql"
)

type Repository interface {
	IsExists(ctx context.Context, userId string) (bool, error)
	AddSession(ctx context.Context, user pkg.User, refreshHash string) error
	UpdateSession(ctx context.Context, refreshHash, newRefreshHash string) error
	GetSessions(ctx context.Context, userId string) ([]pkg.Session, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func New(db *sql.DB) *repositoryImpl {
	return &repositoryImpl{
		db: db,
	}
}

func (r *repositoryImpl) IsExists(ctx context.Context, userId string) (bool, error) {
	query := `SELECT 1 FROM users WHERE id = $1`
	var isExists bool
	if err := r.db.QueryRowContext(ctx, query, userId).Scan(&isExists); err != nil {
		return false, err
	}
	return isExists, nil
}

func (r *repositoryImpl) UpdateSession(ctx context.Context, refreshHash, newRefreshHash string) error {
	query := `UPDATE sessions SET refresh_hash = $1 WHERE refresh_hash = $2;`
	if _, err := r.db.ExecContext(ctx, query, newRefreshHash, refreshHash); err != nil {
		return err
	}
	return nil
}

func (r *repositoryImpl) AddSession(ctx context.Context, user pkg.User, refreshHash string) error {
	ttx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			ttx.Rollback()
		}
	}()

	query := `INSERT INTO sessions (refresh_hash, ip) VALUES ($1, $2) RETURNING id`
	var id int
	if err = ttx.QueryRowContext(ctx, query, refreshHash, user.IP).Scan(&id); err != nil {
		return err
	}
	query = `INSERT INTO user_session (user_id, session_id) VALUES ($1, $2)`
	if _, err = ttx.ExecContext(ctx, query, user.UserId, id); err != nil {
		return err
	}
	if err = ttx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *repositoryImpl) GetSessions(ctx context.Context, userId string) ([]pkg.Session, error) {
	query := `SELECT * FROM session_view WHERE user_id = $1`
	var sessions []pkg.Session
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var session pkg.Session
		if err = rows.Scan(&session.RefreshHash, &session.UserId, &session.UserIP,
			&session.UserEmail, &session.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
