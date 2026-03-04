# Cobra

```sh
go get github.com/spf13/cobra
```

> CLI framework. Used by kubectl, Hugo, GitHub CLI, and most major Go CLIs.

## Basic structure

```go
// main.go
func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}

// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A brief description",
    Long:  "A longer description of your app.",
}

func Execute() error {
    return rootCmd.Execute()
}
```

## Subcommands

```go
// cmd/serve.go
var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    RunE: func(cmd *cobra.Command, args []string) error {
        // RunE returns an error — preferred over Run
        return startServer()
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)
}
```

## Flags

```go
// Persistent flag (available to command and all subcommands)
rootCmd.PersistentFlags().String("config", "", "config file path")

// Local flag (only for this command)
serveCmd.Flags().IntP("port", "p", 8080, "port to listen on")
serveCmd.Flags().BoolP("verbose", "v", false, "enable verbose output")

// Required flag
serveCmd.MarkFlagRequired("port")

// Read flag value inside RunE
port, _ := cmd.Flags().GetInt("port")
```

## Bind flags to viper (common pattern)

```go
func init() {
    serveCmd.Flags().Int("port", 8080, "port to listen on")
    viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
}

// Now both --port flag and SERVER_PORT env var and config file work
```

## Args validation

```go
var getCmd = &cobra.Command{
    Use:   "get [id]",
    Short: "Get an item by ID",
    Args:  cobra.ExactArgs(1),  // must have exactly 1 arg
    RunE: func(cmd *cobra.Command, args []string) error {
        id := args[0]
        return getItem(id)
    },
}

// Other validators
cobra.NoArgs          // no args allowed
cobra.MinimumNArgs(1) // at least 1
cobra.MaximumNArgs(3) // at most 3
cobra.RangeArgs(1, 3) // between 1 and 3
```

## Typical project layout

```
myapp/
├── main.go
└── cmd/
    ├── root.go     // rootCmd, Execute(), persistent flags
    ├── serve.go    // serve subcommand
    ├── migrate.go  // migrate subcommand
    └── version.go  // version subcommand
```

## Pre/Post hooks

```go
var serveCmd = &cobra.Command{
    Use: "serve",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return initConfig() // runs before command and subcommands
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        return startServer()
    },
    PostRunE: func(cmd *cobra.Command, args []string) error {
        return cleanup()
    },
}
```
