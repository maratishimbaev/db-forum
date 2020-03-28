package server

import (
	"database/sql"
	"fmt"
	"forum/forum"
	"forum/forum/delivery/http"
	"forum/forum/repository/postgres"
	"forum/forum/usecase"
	"forum/thread"
	"forum/thread/delivery/http"
	"forum/thread/repository/postgres"
	"forum/thread/usecase"
	"forum/user"
	"forum/user/delivery/http"
	"forum/user/repository/postgres"
	"forum/user/usecase"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type App struct {
	userUseCase user.UseCase
	forumUseCase forum.UseCase
	threadUseCase thread.UseCase
}

func NewApp() *App {
	db := initDB()

	userRepository := userPostgres.NewRepository(db)
	forumRepository := forumPostgres.NewRepository(db)
	threadRepository := threadPostgres.NewRepository(db)

	return &App{
		userUseCase: userUseCase.NewUseCase(userRepository),
		forumUseCase: forumUseCase.NewUseCase(forumRepository),
		threadUseCase: threadUsecase.NewUseCase(threadRepository),
	}
}

func initDB() *sql.DB {
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
						  os.Getenv("FORUM_USER"),
						  os.Getenv("FORUM_PASSWORD"),
						  os.Getenv("FORUM_DBNAME"))

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	return db
}

func (a *App) Run() (err error) {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	userHttp.RegisterHTTPEndpoints(e, a.userUseCase)
	forumHttp.RegisterHTTPEndpoints(e, a.forumUseCase)
	threadHttp.RegisterHTTPEndpoints(e, a.threadUseCase)

	return http.ListenAndServe(":8000", e)
}
