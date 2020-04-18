package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type Just struct {
	Client      *http.Client
	RetryIf429  bool
	RequestType RequestType
}

type SaveToFileConfig struct {
	Filepath string
	ReplaceData bool
	AppendCustomText *string
	Onset bool
}

func (r *Just) MakeR(method string, url string, headers map[string]string, body interface{}, file *os.File) (*http.Response, error) {
	var request *http.Request
	var err error
	switch r.RequestType {
	case JSON:
		request, err = r.generateJsonRequest(method, url, headers, body)
	case XML:
		request, err = r.generateXmlRequest(url, headers, body)
	}
	response, err := r.Client.Do(request)
	if err != nil {
		return nil, err
	}
	if r.RetryIf429 && response.StatusCode == 429 {
		return r.MakeR(method, url, headers, body, file)
	}
	return response, nil
}

func (r *Just) ResponseToStruct(response *http.Response, result interface{}) error {
	if result != nil {
		switch r.RequestType {
		case JSON:
			return json.NewDecoder(response.Body).Decode(result)
		case XML:
			return xml.NewDecoder(response.Body).Decode(result)
		default:
			return errors.New("Just: missed Just request type configuration ")
		}
	}
	return errors.New("Just: object [result] can not be empty ")
}

func (r *Just) ResponseToFile(response *http.Response, conf SaveToFileConfig) error {
	if conf.ReplaceData {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.New("Just: " + err.Error())
		}
		f, err := os.OpenFile(conf.Filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.New("Just: " + err.Error())
		}
		if conf.Onset {
			_, err = f.Write(append([]byte(*conf.AppendCustomText)[:], body[:]...))
		} else {
			_, err = f.Write(append(body[:], []byte(*conf.AppendCustomText)[:]...))
		}

		if err != nil {
			return errors.New("Just: " + err.Error())
		}

		defer f.Close()
		return nil
	} else {
		file, err := os.Open(conf.Filepath)
		if err != nil {
			return errors.New("Just: " + err.Error())
		}
		defer file.Close()
		_, err = io.Copy(file, response.Body)
		if err != nil {
			return errors.New("Just: " + err.Error())
		}
		return nil
	}
}

func (r *Just) generateJsonRequest(method string, url string, headers map[string]string, body interface{}) (request *http.Request, err error) {
	if body != nil {
		requestBytes, err := json.Marshal(body)
		if err != nil {
			return
		}
		request, err = http.NewRequest(method, url, bytes.NewReader(requestBytes))
		if err != nil {
			return
		}
	} else {
		request, err = http.NewRequest(method, url, nil)
		if err != nil {
			return
		}
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	return
}

func (r *Just) generateXmlRequest(url string, headers map[string]string, body interface{}) (request *http.Request, err error) {
	_, ok := headers["SOAPAction"]
	if !ok {
		return nil, errors.New("SOAPAction is not configured ")
	}
	if body != nil {
		requestBytes, err := xml.MarshalIndent(body, " ", " ")
		if err != nil {
			return
		}
		request, err = http.NewRequest("POST", url, bytes.NewReader(requestBytes))
		if err != nil {
			return
		}
	} else {
		request, err = http.NewRequest("POST", url, nil)
		if err != nil {
			return
		}
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	return
}