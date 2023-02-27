package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/bolt"
	"github.com/quantonganh/blog/gmail"
	"github.com/quantonganh/blog/http"
	"github.com/quantonganh/blog/markdown"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.SetDefault("http.addr", ":80")
	viper.SetDefault("posts.dir", "posts")

	var cfg *blog.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	posts, err := markdown.GetAllPosts(cfg.Posts.Dir)
	if err != nil {
		log.Fatal(err)
	}

	a := newApp(cfg, posts)

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
	db         *bolt.DB
	httpServer *http.Server
}

func newApp(config *blog.Config, posts []*blog.Post) *app {
	indexPath := path.Join(path.Dir(config.Posts.Dir), path.Base(config.Posts.Dir)+".bleve")
	httpServer, err := http.NewServer(config, posts, indexPath)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	return &app{
		config:     config,
		db:         bolt.NewDB(config.DB.Path),
		httpServer: httpServer,
	}
}

func (a *app) Run(ctx context.Context) error {
	if err := a.db.Open(); err != nil {
		return err
	}

	a.httpServer.Addr = a.config.HTTP.Addr
	baseURL, err := url.Parse(a.config.Site.BaseURL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse URL %s", a.config.Site.BaseURL)
	}
	a.httpServer.Domain = baseURL.Hostname()

	if err := a.httpServer.Open(); err != nil {
		return err
	}

	a.httpServer.SubscribeService = bolt.NewSubscribeService(a.db)
	a.httpServer.SMTPService = gmail.NewSMTPService(a.config, a.httpServer.URL(), a.httpServer.SubscribeService, a.httpServer.Renderer)

	latestPosts := a.httpServer.PostService.GetLatestPosts(a.config.Newsletter.Frequency)
	if len(latestPosts) > 0 {
		a.httpServer.SMTPService.SendNewsletter(latestPosts)
	}

	return nil
}

func (a *app) Close() error {
	if a.httpServer != nil {
		if a.httpServer.SearchService != nil {
			if err := a.httpServer.SearchService.CloseIndex(); err != nil {
				return err
			}
		}

		if a.httpServer.SMTPService != nil {
			if err := a.httpServer.SMTPService.Stop(); err != nil {
				return err
			}
		}

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
