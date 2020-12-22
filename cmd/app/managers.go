package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shohinsherov/crud/cmd/app/middleware"
	"github.com/shohinsherov/crud/pkg/managers"
)

func (s *Server) handleManagerRegistration(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User by id:", id, "is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if !s.managersSvc.IsAdmin(request.Context(), id) {
		log.Print("manager by id:", id, "is not admin")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	var registration *managers.Registration
	err = json.NewDecoder(request.Body).Decode(&registration)
	if err != nil {
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	token, err := s.managersSvc.Register(request.Context(), registration)
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

func (s *Server) handleManagerGetToken(writer http.ResponseWriter, request *http.Request) {
	var auth *managers.Auth
	err := json.NewDecoder(request.Body).Decode(&auth)
	fmt.Print(auth)
	if err != nil {
		log.Print("Can't Decode login and password")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("Login: ", auth.Phone, "Password: ", auth.Password)

	token, err := s.managersSvc.Token(request.Context(), auth.Phone, auth.Password)
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

func (s *Server) handleManagerChangeProducts(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	product := &managers.Product{}
	err = json.NewDecoder(request.Body).Decode(&product)
	fmt.Print(product)
	if err != nil {
		log.Print("Can't Decode product")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if product.ID == 0 {
		product, err = s.managersSvc.CreateProduct(request.Context(), product)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		product, err = s.managersSvc.UpdateProduct(request.Context(), product)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	data, err := json.Marshal(product)
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

func (s *Server) handleManagerMakeSales(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	sale := &managers.Sale{}
	sale.ManagerID = id
	err = json.NewDecoder(request.Body).Decode(&sale)
	fmt.Print(sale)
	fmt.Print(sale.Positions)
	if err != nil {
		log.Print("Can't Decode product")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sale, err = s.managersSvc.MakeSale(request.Context(), sale)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(sale)
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

func (s *Server) handleManagerGetSales(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	total, err := s.managersSvc.GetSales(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(&managers.Sales{ManagerID: id, Total: total})
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

func (s *Server) handleManagerGetProducts(writer http.ResponseWriter, request *http.Request) {
	items, err := s.managersSvc.Products(request.Context())
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

func (s *Server) handleManagerRemoveProductByID(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		log.Print("Missing id")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	productID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = s.managersSvc.RemoveProductById(request.Context(), productID)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (s *Server) handleManagerRemoveCustomerByID(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		log.Print("Missing id")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	customerID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = s.managersSvc.RemoveCustomerById(request.Context(), customerID)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (s *Server) handleManagerGetCustomers(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Println("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	items, err := s.managersSvc.Customers(request.Context())
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

func (s *Server) handleManagerChangeCustomer(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if id == 0 {
		log.Print("User is not manager")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	customer := &managers.Customer{}
	err = json.NewDecoder(request.Body).Decode(&customer)
	fmt.Println(customer)
	if err != nil {
		log.Print("Can't Decode customer")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	customer, err = s.managersSvc.ChangeCustomer(request.Context(), customer)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(customer)
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
