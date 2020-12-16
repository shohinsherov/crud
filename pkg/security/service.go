package security

import (
	"context"
	//"crypto/rand"
	//"encoding/hex"
	"errors"
	"log"
	"time"

//	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// ErrNoSuchUser если пользователь не найден
var ErrNoSuchUser = errors.New("No such user")

// ErrInvalidPassword если пароль не верный
var ErrInvalidPassword = errors.New("Invalid password")

// ErrInternal если происходить другая ошибка
var ErrInternal = errors.New("Internal error")
// ErrExpired ....
var ErrExpired = errors.New("Token is expired")

// Service описывает сервис работы с менеджерами.
type Service struct {
	pool *pgxpool.Pool
	//	mu    sync.RWMutex
	//items []*Managers
}
// Auth ....
type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
// NewService создаёт сервис
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Managers представляет информацию о менеджере.
type Managers struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Login      string    `json:"login"`
	Password   string    `json:"password"`
	Salary     int       `json:"salary"`
	Plan       int       `json:"plan"`
	BossID     int64     `json:"boss_id"`
	Department string    `json:"department"`
	Active     bool      `json:"active"`
	Created    time.Time `json:"created"`
}

// Auth ...
func (s *Service) Auth(login, password string) bool {

	log.Print("Go to func Auth")
	log.Print(login, password)

	pass := ""
	ctx := context.Background()
	err := s.pool.QueryRow(ctx, `
		SELECT password FROM managers WHERE login = $1
		`, login).Scan(&pass)

	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}

	if err != nil {
		log.Print(err)
		return false
	}
	log.Print(pass)
	if pass == password {
		return true
	}
	return false
}

// AuthenticateCusomer ...
func (s *Service) AuthenticateCusomer(
	ctx context.Context,
	token string,
) (id int64, err error) {

	expiredTime := time.Now()
	nowTimeInSec := expiredTime.UnixNano()
	err = s.pool.QueryRow(ctx, `SELECT customer_id, expire FROM customers_tokens WHERE token = $1`, token).Scan(&id, &expiredTime)
	if err != nil {
		log.Print(err)
		return 0, ErrNoSuchUser
	}

	if nowTimeInSec > expiredTime.UnixNano() {
		return -1, ErrExpired
	}
	return id, nil
}