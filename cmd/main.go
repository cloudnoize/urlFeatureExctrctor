package main

import (
	"net/http"

	"github.com/cloudnoize/urlFeatureExctrctor/transport"
)

func main() {
	addr := ":8989"
	http.DefaultServeMux.Handle("/", urlfettr.GetTemplateHandler())
	http.DefaultServeMux.Handle("/json", urlfettr.GetJsonHandler())
	http.ListenAndServe(addr, nil)
}
