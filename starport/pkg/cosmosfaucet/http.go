package cosmosfaucet

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ignite-hq/cli/starport/pkg/openapiconsole"
	"github.com/rs/cors"
)

// ServeHTTP implements http.Handler to expose the functionality of Faucet.Transfer() via HTTP.
// request/response payloads are compatible with the previous implementation at allinbits/cosmos-faucet.
func (f Faucet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.Handle("/", cors.Default().Handler(http.HandlerFunc(f.faucetHandler))).
		Methods(http.MethodPost)

	router.Handle("/info", cors.Default().Handler(http.HandlerFunc(f.faucetInfoHandler))).
		Methods(http.MethodGet)

	router.HandleFunc("/", openapiconsole.Handler("Faucet", "openapi.yml")).
		Methods(http.MethodGet)

	router.HandleFunc("/openapi.yml", f.openAPISpecHandler).
		Methods(http.MethodGet)

	router.ServeHTTP(w, r)
}
