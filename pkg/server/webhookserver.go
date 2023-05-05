package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"zerok-injector/internal/config"
	"zerok-injector/pkg/inject"
	"zerok-injector/pkg/storage"

	"github.com/kataras/iris/v12"
)

type HttpApiHandler struct {
	injector *inject.Injector
}

func (h *HttpApiHandler) ServeHTTP(ctx iris.Context) {
	body, err := io.ReadAll(ctx.Request().Body)

	fmt.Printf("Got a request from webhook")

	if err != nil {
		webhookErrorResponse(err, ctx)
		return
	}

	response, err := h.injector.Inject(body)

	if err != nil {
		fmt.Printf("Error while injecting zk agent %v\n", err)
	}

	// Sending http status as OK, even when injection failed to not disturb the pods in cluster.
	ctx.StatusCode(iris.StatusOK)
	ctx.Write(response)
}

func webhookErrorResponse(err error, ctx iris.Context) {
	log.Println(err)
	ctx.StatusCode(iris.StatusInternalServerError)
}

func handleRoutes(app *iris.Application, cfg config.ZkInjectorConfig, runtimeMap *storage.ImageRuntimeHandler) {
	injectHandler := &HttpApiHandler{
		injector: &inject.Injector{ImageRuntimeHandler: runtimeMap},
	}
	app.Post(cfg.Webhook.Path, injectHandler.ServeHTTP)
}

func StartWebHookServer(app *iris.Application, cfg config.ZkInjectorConfig, cert *bytes.Buffer, key *bytes.Buffer, runtimeMap *storage.ImageRuntimeHandler) {
	handleRoutes(app, cfg, runtimeMap)
	app.Run(iris.TLS(":"+cfg.Webhook.Port, cert.String(), key.String()))
}

func StartDebugWebHookServer(app *iris.Application, cfg config.ZkInjectorConfig, runtimeMap *storage.ImageRuntimeHandler) {
	handleRoutes(app, cfg, runtimeMap)
	app.Run(iris.Addr(":" + cfg.Webhook.Port))
}
