package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shohinsherov/crud/pkg/customers"
	
)

func (s *Server) handleCustomerRegistration(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Registration
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	saved, err := s.customersSvc.Register(request.Context(), item)
	data, err := json.Marshal(saved)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleCustomerGetToken(writer http.ResponseWriter, request *http.Request) {
	var auth *customers.Auth
	err := json.NewDecoder(request.Body).Decode(&auth)
	fmt.Print(auth)
	if err != nil {
		log.Print("Can't Decode login and password")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("Login: ", auth.Login, "Password: ", auth.Password)

	token, err := s.customersSvc.Token(request.Context(), auth.Login, auth.Password)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(&Token{Token: token})
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (s *Server) handleCustomerGetProducts(writer http.ResponseWriter, request *http.Request) {
	items, err := s.customersSvc.Products(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}


