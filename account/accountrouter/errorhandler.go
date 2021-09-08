package accountrouter

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/stevealexrs/Go-Libra/account"
)

type errorFunc func(http.ResponseWriter,*http.Request) error

func errorHandler(f errorFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		var printable *account.PrintableError
		if errors.As(err, &printable) {
			w.WriteHeader(http.StatusBadRequest)
			
			res, err := json.Marshal(printable.Message)
			if err != nil {
				log.Printf("json encoding failed: %s/n", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			w.Write(res)
		} else if err != nil {
			log.Printf("unknown error: %s/n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
