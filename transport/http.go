package urlfettr

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/alecthomas/template"
	"github.com/cloudnoize/urlFeatureExctrctor/service"
)

func GetTemplateHandler() http.HandlerFunc {
	wd, err := os.Getwd()

	if err != nil {
		log.Panic(err)
	}

	tmpl := template.Must(template.ParseFiles(wd + "/transport/tmpl/features.html"))

	f := func(w http.ResponseWriter, urlf *urlfeatures.UrlFeatures) {
		tmpl.Execute(w, urlf)
	}

	return GetUrlExtractorHandler(f)

}

func GetJsonHandler() http.HandlerFunc {
	f := func(w http.ResponseWriter, urlf *urlfeatures.UrlFeatures) {
		b, err := json.Marshal(urlf)
		if err != nil {
			http.Error(w, "No url param", http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, string(b))
	}

	return GetUrlExtractorHandler(f)
}

func GetUrlExtractorHandler(f func(w http.ResponseWriter, urlf *urlfeatures.UrlFeatures)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		val, ok := r.URL.Query()["url"]
		if !ok {
			http.Error(w, "No url param", http.StatusBadRequest)
			return
		}
		if len(val) == 0 {
			http.Error(w, "No url to analyze", http.StatusBadRequest)
			return
		}

		if len(val) > 1 {
			http.Error(w, "too many urls", http.StatusBadRequest)
			return
		}

		url, err := url.Parse(val[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		urlf := urlfeatures.Extract(val[0], url)
		f(w, urlf)

	}

}
