package customers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"
)

// ErrNotFound возвращается, когда покупатель не найден.
var ErrNotFound = errors.New("item not found")

// ErrInternal возвращаетсяб когда произошла внутренная ошибка.
var ErrInternal = errors.New("internal error")

// Service описывает сервис работы с покупателями.
type Service struct {
	db    *sql.DB
	mu    sync.RWMutex
	items []*Customer
}

// NewService создаёт сервис
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Customer представляет информацию о покупателе.
type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

// ByID возвращает покупателья по индетификатору.
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item := &Customer{}

	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, phone, active, created FROM customers where id = $1`,
		id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}

// All возврашает все существующие покупателей
func (s *Service) All(ctx context.Context) ([]*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	//создаём слайс для хранения результатов
	items := make([]*Customer, 0)

	//делаем запрос
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, phone, active, created FROM customers
	`)

	//проверяем ошибки
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// rows нужно закрывать
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	// rows.Next() возвращает true до тех пор, пока дальше есть строки
	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	// в конце нужно проверять общую ошибку
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// AllActive возврашает всех активных покупателей
func (s *Service) AllActive(ctx context.Context) ([]*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	//создаём слайс для хранения результатов
	items := make([]*Customer, 0)

	//делаем запрос
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, phone, active, created FROM customers where active
	`)

	//проверяем ошибки
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// rows нужно закрывать
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	// rows.Next() возвращает true до тех пор, пока дальше есть строки
	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	// в конце нужно проверять общую ошибку
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// Save save/update
func (s *Service) Save(ctx context.Context, item *Customer) (*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := &Customer{}
	if item.ID == 0 {
		err := s.db.QueryRowContext(ctx, `
			INSERT INTO customers(name, phone) VALUES ($1, $2) ON CONFLICT (phone) DO UPDATE SET name = excluded.name, active = excluded.active, created = excluded.created 
			RETURNING id, name, phone, active, created
		`, item.Name, item.Phone).Scan(&res.ID, &res.Name, &res.Phone, &res.Active, &res.Created)

		if err != nil {
			log.Print(err)
			return nil, err
		}
		return res, nil
	}

	_, err := s.ByID(ctx, item.ID)
	if err != nil {
		log.Print(err)
		return nil, ErrNotFound
	}
	err = s.db.QueryRowContext(ctx, `
			UPDATE customers SET name = $1, phone = $2, active = $3, created = $4 where id = $5 RETURNING id, name, phone, active, created
		`, item.Name, item.Phone, item.Active, item.Created, item.ID).Scan(&res.ID, &res.Name, &res.Phone, &res.Active, &res.Created)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res, nil
}

// RemoveByID удаляет баннер по идентификатору
func (s *Service) RemoveByID(ctx context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := s.ByID(ctx, id)
	if err != nil {
		log.Print(err)
		return ErrNotFound
	}
	_, err = s.db.ExecContext(ctx, `
		DELETE FROM customers where id = $1;
	`, id)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// BlockByID выставляет статус active в false
func (s *Service) BlockByID(ctx context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := s.ByID(ctx, id)
	if err != nil {
		log.Print(err)
		return ErrNotFound
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE customers SET active = false where id = $1
	`, id)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// UnblockByID выставляет статус active в false
func (s *Service) UnblockByID(ctx context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := s.ByID(ctx, id)
	if err != nil {
		log.Print(err)
		return ErrNotFound
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE customers SET active = true where id = $1
	`, id)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
