package genie

import (
	"net/http"
	"reflect"
)

type reflectHandler struct {
	provider Provider
	index    int
}

func (h reflectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, err := h.provider()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	reflect.ValueOf(ctx).Method(h.index).Convert(tHandlerFunc).
		Interface().(http.HandlerFunc)(w, r)
}
