package blog

type Config struct {
	Navbar struct {
		Items []*Item
	}

	Posts struct {
		Dir string
	}

	Templates struct {
		Dir string
	}

	DB struct {
		Path string
	}

	HTTP struct {
		Addr   string
		Domain string
	}

	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
	}

	Newsletter struct {
		Frequency int
		Cron      struct {
			Spec string
		}
		Product struct {
			Name string
		}
		HMAC struct {
			Secret string
		}
	}

	Sentry struct {
		DSN string
	}
}

type Item struct {
	Text string
	URL  string
}
