package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)
import "archive/zip"

type PostmanKV struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}
type PostmanForm struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
	Type  string  `json:"type"`
	Src   string  `json:"src,omitempty"`
}
type Postman struct {
	Info struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	} `json:"info"`
	Item []*PostmanItem `json:"item,omitempty"`
}
type PostmanResponse struct {
	ResponseTime       string         `json:"responseTime"`
	PostmanPreviewLang string         `json:"_postman_previewlanguage,omitempty"`
	OriginalRequest    PostmanRequest `json:"originalRequest,omitempty"`
	Header             *[]PostmanKV   `json:"header"`
	Body               string         `json:"body,omitempty"`
	Code               int            `json:"code,omitempty"`
	Status             string         `json:"status,omitempty"`
	Name               string         `json:"name,omitempty"`
}
type PostmanRequest struct {
	Method string       `json:"method"`
	Header *[]PostmanKV `json:"header"`
	Body   PostmanRaw   `json:"body"`
	URL    struct {
		Raw      string       `json:"raw"`
		Protocol string       `json:"protocol"`
		Host     []string     `json:"host"`
		Path     []string     `json:"path"`
		Query    *[]PostmanKV `json:"query,omitempty""`
		Port     string       `json:"port,omitempty"`
	} `json:"url"`
}

type PostmanItem struct {
	Name     string             `json:"name"`
	Request  PostmanRequest     `json:"request"`
	Response *[]PostmanResponse `json:"response"`
}

type DictionaryTree struct {
	Files    []string
	App      string
	URL      string
	Duration string
	Time     string
	Protocol string
}

type PostmanRaw struct {
	Mode       string             `json:"mode"`
	Raw        string             `json:"raw,omitempty"`
	Urlencoded *[]PostmanForm     `json:"urlencoded,omitempty"`
	Formdata   *[]PostmanForm     `json:"formdata,omitempty"`
	Options    *PostmanRawOptions `json:"options,omitempty"`
}
type PostmanRawOptions struct {
	Raw struct {
		Language string `json:"language"`
	} `json:"raw"`
}
type HttpCanaryJson struct {
	App      string `json:"app"`
	Duration string `json:"duration"`
	//Headers    Headers `json:"headers"`
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	//RemoteIP   string  `json:"remoteIp"`
	//RemotePort int     `json:"remotePort"`
	SessionID string `json:"sessionId"`
	Time      string `json:"time"`
	URL       string `json:"url"`
}

func createPostman(appName string) Postman {
	return Postman{
		Info: struct {
			Name   string `json:"name"`
			Schema string `json:"schema"`
		}{Name: appName, Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},
	}
}
func toPostman(inputFile string, output string) error {

	directoryDict := make(map[string]*DictionaryTree)
	FileDict := make(map[string]*zip.File)
	PostmanDict := make(map[string]*Postman)
	file, err := zip.OpenReader(inputFile)
	if err != nil {
		return err
	}
	for _, i2 := range file.File {
		//fmt.Println(i2.Name)
		fileName := i2.Name
		if strings.Contains(fileName, "/") {
			path := strings.Split(fileName, "/")

			if directoryDict[path[0]] == nil {
				directoryDict[path[0]] = &DictionaryTree{}
			}
			directoryDict[path[0]].Files = append(directoryDict[path[0]].Files, path[1])
			FileDict[fileName] = i2
			if strings.Contains(fileName, "request.json") {
				closer, err := i2.Open()
				if err != nil {
					return err
				}
				var content []byte
				content, err = ioutil.ReadAll(closer)

				if err != nil {
					fmt.Println(err)
					return err
				}

				var httpCanaryJson HttpCanaryJson
				err = json.Unmarshal(content, &httpCanaryJson)
				if err != nil {
					fmt.Println(err)
					return err
				}
				if httpCanaryJson.App != "" && PostmanDict[httpCanaryJson.App] == nil {
					postman := createPostman(httpCanaryJson.App)
					PostmanDict[httpCanaryJson.App] = &postman
				}
				directoryDict[path[0]].App = httpCanaryJson.App
				directoryDict[path[0]].URL = httpCanaryJson.URL
				directoryDict[path[0]].Duration = httpCanaryJson.Duration
				directoryDict[path[0]].Time = httpCanaryJson.Time
				directoryDict[path[0]].Protocol = httpCanaryJson.Protocol

			}
		}
	}
	defer file.Close()
	//fmt.Println(directoryDict)
	for i, i2 := range directoryDict {
		var request *http.Request
		var postmanItem *PostmanItem
		var postmanRequest PostmanRequest
		if _, ok := FileDict[i+"/request.hcy"]; ok {
			open, err := FileDict[i+"/request.hcy"].Open()
			if err != nil {
				fmt.Println(err)
				return err
			}
			var buf *bufio.Reader
			if i2.Protocol == "h2" {
				reg, _ := regexp.Compile(`h2\s*$`)
				content, err := ioutil.ReadAll(open)

				if err != nil {
					fmt.Println(err)
					return err
				}
				lines := strings.Split(string(content), "\n")
				lines[0] = reg.ReplaceAllString(lines[0], "HTTP/1.1")
				buf = bufio.NewReader(bytes.NewBufferString(strings.Join(lines, "\n")))

			} else {
				buf = bufio.NewReader(open)
			}

			request, err = http.ReadRequest(buf)
			if err != nil {
				fmt.Println(err)
				return err
			}

			parse, err := url.Parse(i2.URL)
			if err != nil {
				fmt.Println(err)
				return err
			}
			//fmt.Println(parse)
			host := parse.Host
			port := ""
			if strings.Contains(host, ":") {
				port = strings.Split(host, ":")[1]
				host = strings.Split(host, ":")[0]

			}
			postmanRequest = PostmanRequest{Method: request.Method, URL: struct {
				Raw      string       `json:"raw"`
				Protocol string       `json:"protocol"`
				Host     []string     `json:"host"`
				Path     []string     `json:"path"`
				Query    *[]PostmanKV `json:"query,omitempty""`
				Port     string       `json:"port,omitempty"`
			}{Raw: i2.URL, Protocol: parse.Scheme, Port: port, Host: strings.Split(host, "."), Path: strings.Split(parse.Path, "/")[1:]},
			}
			postmanItem = &PostmanItem{Name: request.RequestURI, Request: postmanRequest}
			if parse.RawQuery != "" {
				query, err := url.ParseQuery(parse.RawQuery)
				if err == nil {
					var queryList []PostmanKV
					for s, i3 := range query {
						if len(i3) > 0 {
							queryList = append(queryList, PostmanKV{s, &i3[len(i3)-1]})
						} else {
							queryList = append(queryList, PostmanKV{s, nil})

						}
					}
					postmanItem.Request.URL.Query = &queryList
				}
			}
			if len(request.Header) > 0 {
				var headers []PostmanKV
				for s, i3 := range request.Header {
					if len(i3) > 0 {
						headers = append(headers, PostmanKV{s, &i3[0]})

					} else {
						headers = append(headers, PostmanKV{s, nil})
					}
				}
				postmanItem.Request.Header = &headers
			}

			if err != nil {
				return err
			}
			contentType := request.Header.Get("Content-Type")
			switch {
			case strings.Contains(contentType, "form-data"):
				err = request.ParseMultipartForm(10240)
				if err == nil {
					var kvs []PostmanForm
					postmanItem.Request.Body.Mode = "formdata"
					for s, i3 := range request.PostForm {
						if len(i3) > 0 {
							kvs = append(kvs, PostmanForm{Key: s, Value: &i3[0], Type: "text"})
						} else {
							kvs = append(kvs, PostmanForm{Key: s, Type: "text"})

						}
					}
					if len(request.MultipartForm.File) > 0 {

						for s, _ := range request.MultipartForm.File {
							kvs = append(kvs, PostmanForm{Key: s, Type: "file"})
						}

					}
					postmanItem.Request.Body.Formdata = &kvs

				}
			case strings.Contains(contentType, "x-www-form-urlencoded"):
				err = request.ParseForm()
				if err == nil {
					var kvs []PostmanForm
					postmanItem.Request.Body.Mode = "urlencoded"
					for s, i3 := range request.PostForm {
						if len(i3) > 0 {
							kvs = append(kvs, PostmanForm{Key: s, Value: &i3[0], Type: "text"})
						} else {
							kvs = append(kvs, PostmanForm{Key: s, Type: "text"})

						}
					}
					postmanItem.Request.Body.Urlencoded = &kvs

				}
			default:
				postmanItem.Request.Body.Mode = "raw"
				var content []byte
				content, err = ioutil.ReadAll(request.Body)
				postmanItem.Request.Body.Raw = string(content)
				previewlang := Previewlang(contentType)
				postmanItem.Request.Body.Options = &PostmanRawOptions{Raw: struct {
					Language string `json:"language"`
				}(struct{ Language string }{Language: previewlang})}
			}
			if postmanItem.Request.Body.Mode == "" {
				postmanItem.Request.Body.Mode = "raw"
				var content []byte
				content, err = ioutil.ReadAll(request.Body)
				postmanItem.Request.Body.Raw = string(content)
			}

			PostmanDict[i2.App].Item = append(PostmanDict[i2.App].Item, postmanItem)

		}
		if _, ok := FileDict[i+"/response.hcy"]; ok {
			open, err := FileDict[i+"/response.hcy"].Open()
			if err != nil {
				fmt.Println(err)
				return err
			}
			var buf *bufio.Reader
			if i2.Protocol == "h2" {
				content, err := ioutil.ReadAll(open)

				if err != nil {
					fmt.Println(err)
					return err
				}
				lines := strings.Split(string(content), "\n")
				lines[0] = strings.Replace(lines[0], "h2", "HTTP/1.1", 1)
				buf = bufio.NewReader(bytes.NewBufferString(strings.Join(lines, "\n")))

			} else {
				buf = bufio.NewReader(open)
			}
			response, err := http.ReadResponse(buf, request)

			if err != nil {
				fmt.Println(err)
				return err
			}
			var postmanResponseWrapper []PostmanResponse
			responsePostman := PostmanResponse{
				ResponseTime: i2.Duration,
				Name:         fmt.Sprintf("%s,Imported From HttpCanary", i2.Time),
			}
			if len(response.Header) > 0 {
				var headers []PostmanKV
				for s, i3 := range response.Header {
					if len(i3) > 0 {
						headers = append(headers, PostmanKV{s, &i3[0]})

					} else {
						headers = append(headers, PostmanKV{s, nil})
					}
				}
				responsePostman.Header = &headers
			}
			if response.ContentLength == 0 {
				continue
			}
			reader := response.Body
			if response.Header.Get("Content-Encoding") == "gzip" {
				reader, err = gzip.NewReader(reader)
				if err != nil {
					fmt.Println(err)

				}
			}
			var content []byte
			content, err = ioutil.ReadAll(reader)
			responsePostman.Body = string(content)
			responsePostman.Code = response.StatusCode
			responsePostman.Status = response.Status
			responsePostman.PostmanPreviewLang = Previewlang(response.Header.Get("Content-Type"))
			responsePostman.OriginalRequest = postmanItem.Request
			postmanResponseWrapper = append(postmanResponseWrapper, responsePostman)
			postmanItem.Response = &postmanResponseWrapper
		}
	}
	if !strings.HasSuffix(output, "/") {
		output += "/"
	}
	for s, postman := range PostmanDict {
		marshal, err := json.Marshal(postman)
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s%s_collection.json", output, s), marshal, 0644)
		if err != nil {
			fmt.Println(err)
			return err
		}

	}
	return nil
}

func Previewlang(contentType string) string {
	if strings.Contains(contentType, "json") {
		return "json"
	}
	if strings.Contains(contentType, "html") {
		return "html"
	}
	if strings.Contains(contentType, "javascript") {
		return "javascript"
	}
	if strings.Contains(contentType, "xml") {
		return "xml"
	}
	return "auto"
}
