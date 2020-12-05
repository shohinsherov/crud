package main

import (
	"github.com/gorilla/mux"
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/dig"

	//"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

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

type handler struct {
	mu       *sync.RWMutex
	handlers map[string]http.HandlerFunc
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.mu.RLock()
	handler, ok := h.handlers[request.URL.Path]
	h.mu.RUnlock()

	if !ok {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	handler(writer, request)
	/*_, err := writer.Write([]byte("Hello Bekhai be k"))
	if err != nil {
		log.Print(err)
	}*/
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

	/*// получения указателья на структуру для раборты с БД
	connectCtx, _ := context.WithTimeout(context.Background(), time.Second * 5)
	pool, err := pgxpool.Connect(connectCtx, dsn)
	if err != nil {
		log.Print(err)
		os.Exit(1)
		return
	}
	// закрытие структуры
	defer pool.Close()
	// TODO: запросы


	ctx := context.Background()
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS customers (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		log.Print(err)
		return
	}

	mux := http.NewServeMux()
	customersSvc := customers.NewService(pool)

	server := app.NewServer(mux, customersSvc)
	server.Init()



	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: server,
	}

	log.Print("server start " + host + ":" + port)
	return srv.ListenAndServe()*/
}
