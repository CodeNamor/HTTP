package apiclient

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kr/pretty"
)

func mockGetClient() *http.Client {
	// in a real getClient, setup timeouts and transport options
	return &http.Client{}
}

func ExampleInstrumentHTTPRequest() {
	url := "https://postman-echo.com/get"
	httpClient := mockGetClient()
	req, err := http.NewRequest(http.MethodGet, url, nil /* body */)
	if err != nil {
		pretty.Println(err)
		return
	}
	req, requestDone := InstrumentHTTPRequest(req)
	res, err := httpClient.Do(req)
	if err != nil {
		pretty.Println(err)
		return
	}
	defer requestDone()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			pretty.Println(err)
		}
	}(res.Body)

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		pretty.Println(err)
		return
	}

	pretty.Println("result:", string(bytes))
}

func ExampleAddExpVarHandlerToRouter() {
	router := mux.NewRouter().StrictSlash(true)
	AddExpVarHandlerToRouter(router, "/debug/vars")
}
