package main

import (
	"encoding/json"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	MentionType  = "mentions"
	EmoticonType = "emoticons"
	UrlType      = "urls"
)

//
// Struct to exchange data from the sub-processors to main processor of
// the message
type channelData struct {
	DataType string
	Value    interface{}
}

//
// Struct that details the url and the associated title
type urlInfo struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

//
// Struct that abstracts the response
type apiResponse struct {
	StatusCode int
	Message    string
	Data       interface{}
}

//
// Main entry point
func main() {
	http.HandleFunc("/v1/getInfo", handleMsg)
	http.ListenAndServe(":8000", nil)
}

//
// Method to write the response to the http writer
func (r *apiResponse) ToWriter(w http.ResponseWriter) {

	w.WriteHeader(r.StatusCode)
	var respData interface{}
	if r.StatusCode == http.StatusOK {
		respData = r.Data
	} else if r.StatusCode >= 400 && r.StatusCode < 500 {
		respData = r.Message
	}

	// Serialize the data to JSON
	respBytes, err := json.Marshal(respData)
	if err != nil {
		respBytes = []byte("{\"code\": 500, \"message\": \"Failed to marshal response\"}")
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	}
	w.Write(respBytes)
}

//
// Function for handling requests
func handleMsg(w http.ResponseWriter, r *http.Request) {

	//
	// Only processing POST methods
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	//
	// Channel on which the completion of request processing will be indicated
	rc := make(chan *apiResponse, 1)

	//
	// Calling the message processor on a seperate go routine.
	go func() {
		resp := processMsg(w, r)
		rc <- resp
	}()

	//
	// Waiting for the message processor to finish.  If it does not
	// complete in the stipulated time(5 secs) then we return a timeout
	// response to the user
	select {
	case resp := <-rc:
		resp.ToWriter(w)
	case <-time.After(time.Second * 5):
		// return a time out response
		http.Error(w, "Timed out", http.StatusRequestTimeout)
	}
}

//
// Function to parse the http request and build http response
func processMsg(w http.ResponseWriter, r *http.Request) *apiResponse {

	//
	// Get the message from the request
	msg := r.FormValue("message")
	if len(msg) == 0 {
		resp := apiResponse{StatusCode: http.StatusBadRequest, Message: "Empty Body in the Request", Data: nil}
		return &resp
		//http.Error(w, "Failed to read the body of the request", http.StatusBadRequest)
	}

	//
	// Build the response struct
	respData, err := processMsgHelper(&msg)
	if err != nil {
		resp := apiResponse{StatusCode: http.StatusInternalServerError, Message: "Server Error", Data: nil}
		return &resp
	}

	resp := apiResponse{StatusCode: http.StatusOK, Message: "", Data: respData}
	return &resp
}

//
// Function that does the message processing
func processMsgHelper(msg *string) (map[string][]interface{}, error) {

	respData := make(map[string][]interface{})
	ch := make(chan channelData)

	//
	// Aggregate method collects all the results and builds a map of results
	var wgAgg sync.WaitGroup
	wgAgg.Add(1)
	go func(c <-chan channelData, wg *sync.WaitGroup, res *map[string][]interface{}) {
		aggregateResult(c, wg, res)
	}(ch, &wgAgg, &respData)

	//
	// Process the message by calling into each rule - @mention, emoticons,
	// and url.  Each of these are on a seperate go-routine. 'wg' will
	// facilitate in waiting until all the go routines have completed
	var wgProc sync.WaitGroup
	wgProc.Add(3)
	go func(msg *string, wg *sync.WaitGroup, ch chan channelData) {
		processMentions(msg, wg, ch)
	}(msg, &wgProc, ch)
	go func(msg *string, wg *sync.WaitGroup, ch chan channelData) {
		processEmoticons(msg, wg, ch)
	}(msg, &wgProc, ch)
	go func(msg *string, wg *sync.WaitGroup, ch chan channelData) {
		processUrls(msg, wg, ch)
	}(msg, &wgProc, ch)

	wgProc.Wait()

	//
	// Close will help the aggregate method to stop processing the channel
	close(ch)

	//
	// Wait until the aggregate finishes reading all the channel elements
	wgAgg.Wait()

	return respData, nil
}

//
// Function that collects the results from channel and builds the result map
func aggregateResult(c <-chan channelData, wg *sync.WaitGroup, res *map[string][]interface{}) {
	defer wg.Done()

	//
	// Keep reading from the channel and building the map
	for r := range c {
		if _, ok := (*res)[r.DataType]; ok == false {
			(*res)[r.DataType] = make([]interface{}, 0)
		}
		(*res)[r.DataType] = append((*res)[r.DataType], r.Value)
	}
}

//
// Function for checking if mentions are in the message
func processMentions(msg *string, wg *sync.WaitGroup, c chan channelData) {
	defer wg.Done()

	//
	// RegEx for finding at mentions.  mentions should start with '@' and
	// have to be at the start of a word or message
	r, _ := regexp.Compile(`(^|\s)@\w+`)
	res := r.FindAllString(*msg, -1)

	//
	// Process each string output emitted by regex
	// We will spawn a new goroutine for each string output so as to
	// parallelize the processing. 'wgLocal' will be a wait group
	// that will enable waiting until all the go routines have completed
	var wgLocal sync.WaitGroup
	wgLocal.Add(len(res))
	for _, s := range res {
		go func(c chan channelData, v string, wgL *sync.WaitGroup) {
			defer wgL.Done()

			var subV string
			if v[0] == '@' {
				// This takes care of beginning of a message
				subV = strings.TrimSpace(v[1:])
			} else {
				// This takes care of beginning of a word
				subV = strings.TrimSpace(v[2:])
			}
			cD := channelData{DataType: MentionType, Value: &subV}
			c <- cD
		}(c, s, &wgLocal)
	}

	//
	// Wait until all the local go routines have completed
	wgLocal.Wait()
}

//
// Function for checking if emoticons are in the message
func processEmoticons(msg *string, wg *sync.WaitGroup, c chan channelData) {
	defer wg.Done()

	//
	// RegEx for finding emoticons
	r, _ := regexp.Compile(`\(\w+\)`)
	res := r.FindAllString(*msg, -1)

	//
	// Process each string output emitted by regex
	// We will spawn a new goroutine for each string output so as to
	// parallelize the processing. 'wgLocal' will be a wait group
	// that will enable waiting until all the go routines have completed
	var wgLocal sync.WaitGroup
	wgLocal.Add(len(res))
	for _, s := range res {
		go func(c chan channelData, e string, wgL *sync.WaitGroup) {
			defer wgL.Done()
			if len(e) != 15 {
				return
			}
			cD := channelData{DataType: EmoticonType, Value: &e}
			c <- cD
		}(c, s[1:len(s)-1], &wgLocal)
	}

	wgLocal.Wait()
}

//
// Function for parsing the title from the html page
func parseTitle(n *html.Node) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "title" {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := parseTitle(c)
		if ok {
			return result, ok
		}
	}
	return "", false
}

//
// Function for processing URLs from the message
func processUrls(msg *string, wg *sync.WaitGroup, c chan channelData) {
	defer wg.Done()

	r, _ := regexp.Compile(`http[s]*://(\w|\d)+(\w|-|\d|/|\.)+((\?|\&)+(.)*)*(\s|\z)`)
	res := r.FindAllString(*msg, -1)

	var wgLocal sync.WaitGroup
	wgLocal.Add(len(res))
	for _, s := range res {
		go func(c chan channelData, u string, wgL *sync.WaitGroup) {
			defer wgL.Done()

			// Get the page from the url
			timeout := time.Duration(3 * time.Second)
			client := http.Client{Timeout: timeout}
			resp, err := client.Get(u)
			if err != nil {
				return
			}

			// Ensures closing the body at the end of this method
			defer resp.Body.Close()

			// Parse the page
			d, parseErr := html.Parse(resp.Body)
			if parseErr != nil {
				return
			}

			// Parse the title of the page
			t, _ := parseTitle(d)

			ui := urlInfo{Url: u, Title: t}
			cD := channelData{DataType: UrlType, Value: &ui}
			c <- cD
		}(c, strings.TrimSpace(s), &wgLocal)
	}
	wgLocal.Wait()

}
