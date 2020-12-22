package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shohinsherov/crud/cmd/app"
	"github.com/shohinsherov/crud/pkg/customers"
	"github.com/shohinsherov/crud/pkg/managers"
	"go.uber.org/dig"
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
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter,
		func() (*pgxpool.Pool, error) {
			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			return pgxpool.Connect(ctx, dsn)
		},
		customers.NewService,
		managers.NewService,
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}
	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}
	err = container.Invoke(func(server *app.Server) {
		server.Init()
	})
	if err != nil {
		return err
	}

	return container.Invoke(func(server *http.Server) error {
		log.Print("server start " + host + ":" + port)
		return server.ListenAndServe()
	})
}
