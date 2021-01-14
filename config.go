package blog

type Config struct {
	Posts struct {
		Dir string
	}

	Templates struct {
		Dir string
	}

	DB struct {
		DSN string
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
}
