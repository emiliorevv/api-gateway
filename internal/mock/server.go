package mock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func Run() string {

	handler := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, `{"message: backend mock running"`)

	}

	ts := httptest.NewServer(http.HandlerFunc(handler))

	return ts.URL
}
