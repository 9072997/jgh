package jgh

import "reflect"
import "time"
import "net/http"
import "net/http/cookiejar"
import "log"
import "io"
import "strconv"
import "io/ioutil"
import "strings"
import "encoding/json"
import "fmt"
import "math/rand"
import "runtime/debug"

// TODO better error checking on these next 3 functions, but right now, I just panic
// on basically every error, so...

// sets pointer to something

func DerefrenceInterface(ptrIface interface{}) (iface interface{}, err error) {
	// derefrence outputPtr (using reflection since it's an interface)
	// turn the interface into a value (of the pointer)
	ptrVal := reflect.ValueOf(ptrIface)
	// derefrence the pointer
	val := reflect.Indirect(ptrVal)
	// turn that back into an interface
	iface = val.Interface()
	return
}

// given an example of type exIface, returns a pointer to a zero value for exIface
func PtrToZeroOf(exIface interface{}) (zeroPtrIface interface{}) {
	// get the type of example
	exType := reflect.TypeOf(exIface)

	// get the address of a newly minted zero
	zeroPtrVal := reflect.New(exType)

	// turn our pointer into an interface
	zeroPtrIface = zeroPtrVal.Interface()

	return
}

// nolint: megacheck, deadcode
func InitSlice(s interface{}, len int) (err error) {
	// turn the interface into a value (of the pointer)
	vInt := reflect.ValueOf(s)

	// derefrence the pointer
	v := reflect.Indirect(vInt)

	// get the type the pointer points to
	t := v.Type()

	// make initialized slice
	is := reflect.MakeSlice(t, len, len)

	// save the initialized slice through the pointer
	// use reflection to bypass go's type checking
	v.Set(is)

	return
}

/*
func setThroughPointer(pointerIface interface{}, somethingIface interface{}) (err error) {
	// turn the interface into a value (of the pointer)
	pointerVal := reflect.ValueOf(pointerIface)

	// derefrence the pointer to it's destination (still a value)
	destVal := reflect.Indirect(pointerVal)

	// turn "something" into a value so we can do reflection stuff on it
	somethingVal := reflect.ValueOf(somethingIface)

	// save to the pointer's destination
	destVal.Set(somethingVal)
}

func zeroOf(exampleIface interface{}) (zeroIface interface{}) {
	exampleType := reflect.TypeOf(exampleIface)
	zeroVal := reflect.Zero(exampleType)
	zeroIface = zeroVal.Interface()
	return
}


func zeroOfElem(ptrIface interface{}) (zeroIface interface{}) {
	destType := reflect.TypeOf(ptrIface).Elem()
	zeroVal := reflect.Zero(exampleType)
	zeroIface = zeroVal.Interface()
	return
}

func initPtr(ptrPtrIface interface{}) {
	// get the naked type
	t := reflect.TypeOf(ptrPtrIface).Elem().Elem()

	// get the address of a newly minted zero
	zeroPtrVal := reflect.New(t)

	// turn the pointer-pointer into a value
	ptrPtrVal := reflect.ValueOf(ptrPtrIface)

	// derefrence
	PtrVal := reflect.Indirect(ptrPtrVal)

	PtrVal.Set(zeroPtrVal)

	return
}

// s should be of type *[]interface{}

*/

// retryes f()(bool) at i second intervals up to t times until f() == true
// note that this function will also retry on panic
// prints "msg (will retry up to t times)" for each try
func Try(interval int, tries int, allowPanic bool, msg string, f func() bool) bool { // nolint: deadcode, megacheck
	// this makes sure we don't panic if f() does
	defer func() {
		// if we are on our last iteration, let the panic continue to bubble up
		if tries > 1 || !allowPanic {
			err := recover()
			if err != nil {
				log.Printf("Panic while %s: %v", msg, err)
				debug.PrintStack()
			}
		}
	}()

	// if tries is negitive, we retry forever
	infinite := tries < 0

	for ; tries > 0 || infinite; tries-- {
		if tries < 0 {
			log.Printf("%s (try %d)", msg, -tries)
		} else {
			log.Printf("%s (will retry up to %d times)", msg, tries)
		}

		success := f()
		if success {
			// f() was successful
			return true
		}

		// no point in sleeping if we are not going to retry f()
		if tries > 1 || infinite {
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
	// we have run f() t times without success
	return false
}

func HTTPClient(cookieJar bool, followRedirects bool) (client *http.Client) {
	log.Printf("Making new http client cookieJar:%t, followRedirects:%t", cookieJar, followRedirects)

	client = new(http.Client)

	if !followRedirects {
		// really basic redirect handeling function that dosen't follow redirects
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	if cookieJar {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic("Failed to create cookie jar")
		}
		client.Jar = jar
	}

	return
}

func HTTPRequest(client *http.Client, method string, url string, user string, pass string, headers map[string]string, reqBody string) (respBody string, status int) {
	log.Printf("HTTP %s %s", method, url)

	// empty string indicates no request body
	hasBody := len(reqBody) > 0
	if hasBody {
		log.Println("Request Body: ", reqBody)
	}

	// turn the request body into an io.Reader
	var reqBodyReader io.Reader
	if hasBody {
		reqBodyReader = strings.NewReader(reqBody)
	} else {
		reqBodyReader = nil
	}

	// create a new request object
	req, err := http.NewRequest(method, url, reqBodyReader)
	if err != nil {
		panic("Failed to create request object")
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	// add useragent (if one wasn't specified)
	if _, keyExists := headers["User-Agent"]; !keyExists {
		req.Header.Add("User-Agent", "qbox-jgh/1.0")
	}

	// add request headers
	if hasBody {
		// even if the user set a content length, replace it with ours
		headers["Content-Length"] = strconv.Itoa(len(reqBody))
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// add auth if present
	if len(user) > 0 || len(pass) > 0 {
		req.SetBasicAuth(user, pass)
	}

	// perform the http request
	resp, err := client.Do(req)
	if err != nil {
		panic("Error while performing http request")
	}

	// get status code
	status = resp.StatusCode

	// get response body into a string
	defer resp.Body.Close() // nolint: errcheck
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read response body from http request")
	}
	respBody = string(bytes)
	log.Println("Response Body: ", respBody)

	return
}

func RESTRequest(client *http.Client, method string, url string, user string, pass string, headers map[string]string, input interface{}, outputPtr interface{}) (status int, reflection bool) {
	hasInput := input != nil
	hasOutput := outputPtr != nil

	var jsonStr string
	if hasInput {
		// convert input strict to json string
		bytes, err := json.Marshal(input)
		if err != nil {
			panic("Failed to marshal json")
		}
		jsonStr = string(bytes)
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	// defaults for content-type and accept
	if _, keyExists := headers["Content-Type"]; hasInput && !keyExists {
		headers["Content-Type"] = "application/json"
	}
	if _, keyExists := headers["Accept"]; (hasInput || hasOutput) && !keyExists {
		headers["Accept"] = "application/json"
	}

	// perform the request
	respStr, status := HTTPRequest(client, method, url, user, pass, headers, jsonStr)

	// even if the user dosen't want output, we still need a place to store
	// it so we can check for reflection
	if hasInput && !hasOutput {
		outputPtr = PtrToZeroOf(input)
	}

	if hasInput || hasOutput {
		bytes := []byte(respStr)
		err := json.Unmarshal(bytes, outputPtr)
		PanicOnErr(err)
	}

	if hasInput {
		// many calls return the input as output on success, so we check for this here
		output, err := DerefrenceInterface(outputPtr)
		PanicOnErr(err)

		reflection = reflect.DeepEqual(input, output)
	}

	return
}

func Expect(expected interface{}, input interface{}, name string) {
	if !reflect.DeepEqual(input, expected) {
		msg := fmt.Sprintf("Expected %s to be %v, got %v", name, expected, input)
		panic(msg)
	}
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	// we need this for the random string
	// TODO this dosen't really give us that much entropy
	rand.Seed(time.Now().UTC().UnixNano())
}

func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
