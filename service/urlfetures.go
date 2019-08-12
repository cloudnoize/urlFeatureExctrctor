package urlfeatures

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/miekg/dns"
	"golang.org/x/net/html"
)

type UrlFeatures struct {
	Fullurl     string
	GoogleScore float64
	NumLinks    int
	IsHttps     bool
	Hostname    string
	Length      int
	IPv4        []*IPFeatures
	MXIPs       []*MxFeatures
}

// func (ul UrlFeatures) String() string {
// 	ret := "\nGoogle knows [" + strconv.FormatFloat(ul.GoogleScore) + "]\nIs https [" + strconv.FormatBool(ul.IsHttps) + "]\nHostname [" + ul.Hostname + "]\nLength [" + strconv.Itoa(ul.Length) + "]\n"
// 	ret += "IPv4:\n"

// 	for _, v := range ul.IPv4 {
// 		ret += v.String() + "\n"
// 	}

// 	ret += "MX:\n"
// 	for _, v := range ul.MXIPs {
// 		ret += v.String() + "\n"
// 	}
// 	return ret
// }

type IPFeatures struct {
	IP          net.IP
	ttl         uint32
	Geolocation *Geolocation
}

type Geolocation struct {
	IP          string  `json:"IP"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	ZIPCode     string  `json:"zIP_code"`
	TimeZone    string  `json:"time_zone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
}

func (IPf *IPFeatures) String() string {
	return "IP [" + IPf.IP.String() + "]\nttl [" + strconv.Itoa(int(IPf.ttl)) + "]\nLocation [" + IPf.Geolocation.CountryName + "] Country code [" + IPf.Geolocation.CountryCode + "]\n"
}

func (IPf *IPFeatures) SetLocation() {
	query := "https://freegeoIP.app/json/" + IPf.IP.String()

	client := &http.Client{}

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	var b bytes.Buffer

	io.Copy(&b, resp.Body)

	defer resp.Body.Close()

	gl := new(Geolocation)
	err = json.Unmarshal(b.Bytes(), gl)

	if err != nil {
		log.Println(err)
		return
	}

	IPf.Geolocation = gl

}

type MxFeatures struct {
	IPFeatures
	name string
}

func (mpf MxFeatures) String() string {
	return "MX name: [" + mpf.name + "]\n" + mpf.IPFeatures.String()
}

func DnsFeatures() func(host string) ([]*IPFeatures, []*MxFeatures) {
	cl := new(dns.Client)
	dnsaddr := "8.8.8.8:53"
	return func(host string) ([]*IPFeatures, []*MxFeatures) {
		var (
			IPfs []*IPFeatures
			mxfs []*MxFeatures
		)
		ss := &dns.Msg{
			Question: make([]dns.Question, 1),
		}
		ss.SetQuestion(host+".", dns.TypeA)
		res, _, err := cl.Exchange(ss, dnsaddr)
		if err != nil {
			log.Println(err)
			return nil, nil
		}
		for _, record := range res.Answer {
			switch record.(type) {
			case *dns.A:
				rec := record.(*dns.A)
				IPf := &IPFeatures{IP: rec.A, ttl: rec.Hdr.Ttl}
				IPfs = append(IPfs, IPf)
			}
		}

		ss.SetQuestion(host+".", dns.TypeMX)
		res, _, err = cl.Exchange(ss, dnsaddr)
		if err != nil {
			log.Println(err)
			return nil, nil
		}
		for _, record := range res.Answer {
			switch record.(type) {
			case *dns.MX:
				rec := record.(*dns.MX)
				mxf := &MxFeatures{IPFeatures: IPFeatures{ttl: rec.Hdr.Ttl}, name: rec.Mx}
				mxfs = append(mxfs, mxf)
			}
		}
		return IPfs, mxfs
	}
}

func AddScheme(url string) string {
	if strings.HasPrefix(url, "http://") {
		return url
	}
	if strings.HasPrefix(url, "https://") {
		return url
	}
	return "http://" + url
}
func SliceContains(s []*UrlFeatures, val string) bool {
	for _, v := range s {
		if val == v.Hostname {
			return true
		}
	}
	return false
}

func GetGoogleUrls(url string, parsedurl *url.URL, numres int) []string {
	query := "http://www.google.com/search?hl=en&num=" + strconv.Itoa(numres) + "&q=" + url + "&as_sitesearch=" + parsedurl.Hostname()

	client := &http.Client{}

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")

	resp, err := client.Do(req)

	log.Println("Running ", query)

	if err != nil {
		log.Println(err)
		return nil
	}

	var b bytes.Buffer

	io.Copy(&b, resp.Body)

	node, err := html.Parse(&b)

	if err != nil {
		log.Println(err)
		return nil
	}

	doc := goquery.NewDocumentFromNode(node)
	var links []string

	doc.Find("#res").Each(func(index int, item *goquery.Selection) {
		item.Find("#search").Each(func(index int, item *goquery.Selection) {
			item.Find("a").Each(func(index int, item *goquery.Selection) {
				link, ok := item.Attr("href")
				if ok && len(link) > 7 {
					links = append(links, link)
				}
			})
		})
	})
	// doc.Find("#search").Each(func(index int, item *goquery.Selection) {
	// 	item.Find("a").Each(func(index int, item *goquery.Selection) {
	// 		title := item.Text()

	// 		fmt.Printf("search - %s\n", title)
	// 	})
	// })

	// doc.Find("#resultStats").Each(func(index int, item *goquery.Selection) {
	// 	title := item.Text()

	// 	fmt.Printf("resultStats - %s\n", title)
	// })

	// 	if z == nil {
	// 		log.Fatal("tokenization failed")
	// 	}

	// 	for {
	// 		tt := z.Next()
	// 		switch tt {
	// 		case html.ErrorToken:
	// 			fmt.Println("End")
	// 			goto endloop
	// 		case html.TextToken:
	// 			// te := z.Token()
	// 			// log.Println("text.Data ", te.Data)
	// 		case html.StartTagToken:
	// 			t := z.Token()
	// 			log.Println("t.Data ", t.Data)
	// 			if t.Data == "a" {
	// 				fmt.Println("We found an anchor!")
	// 				log.Println(t.String())
	// 			}
	// 			if t.Data == "div" {
	// 				fmt.Println("We found an div!")
	// 				for _, v := range t.Attr {
	// 					log.Println(v.Key, v.Val)
	// 				}
	// 				log.Println(t.String())
	// 			}
	// 		}
	// 	}
	// endloop:

	defer resp.Body.Close()

	return links
}

func Extract(surl string, url *url.URL) *UrlFeatures {
	links := GetGoogleUrls(surl, url, 100)
	score := len(links)
	normScore := float64(score) / float64(251) // Num of links when query for google
	if normScore > 1 {
		normScore = 1
	}
	urlf := &UrlFeatures{GoogleScore: normScore, Length: len(surl), Fullurl: surl, NumLinks: score}

	urlf.Hostname = url.Hostname()

	for _, v := range links {
		valurl, err := url.Parse(v)
		if err != nil {
			log.Panicln("Error ", err)
			continue
		}
		if strings.Contains(surl, valurl.Hostname()) {
			urlf.IsHttps = valurl.Scheme == "https"
			break
		}
	}

	urlf.IPv4, urlf.MXIPs = DnsFeatures()(urlf.Hostname)

	for _, v := range urlf.IPv4 {
		v.SetLocation()
	}

	return urlf

}
