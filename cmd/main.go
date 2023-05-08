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

// TODO:
// Don't use same redis client across threads in golang? -- why?
// Implement restart of workload.
// Add zklogger in the project.
// Add comments wherever required.
// Break down methods for testing convenience.
// Merge injector with operator.
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

	app := newApp()

	config := iris.WithConfiguration(iris.Configuration{
		DisablePathCorrection: true,
		LogLevel:              "debug",
	})

	if cfg.Local {
		server.StartDebugWebHookServer(app, cfg, runtimeMap, config)
	} else {
		// initialize certificates
		caPEM, cert, key, err := cert.InitializeKeysAndCertificates(cfg.Webhook)
		if err != nil {
			msg := fmt.Sprintf("Failed to create keys and certificates for webhook %v. Stopping initialization of the pod.\n", err)
			fmt.Println(msg)
			return
		}

		// start mutating webhook
		err = utils.CreateOrUpdateMutatingWebhookConfiguration(caPEM, cfg.Webhook)
		if err != nil {
			msg := fmt.Sprintf("Failed to create or update the mutating webhook configuration: %v. Stopping initialization of the pod.\n", err)
			fmt.Println(msg)
			return
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
			ctx.Header("Access-Control-Methods", "POST")

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
