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
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/urfave/cli/v2"
)

func runServer() {
	r := gin.New()
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.TimeoutMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)
	r.Static("/assets", "./assets")
	version := "/api/v1"
	v1 := r.Group(version)
	{
		v1.GET("/todos", controller.GetAllTodosHandler)
		v1.PUT("/todos/:id", controller.UpdateTodoByIdHandler)
		v1.POST("/todos", controller.CreateTodoHandler)
		v1.GET("/todos/:id", controller.GetTodoByIdHandler)
		v1.DELETE("/todos/:id", controller.DeleteTodoByIdHandler)
	}
	if os.Getenv("ENVIRONMENT") != "production" {
		r.GET("/", controller.GetRootRedirectHandler)
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
	port := os.Getenv("PORT")
	r.Run(port)

}

// @title Gin Todo API
// @version 1.0
// @description CLI and API for managing TODOs in MongoDB
// @contact.name Charles Patterson
// @contact.url https://github.com/CharlesPatterson/
// @contact.email pattercm@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @query.collection.format multi
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
