package sqllite

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	"grpcAuthentication/internal/database"
	"grpcAuthentication/internal/domain/models"

	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func New(storagePath string) (*Database, error) {
	const mark = "database.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", mark, err)
	}

	return &Database{db: db}, nil
}

func (d *Database) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const mark = "database.sqlite.SaveUser"

	stmt, err := d.db.Prepare("INSERT INTO users(email,pass_hash) VALUES(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s:%w", mark, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		// module bs, no fix
		var sqliteErr sqlite.Error
		if sqliteErr.Code() == 19 {
			return 0, fmt.Errorf("%s: %w", mark, database.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", mark, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", mark, err)
	}

	return id, nil
}

func (d *Database) User(ctx context.Context, email string) (models.User, error) {
	const mark = "database.sqlite.User"

	stmt, err := d.db.Prepare("SELECT id,email,pass_hash FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s:%w", mark, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", mark, database.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", mark, err)
	}

	return user, nil
}

func (d *Database) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const mark = "database.sqlite.IsAdmin"

	stmt, err := d.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s:%w", mark, err)
	}
	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", mark, database.ErrAppNotFound)
		}

		return false, fmt.Errorf("%s: %w", mark, err)
	}

	return isAdmin, nil
}

func (d *Database) App(ctx context.Context, id int) (models.App, error) {
	const mark = "database.sqlite.App"

	stmt, err := d.db.Prepare("SELECT id,name,secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s:%w", mark, err)
	}
	row := stmt.QueryRowContext(ctx, id)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", mark, database.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", mark, err)
	}

	return app, nil

}
