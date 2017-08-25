package bmservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bitmark-inc/logger"
)

type config struct {
	gateway  string
	registry string
	storage  string
}

var (
	client      *http.Client
	isTestChain bool
	cfg         *config
	log         *logger.L
)

type ServiceError struct {
	status int
	msg    string
}

func (e *ServiceError) Status() int {
	return e.status
}

func (e *ServiceError) Error() string {
	return e.msg
}

// Init the global HTTP client for interacting with bitmark services
func Init(chain string) {
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	isTestChain = chain != "production"

	switch chain {
	case "local":
		cfg = &config{
			gateway:  "https://bdgw.devel.bitmark.com",
			registry: "https://registry.devel.bitmark.com",
			storage:  "http://localhost:8900",
		}
	case "devel":
		cfg = &config{
			gateway:  "https://api.devel.bitmark.com",
			registry: "https://api.devel.bitmark.com",
			storage:  "https://storage.devel.bitmark.com",
		}
	case "test":
		cfg = &config{
			gateway:  "https://api.test.bitmark.com",
			registry: "https://api.test.bitmark.com",
			storage:  "https://storage.test.bitmark.com",
		}
	case "live":
		cfg = &config{
			gateway:  "https://api.bitmark.com",
			registry: "https://api.bitmark.com",
			storage:  "https://storage.live.bitmark.com",
		}
	}

	log = logger.New("bitmark-service")
}

func newJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(body)
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return nil, err
	}

	req, err := http.NewRequest(method, url, b)
	if nil != err {
		log.Errorf("%s: %v", url, err)
		return nil, err
	}

	return req, nil
}

func newFileUploadRequest(url, fieldname, filename string, filecontent []byte) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldname, filepath.Base(filename))
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return nil, err
	}
	_, err = part.Write(filecontent)
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func submitReqWithJSONResp(req *http.Request, reply interface{}) error {
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("%s: %v", req.URL.String(), err)
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("%s: %v", req.URL.String(), err)
		return err
	}

	if resp.StatusCode/100 != 2 {
		return &ServiceError{resp.StatusCode, string(data)}
	}

	if reply != nil {
		err = json.Unmarshal(data, reply)
		if err != nil {
			log.Errorf("%s: %s", req.URL.String(), string(data))
			return err
		}
	}

	return nil
}

func submitReqWithFileResp(url string) (string, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return "", nil, err
	}

	req.Header.Set("Content-Type", "multipart/form-data")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return "", nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("%s: %v", url, err)
		return "", nil, err
	}

	if resp.StatusCode/100 != 2 { // not 2xx HTTP status code
		return "", nil, &ServiceError{resp.StatusCode, string(data)}
	}

	_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	filename := strings.TrimSuffix(params["filename"], ".enc")

	return filename, data, nil
}
