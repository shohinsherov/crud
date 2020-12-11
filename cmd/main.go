package main

import (
	"context"
	"time"

	"github.com/shohinsherov/crud/pkg/security"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/dig"

	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/shohinsherov/crud/cmd/app"
	"github.com/shohinsherov/crud/pkg/customers"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	// адрес подключения
	//протокол://логин:палоь@хост:порт/бд
	dsn := "postgres://app:postgres@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(host string, port string, dsn string) (err error) {
	//
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter, //http.NewServeMux,
		func() (*pgxpool.Pool, error) {
			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			return pgxpool.Connect(ctx, dsn)
		},
		customers.NewService,
		security.NewService,
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}

	//
	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) {
		server.Init()
	})
	if err != nil {
		log.Print(err)
		return err
	}

	return container.Invoke(func(server *http.Server) error {
		log.Print("server start " + host + ":" + port)
		return server.ListenAndServe()
	})

}
