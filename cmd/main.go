package main

import (
	"net/http"

	"github.com/cloudnoize/urlFeatureExctrctor/transport"
)

func main() {
	addr := ":8989"
	http.DefaultServeMux.Handle("/", urlfettr.GetUrlExtractorHandler())
	http.ListenAndServe(addr, nil)
}
