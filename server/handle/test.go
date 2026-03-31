package handle

import (
	"net/http"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	// 成功 200
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world"))

}
