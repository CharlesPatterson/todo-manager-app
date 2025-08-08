package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	golangtodomanager "github.com/CharlesPatterson/todos-app"
	"github.com/CharlesPatterson/todos-app/controller"
	docs "github.com/CharlesPatterson/todos-app/docs"
	"github.com/CharlesPatterson/todos-app/middleware"
	"github.com/CharlesPatterson/todos-app/model"
	jwt "github.com/appleboy/gin-jwt/v2"
	cache "github.com/chenyahui/gin-cache"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/urfave/cli/v2"
)

// @Summary	Login
// @ID			login
// @Tags		Auth
// @Produce	json
// @Param		data	body		middleware.Login	true	"Login credentials"
// @Success	200		{object}	model.Todo
// @Router		/login [post]
func runServer() {
	cacheConfig := model.SetupRedisCache()

	r := gin.New()
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.TimeoutMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	err := r.SetTrustedProxies(nil)
	if err != nil {
		return
	}
	authMiddleware, err := jwt.New(middleware.InitJWTParams())
	r.Use(middleware.HandlerMiddleware(authMiddleware))
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	r.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 page not found"})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, "")
	})
	r.GET("/readyz", func(c *gin.Context) {
		redisStatusError := cacheConfig.Store.RedisClient.Ping(c).Err()
		if redisStatusError != nil {
			c.JSON(500, "Redis is unreachable")
		}
		mongoStatusError := model.Collection.Database().Client().Ping(c, readpref.Primary())
		if mongoStatusError != nil {
			c.JSON(500, "MongoDB is unreachable")
		}
		c.JSON(200, "")
	})

	r.Static("/assets", "./assets")
	version := "/api/v1"
	r.POST("/api/v1/login", authMiddleware.LoginHandler)
	auth := r.Group("/auth", authMiddleware.MiddlewareFunc())
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	v1 := r.Group(version, authMiddleware.MiddlewareFunc())
	{
		v1.GET("/todos", cache.CacheByRequestURI(cacheConfig.Store, cacheConfig.DefaultCacheTime), controller.GetAllTodosHandler)
		v1.PUT("/todos/:id", controller.UpdateTodoByIdHandler)
		v1.POST("/todos", controller.CreateTodoHandler)
		v1.GET("/todos/:id", cache.CacheByRequestURI(cacheConfig.Store, cacheConfig.DefaultCacheTime), controller.GetTodoByIdHandler)
		v1.DELETE("/todos/:id", controller.DeleteTodoByIdHandler)
	}
	if os.Getenv("ENVIRONMENT") != "production" {
		authorized := r.Group("/")
		authorized.Use(middleware.BasicAuthMiddleware())
		{
			authorized.GET("/", controller.GetRootRedirectHandler)
			authorized.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		}
	}
	port := os.Getenv("PORT")
	err = r.Run(port)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

// @title						Gin Todo API
// @version					1.0
// @description				CLI and API for managing TODOs in MongoDB
// @contact.name				Charles Patterson
// @contact.url				https://github.com/CharlesPatterson/
// @contact.email				pattercm@gmail.com
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
// @host						localhost:8080
// @BasePath					/api/v1
// @schemes					http https
// @query.collection.format	multi
// @securityDefinitions.apiKey	JWT
// @in							header
// @name						Authorization
func main() {

	app := &cli.App{
		Version: golangtodomanager.Version,
		Name:    "Todos App",
		Usage:   "A simple CLI program to manage your todos",
		Action: func(c *cli.Context) error {
			var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			todos, err := model.GetPending(ctx)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					fmt.Print("Nothing to see here.\nRun `add 'todo'` to add a todo")
					return nil
				}
				return err
			}

			model.PrintTodos(todos)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add a todo to the list",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						return errors.New("cannot add an empty todo")
					}

					todo := &model.Todo{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Text:      str,
						Completed: false,
					}
					var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					return model.CreateTodo(ctx, todo)
				},
			},
			{
				Name:    "all",
				Aliases: []string{"l"},
				Usage:   "List all todos",
				Action: func(c *cli.Context) error {
					var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					todos, err := model.GetAll(ctx)
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `add 'todo'` to add a todo")
							return nil
						}

						return err
					}
					model.PrintTodos(todos)
					return nil
				},
			},
			{
				Name:    "done",
				Aliases: []string{"d"},
				Usage:   "Complete a todo on the list",
				Action: func(c *cli.Context) error {
					var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					text := c.Args().First()
					return model.CompleteTodo(ctx, text)
				},
			},
			{
				Name:    "finished",
				Aliases: []string{"f"},
				Usage:   "List completed todos",
				Action: func(c *cli.Context) error {
					var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					todos, err := model.GetFinished(ctx)
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `add 'todo'` to add a todo")
							return nil
						}
						return err
					}

					model.PrintTodos(todos)
					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"rm"},
				Usage:   "Deletes a todo on the list",
				Action: func(c *cli.Context) error {
					var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					text := c.Args().First()
					err := model.DeleteTodo(ctx, text)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Starts a server to interact with mongodb",
				Action: func(c *cli.Context) error {
					runServer()
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
