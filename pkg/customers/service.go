package customers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ErrTokenNotFound ...
var ErrTokenNotFound = errors.New("token not found")

// ErrNotFound возвращается, когда покупатель не найден.
var ErrNotFound = errors.New("item not found")

// ErrInternal возвращаетсяб когда произошла внутренная ошибка.
var ErrInternal = errors.New("internal error")

// ErrNoSuchUser если пользователь не найден
var ErrNoSuchUser = errors.New("No such user")

// ErrInvalidPassword если пароль не верный
var ErrInvalidPassword = errors.New("Invalid password")

// ErrPhoneUsed ...
var ErrPhoneUsed = errors.New("phone alredy registered")

// ErrTokenExpired ...
var ErrTokenExpired = errors.New("token expired")

// Service описывает сервис работы с покупателями.
type Service struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	items []*Customer
}

// NewService создаёт сервис
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Customer представляет информацию о покупателе.
type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

type Registration struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type Product struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Qty   int    `json:"qty`
}

func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, `
	SELECT id,name, phone, active, created FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

func (s *Service) All(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)
	rows, err := s.pool.Query(ctx, `
	SELECT id,name, phone, active, created FROM customers ORDER BY id
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	for rows.Next() {
		item := &Customer{}
		rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) AllActive(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)
	rows, err := s.pool.Query(ctx, `
	SELECT id,name, phone, active, created FROM customers WHERE active= true ORDER BY id;
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	for rows.Next() {
		item := &Customer{}
		rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) Register(ctx context.Context, item *Registration) (*Customer, error) {
	customer := &Customer{}
	hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	log.Print(hex.EncodeToString(hash))
	err = s.pool.QueryRow(ctx, `
	INSERT INTO customers(name,phone,password) VALUES ($1,$2,$3) ON CONFLICT (phone) DO NOTHING RETURNING id, name, phone, active, created;
	`, item.Name, item.Phone, hash).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Active, &customer.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return customer, nil
}

func (s *Service) Update(ctx context.Context, item *Customer) (*Customer, error) {
	customer := &Customer{
		ID:    item.ID,
		Name:  item.Name,
		Phone: item.Phone,
	}

	err := s.pool.QueryRow(ctx, `
	UPDATE customers SET name =$1,phone=$2 WHERE id =$3 RETURNING active,created
	`, item.Name, item.Phone, item.ID).Scan(&customer.Active, &customer.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return customer, nil
}

func (s *Service) RemoveByID(ctx context.Context, id int64) (*Customer, error) {
	customer := &Customer{}
	err := s.pool.QueryRow(ctx, `
	DELETE FROM customers WHERE id= $1 RETURNING id,name,phone,active,created
	`, id).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Active, &customer.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return customer, nil
}

func (s *Service) BlockByID(ctx context.Context, id int64) (*Customer, error) {
	customer := &Customer{}
	err := s.pool.QueryRow(ctx, `
	UPDATE customers SET active= false WHERE id= $1 RETURNING id,name,phone,active,created
	`, id).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Active, &customer.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return customer, nil
}

func (s *Service) UnBlockByID(ctx context.Context, id int64) (*Customer, error) {
	customer := &Customer{}
	err := s.pool.QueryRow(ctx, `
	UPDATE customers SET active= true WHERE id= $1 RETURNING id,name,phone,active,created
	`, id).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Active, &customer.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return customer, nil
}

func (s *Service) Token(
	ctx context.Context,
	phone string, password string,
) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id,password From customers WHERE phone = $1`, phone).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", ErrInvalidPassword
	}
	if err != nil {
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}
	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token,customer_id) VALUES($1,$2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}

func (s *Service) Products(ctx context.Context) ([]*Product, error) {
	items := make([]*Product, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, price, qty FROM products WHERE active = TRUE ORDER BY id LIMIT 500
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
	SELECT customer_id FROM customers_tokens WHERE token = $1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, ErrInternal
	}

	return id, nil
}
