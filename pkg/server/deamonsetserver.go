package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"zerok-injector/pkg/common"
	"zerok-injector/pkg/storage"
)

var syncRunTimePath = "/sync-runtime"

type SyncRuntimeApiHandler struct {
	ImageRuntimeHandler *storage.ImageRuntimeHandler
}

func StartServer(runtimeMap *storage.ImageRuntimeHandler) {
	fmt.Println("Starting server.")
	mux := http.NewServeMux()
	syncRuntimeHandler := SyncRuntimeApiHandler{
		ImageRuntimeHandler: runtimeMap,
	}
	mux.Handle(syncRunTimePath, &syncRuntimeHandler)
	s := &http.Server{
		Addr:    ":8444",
		Handler: mux,
	}
	s.ListenAndServe()
}

func (h *SyncRuntimeApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s call received at %s\n", r.Method, syncRunTimePath)

	if r.Method == http.MethodGet {
		fmt.Fprintln(w, h.getImageData())
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		errorResponse(err, w)
		return
	}

	err = h.setDataForImages(body)

	if err != nil {
		fmt.Printf("Error while injecting zk agent %v\n", err)
		w.WriteHeader(500)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Done"))
	}

	r.Body.Close()
}

func (h *SyncRuntimeApiHandler) getImageData() interface{} {

	//TODO Send data for images
	return nil
}

func (h *SyncRuntimeApiHandler) setDataForImages(body []byte) error {
	var result common.RuntimeSyncRequest
	err := json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	for _, detail := range result.RuntimeDetails {
		h.ImageRuntimeHandler.SaveRuntimeForImage(detail.Image, &detail)
	}
	return err
}

func errorResponse(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}
