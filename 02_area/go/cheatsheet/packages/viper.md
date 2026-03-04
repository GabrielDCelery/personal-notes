# Viper

```sh
go get github.com/spf13/viper
```

> Config management — reads from files, env vars, flags, and remote stores. Commonly paired with Cobra.

## Basic setup

```go
import "github.com/spf13/viper"

viper.SetConfigName("config")    // config.yaml, config.json, etc.
viper.SetConfigType("yaml")
viper.AddConfigPath(".")
viper.AddConfigPath("/etc/myapp")

if err := viper.ReadInConfig(); err != nil {
    log.Fatal(err)
}
```

## Read values

```go
viper.GetString("database.host")
viper.GetInt("database.port")
viper.GetBool("feature.enabled")
viper.GetDuration("server.timeout")
viper.GetStringSlice("allowed.origins")
```

## Environment variables

```go
viper.AutomaticEnv() // reads env vars automatically

// Map env var to config key
viper.SetEnvPrefix("APP")          // APP_DATABASE_HOST → database.host
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

// Explicit bind
viper.BindEnv("database.host", "DB_HOST")
```

## Defaults

```go
viper.SetDefault("server.port", 8080)
viper.SetDefault("server.timeout", 30*time.Second)
viper.SetDefault("log.level", "info")
```

## Unmarshal into struct (recommended)

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
    Port    int           `mapstructure:"port"`
    Timeout time.Duration `mapstructure:"timeout"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Name     string `mapstructure:"name"`
    Password string `mapstructure:"password"`
}

var cfg Config
if err := viper.Unmarshal(&cfg); err != nil {
    log.Fatal(err)
}
```

## Example config.yaml

```yaml
server:
  port: 8080
  timeout: 30s

database:
  host: localhost
  port: 5432
  name: myapp
  password: secret

log:
  level: info
```

## Watch for config changes

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    fmt.Println("config changed:", e.Name)
    // reload config
})
```
