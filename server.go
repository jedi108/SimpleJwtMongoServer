package main

import (
	"SimpleJwtMongoServer/config"
	"SimpleJwtMongoServer/handlers"
	"SimpleJwtMongoServer/models"
	"log"
	"os"
)

func configServices() {
	cfg := config.New()
	err := cfg.FileNameFromArgs(os.Args)
	if err != nil {
		panic(err)
	}

	err = cfg.LoadFromFile()
	if err != nil {
		panic(err)
	}

	configDb := &models.ConfigMongo{}
	err = cfg.ApplyStruct(&configDb)
	if err != nil {
		panic(err)
	}

	env := handlers.Env{}
	c, err := models.NewMongoDbCollection(configDb)
	if err != nil {
		panic(err)
	}
	env.MongoCollection = c

	configHandles := handlers.Handlers{}
	err = cfg.ApplyStruct(&configHandles)
	if err != nil {
		panic(err)
	}

	// user := models.User{Email: "test@test.com"}
	// q := models.FindByEmail(env.MongoCollection, &user)
	// count, err := q.Count()
	// if err != nil {
	// 	panic(err)
	// }
	// spew.Dump(count)
	// os.Exit(1)

	// user = models.User{Email: "test@test1.com"}
	// userData, err := models.GetUserByEmail(env.MongoCollection, &user)
	// if err != nil {
	// 	panic(err)
	// }
	// spew.Dump(userData)
	// os.Exit(000)

	// tokens := []string{"aaa", "vvv", "ccc"}
	// user = models.User{Email: "test@test.com", Tokens: tokens}
	// models.Insert(env.MongoCollection, &user)
	// os.Exit(000)

	log.Printf("Service now running..")
	handlers.Run(&configHandles, &env)

}

func main() {
	configServices()

}
