package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http"
	"github.com/quantonganh/blog/kafka"
	"github.com/quantonganh/blog/markdown"
	"github.com/quantonganh/blog/rabbitmq"
	"github.com/quantonganh/blog/sqlite"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.SetDefault("http.addr", ":8009")
	viper.SetDefault("posts.dir", "posts")

	var config *blog.Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.Sentry.DSN,
	}); err != nil {
		log.Fatalf("sentry.Init: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	posts, err := markdown.GetAllPosts(config.Posts.Dir)
	if err != nil {
		log.Fatal(err)
	}

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	a, err := newApp(logger, config, posts)
	if err != nil {
		logger.Error().Err(err).Msg("error creating new app")
	}

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
	db         *sqlite.DB
	config     *blog.Config
	httpServer *http.Server
}

func newApp(logger zerolog.Logger, config *blog.Config, posts []*blog.Post) (*app, error) {
	httpServer, err := http.NewServer(logger, config, posts)
	if err != nil {
		logger.Error().Err(err).Msg("error creating new HTTP server")
		return nil, err
	}

	queueService, err := rabbitmq.NewQueueService(config.AMQP.URL)
	if err != nil {
		return nil, err
	}
	httpServer.QueueService = queueService

	eventService, err := kafka.NewEventService(config.Kafka.Broker)
	if err != nil {
		return nil, err
	}
	httpServer.EventService = eventService

	db := sqlite.NewDB("db/stats.db")
	statService := sqlite.NewStatService(logger, db)
	httpServer.StatService = statService

	return &app{
		db:         db,
		config:     config,
		httpServer: httpServer,
	}, nil
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

	if err := a.httpServer.ProcessActivityStream(ctx, a.config.IP2Location.Token); err != nil {
		return err
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

		if err := a.httpServer.Close(); err != nil {
			return err
		}
	}

	return nil
}
