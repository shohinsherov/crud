package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	//"github.com/shohinsherov/crud/cmd/app/middleware"

	"github.com/shohinsherov/crud/pkg/security"

	//"time"

	"github.com/gorilla/mux"

	"github.com/shohinsherov/crud/pkg/customers"
)

// Server предостовляет собой логический сервер нашего приложения
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	securitySvc  *security.Service
}

// Token ...
type Token struct {
	Token string `json:"token"`
}

// Responce ...
type Responce struct {
	CustomerID int64  `json:"customerId"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}

// ResponceOk ...
type ResponceOk struct {
	Status     string `json:"status"`
	CustomerID int64  `json:"customerId"`
}

// ResponceFail ...
type ResponceFail struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// ErrNotFound ...
var ErrNotFound = errors.New("item not found")

// ErrNoSuchUser если пользователь не найден
var ErrNoSuchUser = errors.New("No such user")

// ErrInvalidPassword если пароль не верный
var ErrInvalidPassword = errors.New("Invalid password")

// ErrInternal если происходить другая ошибка
var ErrInternal = errors.New("Internal error")

// ErrExpired ....
var ErrExpired = errors.New("Token is expired")

// NewServer - функция-конструктор для создания сервера.
func NewServer(mux *mux.Router, customersSvc *customers.Service, securitySvc *security.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc, securitySvc: securitySvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

const (
	// GET ....
	GET = "GET"
	// POST ...
	POST = "POST"
	// DELETE ...
	DELETE = "DELETE"
)

// Init инициализирует сервер (регистрирует все Handler-ы)
func (s *Server) Init() {
	//s.mux.Use(s.mdw.Logger)

	s.mux.HandleFunc("/customers", s.handleGetAllCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/active", s.handleGetAllActiveCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/{id}", s.handleCustomerByID).Methods(GET)
	s.mux.HandleFunc("/customers", s.handleSaveCustomer).Methods(POST)
	s.mux.HandleFunc("/customers/{id}", s.handleRemoveByID).Methods(DELETE)
	s.mux.HandleFunc("/customers/{id}/block", s.handleBlockByID).Methods(POST)
	s.mux.HandleFunc("/customers/{id}/block", s.handleUnblockByID).Methods(DELETE)
	//s.mux.Use(middleware.Basic(s.securitySvc.Auth))

	s.mux.HandleFunc("/api/customers", s.saveCustomers).Methods(POST)
	s.mux.HandleFunc("/api/customers/token", s.handleGetToken).Methods(POST)
	s.mux.HandleFunc("/api/customers/token/validate", s.handleValidateToken).Methods(POST)

	//s.mux.HandleFunc("/managers", )

}

func (s *Server) handleGetToken(w http.ResponseWriter, r *http.Request) {
	var auth *security.Auth
	var tok Token
	err := json.NewDecoder(r.Body).Decode(&auth)
	log.Print(auth)
	if err != nil {
		log.Print("Can't Decode login and password")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("Login: ", auth.Login, "Password: ", auth.Password)

	token, err := s.customersSvc.TokenForCustomer(r.Context(), auth.Login, auth.Password)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tok.Token = token
	data, err := json.Marshal(tok)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	var fail ResponceFail
	var ok ResponceOk
	var token Token
	var data []byte
	code := 200

	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		log.Print("Can't Decode token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, er := s.securitySvc.AuthenticateCusomer(r.Context(), token.Token)

	if er == security.ErrNoSuchUser {
		code = 404
		fail.Status = "fail"
		fail.Reason = "not found"
	} else if er == security.ErrExpired {
		code = 400
		fail.Status = "fail"
		fail.Reason = "expired"
	} else if er == nil {
		log.Print(id)
		ok.Status = "ok"
		ok.CustomerID = id
	} else {
		log.Print("err", er)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if code != 200 {
		w.WriteHeader(code)

		data, err = json.Marshal(fail)
		if err != nil {
			log.Print(err)
		}
	} else {
		data, err = json.Marshal(ok)
		if err != nil {
			log.Print(err)
		}
	}
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	return
}

func (s *Server) saveCustomers(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer
	err := json.NewDecoder(request.Body).Decode(&item)
	log.Print(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	cust, err := s.customersSvc.SaveCustomer(request.Context(), item)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(cust)

	data, err := json.Marshal(cust)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Contetn-Type", "applicatrion/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
	http.Error(writer, http.StatusText(200), 200)
	return

}

//  get all
func (s *Server) handleGetAllCustomers(writer http.ResponseWriter, request *http.Request) {
	b, err := s.customersSvc.All(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusNotImplemented), 401)
		return
	}

	data, err := json.Marshal(b)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//	log.Print("ready")
	_, err = writer.Write([]byte(data))
	if err != nil {
		log.Print(err)
	}

}

//  get all active
func (s *Server) handleGetAllActiveCustomers(writer http.ResponseWriter, request *http.Request) {
	b, err := s.customersSvc.AllActive(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	data, err := json.Marshal(b)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//	log.Print("ready")
	_, err = writer.Write([]byte(data))
	if err != nil {
		log.Print(err)
	}

}

// get by id
func (s *Server) handleCustomerByID(writer http.ResponseWriter, request *http.Request) {
	//idParam := request.URL.Query().Get("id")
	idParam, ok := mux.Vars(request)["id"]

	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.customersSvc.ByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Contetn-Type", "applicatrion/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}

}

// add or update
func (s *Server) handleSaveCustomer(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer
	err := json.NewDecoder(request.Body).Decode(&item)
	//log.Print(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	cust, err := s.customersSvc.Save(request.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(cust)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Contetn-Type", "applicatrion/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}

	return

}

// delete customer byID
func (s *Server) handleRemoveByID(writer http.ResponseWriter, request *http.Request) {
	//idParam := request.URL.Query().Get("id")
	idParam, ok := mux.Vars(request)["id"]

	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.customersSvc.RemoveByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

// handleBlockById  выставляет статус active в false
func (s *Server) handleBlockByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.customersSvc.BlockByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// handleUnblockById  вsставлzет статус active в true
func (s *Server) handleUnblockByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.customersSvc.UnblockByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// =====================

// Vars return the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	if rv := r.Context().Value(r); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

/*
func (s *Server) process(writer http.ResponseWriter, request *http.Request) {
	log.Print(request.RequestURI) // полный урл
	log.Print(request.Method)     // метод
	/*	log.Print(request.Header)                     // все заголовки
		log.Print(request.Header.Get("Content-Type")) // конкретный заголовок

		log.Print(request.FormValue("tags"))     // только первое значение Query + POST
		log.Print(request.PostFormValue("tags")) // только первое значение POST ------

	body, err := ioutil.ReadAll(request.Body) // теле запроса
	if err != nil {
		log.Print(err)
	}
	log.Printf("%s", body)

	/*err = request.ParseMultipartForm(10 * 1024 * 1024)  // 10MB
	if err != nil {
		log.Print(err)
	}

	// доступно только после ParseForm (либо FormValue, PostFormValue)
	log.Print(request.Form)     // все значения формы
	log.Print(request.PostForm) // все значения формы

	// доступно только после ParseMultipart (FormValue, PostFromValue автоматически вызывают ParseMultipartForm)
	log.Print(request.FormFile("image"))
	// request.MultipartForm.Value - только "обычные поля"
	// request.MultipartForm.File - только файлы*/

//}
