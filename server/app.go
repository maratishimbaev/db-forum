package server

import (
	"database/sql"
	"fmt"
	"forum/forum"
	"forum/forum/delivery/http"
	"forum/forum/repository/postgres"
	"forum/forum/usecase"
	"forum/post"
	postHttp "forum/post/delivery/http"
	postPostgres "forum/post/repository/postgres"
	postUseCase "forum/post/usecase"
	"forum/service"
	serviceHttp "forum/service/delivery/http"
	servicePostgres "forum/service/repository/postgres"
	serviceUsecase "forum/service/usecase"
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
	postUseCase post.UseCase
	serviceUseCase service.UseCase
}

func NewApp() *App {
	db := initDB()

	userRepository := userPostgres.NewRepository(db)
	forumRepository := forumPostgres.NewRepository(db)
	threadRepository := threadPostgres.NewRepository(db)
	postRepository := postPostgres.NewRepository(db)
	serviceRepository := servicePostgres.NewRepository(db)

	return &App{
		userUseCase: userUseCase.NewUseCase(userRepository),
		forumUseCase: forumUseCase.NewUseCase(forumRepository),
		threadUseCase: threadUsecase.NewUseCase(threadRepository),
		postUseCase: postUseCase.NewUseCase(postRepository),
		serviceUseCase: serviceUsecase.NewUseCase(serviceRepository),
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
	postHttp.RegisterHTTPEndpoints(e, a.postUseCase)
	serviceHttp.RegisterHTTPEndpoints(e, a.serviceUseCase)

	return http.ListenAndServe(":8000", e)
}
