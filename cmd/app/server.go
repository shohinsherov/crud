package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	//"time"

	"github.com/gorilla/mux"

	"github.com/shohinsherov/crud/pkg/customers"
)

// Server предостовляет собой логический сервер нашего приложения
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
}

// NewServer - функция-конструктор для создания сервера.
func NewServer(mux *mux.Router, customersSvc *customers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc}
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

	s.mux.HandleFunc("/customers", s.handleGetAllActiveCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/active", s.handleGetAllActiveCustomers)
	s.mux.HandleFunc("/customers/{id}", s.handleCustomerByID).Methods(GET)
	s.mux.HandleFunc("/customers", s.handleSaveCustomer).Methods(POST)
	s.mux.HandleFunc("/customers/{id}", s.handleRemoveByID).Methods(DELETE)
	s.mux.HandleFunc("/customers/{id}/block", s.handleBlockByID).Methods(POST)
	s.mux.HandleFunc("/customers/{id}/unblock", s.handleUnblockByID).Methods(POST)

}

//  get all
func (s *Server) handleGetAllCustomers(writer http.ResponseWriter, request *http.Request) {
	b, err := s.customersSvc.All(request.Context())
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
