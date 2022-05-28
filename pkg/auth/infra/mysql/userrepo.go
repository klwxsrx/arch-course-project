package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/auth/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
)

type userRepo struct {
	client mysql.Client
}

func (r *userRepo) NextID() uuid.UUID {
	return uuid.New()
}

func (r *userRepo) GetByLogin(login string) (*service.User, error) {
	const query = "SELECT id, login, encoded_password FROM user WHERE login = ?"

	var userSqlx sqlxUser
	err := r.client.Get(&userSqlx, query, login)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserByLoginNotExists
	}
	if err != nil {
		return nil, err
	}

	return &service.User{
		ID:              userSqlx.ID,
		Login:           userSqlx.Login,
		EncodedPassword: userSqlx.EncodedPassword,
	}, nil
}

func (r *userRepo) Store(user *service.User) error {
	const query = `
		INSERT INTO` + " `user` " + `(id, login, encoded_password, created_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			login = VALUES(login), encoded_password = VALUES(encoded_password), updated_at = NOW()
	`

	binaryUserID, err := user.ID.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = r.client.Exec(query, binaryUserID, user.Login, user.EncodedPassword)
	return err
}

func NewUserRepository(db mysql.Client) service.UserRepository {
	return &userRepo{client: db}
}

type sqlxUser struct {
	ID              uuid.UUID `db:"id"`
	Login           string    `db:"login"`
	EncodedPassword string    `db:"encoded_password"`
}
