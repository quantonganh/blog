package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/viper"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/gmail"
	"github.com/quantonganh/blog/http"
	"github.com/quantonganh/blog/mongo"
	"github.com/quantonganh/blog/ondisk"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.SetDefault("posts.dir", "posts")
	viper.SetDefault("http.domain", "http://localhost")
	viper.SetDefault("db.dsn", "mongodb://localhost:27017")

	var cfg *blog.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	posts, err := ondisk.GetAllPosts(cfg.Posts.Dir)
	if err != nil {
		log.Fatal(err)
	}

	a := NewApp(cfg, posts)

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	if err := a.Run(ctx); err != nil {
		_ = a.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := a.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type app struct {
	config     *blog.Config
	db         *mongo.DB
	httpServer *http.Server
}

func NewApp(config *blog.Config, posts []*blog.Post) *app {
	return &app{
		config:     config,
		db:         mongo.NewDB(config.DB.DSN),
		httpServer: http.NewServer(config, posts),
	}
}

func (a *app) Run(ctx context.Context) error {
	if err := a.db.Open(); err != nil {
		return err
	}

	subscribeService := mongo.NewSubscribeService(a.db)
	a.httpServer.SubscribeService = subscribeService
	a.httpServer.SMTPService = gmail.NewSMTPService(a.config, a.httpServer.Templates, subscribeService)

	latestPosts := a.httpServer.PostService.GetLatestPosts(a.config.Newsletter.Frequency)
	if len(latestPosts) > 0 {
		a.httpServer.SMTPService.SendNewsletter(latestPosts)
	}

	if err := a.httpServer.Open(); err != nil {
		return err
	}

	return nil
}

func (a *app) Close() error {
	if a.httpServer != nil {
		if err := a.httpServer.Close(); err != nil {
			return err
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return err
		}
	}

	return nil
}
