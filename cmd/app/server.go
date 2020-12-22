package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shohinsherov/crud/cmd/app/middleware"
	"github.com/shohinsherov/crud/pkg/customers"
	"github.com/shohinsherov/crud/pkg/managers"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	// GET ....
	GET = "GET"
	// POST ...
	POST = "POST"
	// DELETE ...
	DELETE = "DELETE"
)

// Server предостовляет собой логический сервер нашего приложения
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	managersSvc  *managers.Service
}

// Token ...
type Token struct {
	Token string `json:"token"`
}

// NewServer - функция-конструктор для создания сервера.
func NewServer(mux *mux.Router, customersSvc *customers.Service, managersSvc *managers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc, managersSvc: managersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

// Init инициализирует сервер (регистрирует все Handler-ы)
func (s *Server) Init() {
	customersAuthenticateMd := middleware.Authenticate(s.customersSvc.IDByToken)
	customersSubrouter := s.mux.PathPrefix("/api/customers").Subrouter()
	customersSubrouter.Use(customersAuthenticateMd)

	customersSubrouter.HandleFunc("", s.handleCustomerRegistration).Methods(POST)
	customersSubrouter.HandleFunc("/token", s.handleCustomerGetToken).Methods(POST)
	customersSubrouter.HandleFunc("/products", s.handleCustomerGetProducts).Methods(GET)

	managersAuthenticateMd := middleware.Authenticate(s.managersSvc.IDByToken)
	managersSubRouter := s.mux.PathPrefix("/api/managers").Subrouter()
	managersSubRouter.Use(managersAuthenticateMd)
	managersSubRouter.HandleFunc("", s.handleManagerRegistration).Methods(POST)
	managersSubRouter.HandleFunc("/token", s.handleManagerGetToken).Methods(POST)
	managersSubRouter.HandleFunc("/sales", s.handleManagerGetSales).Methods(GET)
	managersSubRouter.HandleFunc("/sales", s.handleManagerMakeSales).Methods(POST)
	managersSubRouter.HandleFunc("/products", s.handleManagerGetProducts).Methods(GET)
	managersSubRouter.HandleFunc("/products", s.handleManagerChangeProducts).Methods(POST)
	managersSubRouter.HandleFunc("/products/{id}", s.handleManagerRemoveProductByID).Methods(DELETE)
	managersSubRouter.HandleFunc("/customers", s.handleManagerGetCustomers).Methods(GET)
	managersSubRouter.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods(POST)
	managersSubRouter.HandleFunc("/customers/{id}", s.handleManagerRemoveCustomerByID).Methods(DELETE)

}
