package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CharlesPatterson/todos-app/controller"
	"github.com/CharlesPatterson/todos-app/middleware"
	"github.com/CharlesPatterson/todos-app/model"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/urfave/cli/v2"
)

func runServer() {
	r := gin.Default()
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r.Use(middleware.TimeoutMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)
	version := "/v1"
	v1 := r.Group(version)
	{
		v1.GET("/todos", controller.GetAllTodosHandler)
		v1.PUT("/todos/:id", controller.UpdateTodoHandler)
		v1.POST("/todos", controller.CreateTodoHandler)
		v1.GET("/todos/:id", controller.GetTodoByIdHandler)
		v1.DELETE("/todos/:id", controller.DeleteTodoByIdHandler)
	}
	/*
		{
			r.Group("/swagger/").Handler(http.StripPrefix("/swagger", http.FileServer(http.Dir("swagger"))))
		}
	*/
	port := os.Getenv("PORT")
	r.Run(port)

}

func main() {
	app := &cli.App{
		Name:  "Todos App",
		Usage: "A simple CLI program to manage your todos",
		Action: func(c *cli.Context) error {
			todos, err := model.GetPending()
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
				Usage:   "add a todo to the list",
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

					return model.CreateTodo(todo)
				},
			},
			{
				Name:    "all",
				Aliases: []string{"l"},
				Usage:   "list all todos",
				Action: func(c *cli.Context) error {
					todos, err := model.GetAll()
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
				Usage:   "complete a todo on the list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					return model.CompleteTodo(text)
				},
			},
			{
				Name:    "finished",
				Aliases: []string{"f"},
				Usage:   "list completed todos",
				Action: func(c *cli.Context) error {
					todos, err := model.GetFinished()
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
				Usage:   "deletes a todo on the list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					err := model.DeleteTodo(text)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "starts a server to interact with mongodb",
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
