package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	ansiReset     = "\033[0m"
	ansiRed       = "\033[31m"
	ansiGreen     = "\033[32m"
	ansiBgRed     = "\033[41m"
	ansiBgGreen   = "\033[42m"
	ansiBold      = "\033[1m"
	ansiUnderline = "\033[4m"
)

type testStruct struct {
	Method        string `json:"method"`
	RequestHeader []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	RequestBody    string `json:"requestBody"`
	ResponseStatus int    `json:"responseStatus"`
	ResponseHeader []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	ResponseBody string `json:"responseBody"`
}

type tableStruct struct {
	Name     string
	Received string
	Should   string
}

func main() {
	testsFile := flag.String("f", "", "the path to the tests file")

	flag.Parse()

	if *testsFile == "" {
		flag.PrintDefaults()

		os.Exit(1)
	}

	tests := make(map[string][]testStruct)

	fmt.Printf("Using File %s\n", *testsFile)

	file, err := os.Open(*testsFile)
	if err != nil {
		ansiPrint([]string{"red"}, fmt.Sprintf("could not open the %s file!!", *testsFile))
		os.Exit(1)
	}

	err = json.NewDecoder(file).Decode(&tests)

	testFailed := false

	fmt.Printf("Found %d Tests\n\n", len(tests))

	for url, v := range tests {
		for _, test := range v {
			ansiPrint([]string{"bold", "underline"}, fmt.Sprintf("%s %s", test.Method, url))

			client := &http.Client{}

			var request *http.Request
			var err error
			var errorTable []tableStruct

			if test.Method == "GET" {
				var body []byte
				request, err = http.NewRequest(http.MethodGet, url, bytes.NewBuffer(body))
			} else if test.Method == "POST" {
				var body []byte

				if test.RequestBody != "" {
					body = []byte(test.RequestBody)
				}

				request, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
			}

			if err != nil {
				ansiPrint([]string{"red"}, fmt.Sprintf("%s", err))

				continue
			}

			// Set Request Header
			if len(test.RequestHeader) > 0 {
				for _, header := range test.RequestHeader {
					request.Header.Set(header.Name, header.Value)
				}
			}

			response, err := client.Do(request)

			if err != nil {
				ansiPrint([]string{"red"}, fmt.Sprintf("%s", err))

				continue
			}

			// Check the Response Code
			if response.StatusCode != test.ResponseStatus {
				errorTable = append(errorTable, tableStruct{Name: "StatusCode", Received: strconv.Itoa(response.StatusCode), Should: strconv.Itoa(test.ResponseStatus)})
			}

			// Check the Response Header
			if len(test.ResponseHeader) > 0 {
				for _, header := range test.ResponseHeader {
					headerValue := response.Header.Get(header.Name)

					if headerValue != header.Value {
						errorTable = append(errorTable, tableStruct{Name: "Header " + header.Name, Received: headerValue, Should: header.Value})
					}
				}
			}

			// Check the Response Body
			if test.ResponseBody != "" {
				responseBody, err := ioutil.ReadAll(response.Body)

				if err != nil {
					fmt.Println(err)
				}

				if string(responseBody) != test.ResponseBody {
					errorTable = append(errorTable, tableStruct{Name: "Body", Received: string(responseBody), Should: test.ResponseBody})
				}
			}

			if len(errorTable) == 0 {
				ansiPrint([]string{"green"}, "successful")
			} else {
				testFailed = true

				ansiPrint([]string{"bold", "red"}, "failed")

				printTable(errorTable)
			}

			fmt.Println("")
		}
	}

	if testFailed == false {
		ansiPrint([]string{"bggreen"}, "All tests was successful :)")
	} else {
		ansiPrint([]string{"bgred"}, "Some tests failed :(")
	}
}

func ansiPrint(formatMap []string, text string) {
	var format string

	for _, v := range formatMap {
		if v == "red" {
			format += ansiRed
		} else if v == "green" {
			format += ansiGreen
		} else if v == "bgred" {
			format += ansiBgRed
		} else if v == "bggreen" {
			format += ansiBgGreen
		} else if v == "bold" {
			format += ansiBold
		} else if v == "underline" {
			format += ansiUnderline
		}
	}

	fmt.Printf("%s%s%s\n", format, text, ansiReset)
}

func printTable(table []tableStruct) {
	maxName := 4     // word Name
	maxReceived := 8 // word Received
	maxShould := 6   // word Should

	for _, v := range table {
		if len(v.Name) > maxName {
			maxName = len(v.Name)
		}

		if len(v.Received) > maxReceived {
			maxReceived = len(v.Received)
		}

		if len(v.Should) > maxShould {
			maxShould = len(v.Should)
		}
	}

	tabelwidth := maxName + maxReceived + maxShould + 10 // 4 for the spaces and 6 fore the padding around the text

	fmt.Println(strings.Repeat("-", tabelwidth))

	titleName := "Name" + strings.Repeat(" ", maxName-4)
	titleReceived := "Received" + strings.Repeat(" ", maxReceived-8)
	titleShould := "Should" + strings.Repeat(" ", maxShould-6)

	fmt.Printf("| %s | %s | %s |\n", titleName, titleReceived, titleShould)

	fmt.Println(strings.Repeat("-", tabelwidth))

	for _, v := range table {
		name := v.Name + strings.Repeat(" ", maxName-len(v.Name))
		received := v.Received + strings.Repeat(" ", maxReceived-len(v.Received))
		should := v.Should + strings.Repeat(" ", maxShould-len(v.Should))

		fmt.Printf("| %s | %s | %s |\n", name, received, should)
	}

	fmt.Println(strings.Repeat("-", tabelwidth))
}
