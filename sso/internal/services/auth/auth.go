package auth

import (
	"context"
	"errors"
	"fmt"
	"grpcAuthentication/internal/database"
	"grpcAuthentication/internal/domain/models"
	"grpcAuthentication/internal/jwt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

// New returns a new instance of the Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		userSaver:    userSaver,
		userProvider: userProvider,
		log:          log,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const mark = "auth.Login"
	// Could avoid logging email
	log := a.log.With(slog.String("mark", mark), slog.String("email", email))

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			a.log.Warn("user not found", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})

			return "", fmt.Errorf("%s:%w", mark, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})

		return "", fmt.Errorf("%s:%w", mark, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})

		return "", fmt.Errorf("%s: %w", mark, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s:%w", mark, err)
	}

	log.Info("user logged in")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		return "", fmt.Errorf("%s:%w", mark, err)
	}
	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const mark = "auth.RegisterNewUser"
	// Could avoid logging email
	log := a.log.With(slog.String("mark", mark), slog.String("email", email))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		return 0, fmt.Errorf("%s:%w", mark, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, database.ErrUserExists) {
			log.Warn(("user already exists"))

			return 0, fmt.Errorf("%s:%w", mark, database.ErrUserExists)
		}
		log.Error("failed to save user", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		return 0, fmt.Errorf("%s:%w", mark, err)
	}
	log.Info("user registered")

	return id, nil
}

// func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
// 	const mark = "auth.IsAdmin"

// 	log := a.log.With(
// 		slog.String("mark", mark),
// 		slog.Int64("user_id", userID),
// 	)

// 	log.Info("checking if user is admin")

// 	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
// 	if err != nil {
// 		return false, fmt.Errorf("%s: %w", mark, err)
// 	}

// 	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

// 	return isAdmin, nil
// }
