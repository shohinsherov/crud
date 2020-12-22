package managers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrTokenNotFound = errors.New("token not found")
var ErrNotFound = errors.New("item not found")
var ErrInternal = errors.New("internal error")
var ErrNoSuchUser = errors.New("no such user")
var ErrInvalidPassword = errors.New("invalid password")
var ErrPhoneUsed = errors.New("phone alredy registered")
var ErrTokenExpired = errors.New("token expired")

const (
	ADMIN = "ADMIN"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Auth struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type Product struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Price   int       `json:"price"`
	Qty     int       `json:"qty"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}
type Sales struct {
	ManagerID int64 `json:"manager_id"`
	Total     int   `json:"total"`
}
type Sale struct {
	ID         int64           `json:"id"`
	ManagerID  int64           `json:"manager_id"`
	CustomerID int64           `json:"customer_id"`
	Created    time.Time       `json:"created"`
	Positions  []*SalePosition `json:"positions"`
}

type SalePosition struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	SaleID    int64     `json:"sale_id"`
	Price     int       `json:"price"`
	Qty       int       `json:"qty"`
	Created   time.Time `json:"created"`
}

type Registration struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Phone string   `json:"phone"`
	Roles []string `json:"roles"`
}
type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
	SELECT manager_id FROM managers_tokens WHERE token = $1
	`, token).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Print(err)
		return 0, nil
	}
	if err != nil {
		log.Print(err)
		return 0, nil
	}

	return id, nil
}

func (s *Service) IsAdmin(ctx context.Context, id int64) (ok bool) {
	err := s.pool.QueryRow(ctx, `
	SELECT is_admin FROM managers  WHERE id = $1
	`, id).Scan(&ok)
	if err != nil {
		return false
	}

	return ok
}

func (s *Service) Register(ctx context.Context, reg *Registration) (string, error) {
	var token string
	isAdmin := false
	var id int64

	for _, role := range reg.Roles {
		if role == ADMIN {
			isAdmin = true
		}
	}

	err := s.pool.QueryRow(ctx, `
	INSERT INTO managers(name,phone,is_admin) VALUES ($1,$2,$3) ON CONFLICT (phone) DO NOTHING RETURNING id;
	`, reg.Name, reg.Phone, isAdmin).Scan(&id)
	if err != nil {
		log.Print(err)
		return "", ErrInternal
	}
	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token,manager_id) VALUES($1,$2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}

func (s *Service) Token(
	ctx context.Context,
	phone string, password string,
) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id,password From managers WHERE phone = $1`, phone).Scan(&id, &hash)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidPassword
	}
	if err != nil {
		log.Print(err, "Managers svc 130")
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
	log.Print("id", id)
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token,manager_id) VALUES($1,$2)`, token, id)
	if err != nil {
		log.Print(err)
		return "", ErrInternal
	}

	return token, nil
}

func (s *Service) CreateProduct(ctx context.Context, product *Product) (*Product, error) {
	err := s.pool.QueryRow(ctx, `
	INSERT INTO products(name,qty,price) VALUES ($1,$2,$3) RETURNING id,name,qty,price,active,created;
	`, product.Name, product.Qty, product.Price).Scan(&product.ID, &product.Name, &product.Qty, &product.Price, &product.Active, &product.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return product, nil
}

func (s *Service) UpdateProduct(ctx context.Context, product *Product) (*Product, error) {
	err := s.pool.QueryRow(ctx, `
	UPDATE  products SET  name=$1,qty=$2,price=$3  WHERE id = $4 RETURNING id,name,qty,price,active,created;
	`, product.Name, product.Qty, product.Price, product.ID).Scan(&product.ID, &product.Name, &product.Qty, &product.Price, &product.Active, &product.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return product, nil
}

func (s *Service) MakekSalePosition(ctx context.Context, position *SalePosition) bool {
	active := false
	qty := 0
	err := s.pool.QueryRow(ctx, `
	SELECT qty,active FROM products WHERE id = $1
	`, position.ProductID).Scan(&qty, &active)
	if err != nil {
		return false
	}
	if qty < position.Qty || !active {
		return false
	}
	_, err = s.pool.Exec(ctx, `
	UPDATE products SET qty = $1 WHERE id = $2
	`, qty-position.Qty, position.ProductID)
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

func (s *Service) MakeSale(ctx context.Context, sale *Sale) (*Sale, error) {
	positionsSql := "INSERT INTO sales_positions (sale_id,product_id,qty,price) VALUES "

	err := s.pool.QueryRow(ctx, `
	INSERT INTO sales(manager_id,customer_id) VALUES ($1,$2) RETURNING id, created;
	`, sale.ManagerID, sale.CustomerID).Scan(&sale.ID, &sale.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	for _, position := range sale.Positions {
		if !s.MakekSalePosition(ctx, position) {
			log.Print("Invalid position")
			return nil, ErrInternal
		}
		positionsSql += "(" + strconv.FormatInt(sale.ID, 10) + "," + strconv.FormatInt(position.ProductID, 10) + "," + strconv.Itoa(position.Price) + "," + strconv.Itoa(position.Qty) + "),"
	}

	positionsSql = positionsSql[0 : len(positionsSql)-1] //Remove last comma

	log.Print(positionsSql)
	_, err = s.pool.Exec(ctx, positionsSql)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return sale, nil
}

func (s *Service) GetSales(ctx context.Context, id int64) (sum int, err error) {
	err = s.pool.QueryRow(ctx, `
	SELECT COALESCE(SUM(sp.qty * sp.price),0) total
	FROM managers m
	left JOIN sales s ON s.manager_id= $1
	left JOIN sales_positions sp ON sp.sale_id = s.id 
	GROUP BY m.id
	LIMIT 1`, id).Scan(&sum)
	if err != nil {
		log.Print(err)
		return 0, ErrInternal
	}
	return sum, nil
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

func (s *Service) RemoveProductById(ctx context.Context, id int64) (err error) {
	_, err = s.pool.Exec(ctx, `
	DELETE from products where id = $1`, id)
	if err != nil {
		log.Print(err)
		return ErrInternal
	}
	return nil
}

func (s *Service) RemoveCustomerById(ctx context.Context, id int64) (err error) {
	_, err = s.pool.Exec(ctx, `
	DELETE from customers where id = $1`, id)
	if err != nil {
		log.Print(err)
		return ErrInternal
	}
	return nil
}

func (s *Service) Customers(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE active = TRUE ORDER BY id LIMIT 500
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
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

func (s *Service) ChangeCustomer(ctx context.Context, customer *Customer) (*Customer, error) {
	err := s.pool.QueryRow(ctx, `
	UPDATE customers SET name = $2, phone = $3, active = $4  where id = $1 RETURNING name,phone,active
	`, customer.ID, customer.Name, customer.Phone, customer.Active).Scan(&customer.Name, &customer.Phone, &customer.Active)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return customer, nil
}
