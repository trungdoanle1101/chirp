package main

import (
	"fmt"
	"net/http"
)

const metricsHtml = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	status := http.StatusOK
	content := fmt.Sprintf(metricsHtml, cfg.fileserverHits.Load())
	w.WriteHeader(status)
	w.Write([]byte(content))

}
