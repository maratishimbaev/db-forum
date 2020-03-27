package server

import (
	"database/sql"
	"fmt"
	"forum/forum"
	forumHttp "forum/forum/delivery/http"
	forumPostgres "forum/forum/repository/postgres"
	forumUseCase "forum/forum/usecase"
	"forum/user"
	"forum/user/delivery/http"
	"forum/user/repository/postgres"
	"forum/user/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type App struct {
	userUseCase user.UseCase
	forumUseCase forum.UseCase
}

func NewApp() *App {
	db := initDB()

	userRepository := userPostgres.NewRepository(db)
	forumRepository := forumPostgres.NewRepository(db, userRepository)

	return &App{
		userUseCase: userUseCase.NewUseCase(userRepository),
		forumUseCase: forumUseCase.NewUseCase(forumRepository),
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
	router := gin.Default()

	userHttp.RegisterHTTPEndpoints(router, a.userUseCase)
	forumHttp.RegisterHTTPEndpoints(router, a.forumUseCase)

	return http.ListenAndServe(":8000", router)
}
