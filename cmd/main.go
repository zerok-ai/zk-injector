package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"

	"zerok-injector/internal/config"
	"zerok-injector/pkg/cert"
	"zerok-injector/pkg/server"
	"zerok-injector/pkg/storage"
	"zerok-injector/pkg/utils"

	"github.com/kataras/iris/v12"
)

//TODO:
// Integrate iris into the repo.
// Get new architecture flow from Mudit and implement polling from redis.
// Error handling in redis integration.
// Check todos in image map handler file.
// Add a script for stage and prod as well.
// What if the processes change in a pod based on args to a container?
// Add service name from workload to otel agent.
// Integrate a logger in the project.

func main() {

	var cfg config.ZkInjectorConfig
	args := ProcessArgs(&cfg)

	// read configuration from the file and environment variables
	log.Println("args.ConfigPath==", args.ConfigPath)

	if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
		log.Println(err)
	}

	runtimeMap := &storage.ImageRuntimeHandler{ImageRuntimeMap: &sync.Map{}}
	runtimeMap.Init(cfg.Redis)

	// start data collector
	//go server.StartServer(runtimeMap)

	app := newApp()

	config := iris.WithConfiguration(iris.Configuration{
		DisablePathCorrection: true,
		LogLevel:              "debug",
	})

	if cfg.Debug {
		server.StartDebugWebHookServer(app, cfg, runtimeMap, config)
	} else {
		// initialize certificates
		caPEM, cert, key, err := cert.InitializeKeysAndCertificates(cfg.Webhook)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		// start mutating webhook
		err = utils.CreateOrUpdateMutatingWebhookConfiguration(caPEM, cfg.Webhook)
		if err != nil {
			msg := fmt.Sprintf("Failed to create or update the mutating webhook configuration: %v\n", err)
			fmt.Println(msg)
			panic(msg)
		}

		// start webhook server
		server.StartWebHookServer(app, cfg, cert, key, runtimeMap, config)
	}
}

func newApp() *iris.Application {
	app := iris.Default()

	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods",
				"POST, PUT, PATCH, DELETE")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}
	app.UseRouter(crs)
	app.AllowMethods(iris.MethodOptions)

	return app
}

// Args command-line parameters
type Args struct {
	ConfigPath string
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs(cfg interface{}) Args {
	var a Args

	f := flag.NewFlagSet("Example server", 1)
	f.StringVar(&a.ConfigPath, "c", "config.yaml", "Path to configuration file")

	fu := f.Usage
	f.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), envHelp)
	}

	f.Parse(os.Args[1:])
	return a
}
