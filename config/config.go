package config

type Config struct {
	Navbar     navbar
	SMTP       smtp
	Newsletter newsletter
}

type navbar struct {
	Items []item
}

type item struct {
	Text string
	URL  string
}

type smtp struct {
	Host     string
	Port     int
	Username string
	Password string
}

type newsletter struct {
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
