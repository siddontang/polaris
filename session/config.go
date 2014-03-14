package session

type RedisConfig struct {
	//redis data source name(dsn):
	//format: <password>@<host>:<port>/<db>
	DSN string

	//redis max idle connection
	MaxIdle int
}

type Config struct {
	//session default timeout, 0 for no expire
	Timeout int

	//session encode and decode serializer
	Serializer Codec

	//below for redis session
	Redis RedisConfig
}
