package jgh

import (
	"bufio"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"sync"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// I don't care enough to use the windows API in a threadsafe way
var oleMutex sync.Mutex

// WinHTTPRequest makes an HTTP request using
// https://docs.microsoft.com/en-us/windows/win32/winhttp/winhttprequest .
// The result is that you will be automatically authenticated as the
// currently logged in user
func WinHTTPRequest(
	method string, url string, reqHeaders map[string]string, reqBody string,
) (
	respBody string, respStatus int, respHeaders map[string]string,
) {
	log.Printf("HTTP %s %s", method, url)

	// lock OLE and initialize it (this is some windows API resource)
	oleMutex.Lock()
	defer oleMutex.Unlock()
	err := ole.CoInitialize(0)
	PanicOnErr(err)
	defer ole.CoUninitialize()

	winHTTP, err := oleutil.CreateObject("WinHTTP.WinHTTPRequest.5.1")
	PanicOnErr(err)
	req, err := winHTTP.QueryInterface(ole.IID_IDispatch)
	PanicOnErr(err)
	winErr, err := oleutil.CallMethod(req, "SetAutoLogonPolicy", 0)
	PanicOnErr(err)
	Expect(nil, winErr.Value(), "SetAutoLogonPolicy()")
	winErr, err = oleutil.CallMethod(req, "Open", method, url, false)
	PanicOnErr(err)
	Expect(nil, winErr.Value(), "Open()")

	// set request headers
	for key, value := range reqHeaders {
		winErr, err = oleutil.CallMethod(req, "SetRequestHeader", key, value)
		PanicOnErr(err)
		Expect(nil, winErr.Value(), "SetRequestHeader()")
	}

	// send with request body
	if len(reqBody) > 0 {
		winErr, err = oleutil.CallMethod(req, "Send", reqBody)
		PanicOnErr(err)
		Expect(nil, winErr.Value(), "Send()")
	} else {
		winErr, err = oleutil.CallMethod(req, "Send")
		PanicOnErr(err)
		Expect(nil, winErr.Value(), "Send()")
	}

	// get status code
	respStatus = int(oleutil.MustGetProperty(req, "Status").Value().(int32))

	// get headers
	headersObj, err := oleutil.CallMethod(req, "GetAllResponseHeaders")
	PanicOnErr(err)
	headersStr := headersObj.ToString()
	headersReader := bufio.NewReader(strings.NewReader(headersStr))
	tp := textproto.NewReader(headersReader)
	mimeHeaders, err := tp.ReadMIMEHeader()
	PanicOnErr(err)
	headers := http.Header(mimeHeaders)
	// remove duplicate headers
	respHeaders = make(map[string]string)
	for key, values := range headers {
		respHeaders[key] = values[0]
	}

	// get response body
	respBody = oleutil.MustGetProperty(req, "ResponseText").ToString()

	return
}
