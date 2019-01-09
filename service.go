package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

type Service struct {
	client      *http.Client
	apiEndpoint string
	keyEndpoint string
}

func (s *Service) newAPIRequest(method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, s.apiEndpoint+path, body)
}

func (s *Service) newSignedAPIRequest(method, path string, body io.Reader, acct account.Account, parts ...string) (*http.Request, error) {
	req, err := http.NewRequest(method, s.apiEndpoint+path, body)
	if err != nil {
		return nil, err
	}

	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts = append(parts, acct.AccountNumber(), ts)
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.Sign([]byte(message)))

	req.Header.Add("requester", acct.AccountNumber())
	req.Header.Add("timestamp", ts)
	req.Header.Add("signature", sig)

	return req, nil
}

func (s *Service) newKeyRequest(method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, s.keyEndpoint+path, body)
}

func (s *Service) submitRequest(req *http.Request, result interface{}) ([]byte, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		var se ServiceError
		if e := json.Unmarshal(data, &se); e != nil {
			return nil, fmt.Errorf("unexpected response: %s", string(data))
		}
		return nil, &se
	}

	if result != nil {
		if err = json.Unmarshal(data, result); err != nil {
			return nil, fmt.Errorf("unexpected response: %s", string(data))
		}
	}

	return data, nil
}

func (s *Service) uploadAsset(acct account.Account, assetId string, fileName string, fileContent []byte) error {
	body := new(bytes.Buffer)

	bodyWriter := multipart.NewWriter(body)
	bodyWriter.WriteField("asset_id", assetId)
	bodyWriter.WriteField("accessibility", "private") // NOTE: always private in bitmark-trade

	fileWriter, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	dataKey, e := NewDataKey()
	if e != nil {
		return err
	}
	encryptedContent, e := dataKey.Encrypt(fileContent)
	if e != nil {
		return err
	}
	encrKey := getEncrKey(acct)
	sessData, e := createSessionData(acct, dataKey, encrKey.PublicKeyBytes())
	if e != nil {
		return err
	}
	if _, e := fileWriter.Write(encryptedContent); e != nil {
		return err
	}
	bodyWriter.WriteField("session_data", sessData.String())

	err = bodyWriter.Close()
	if err != nil {
		return err
	}

	req, _ := s.newSignedAPIRequest("POST", "/v1/assets", body, acct, "uploadAsset", assetId)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	_, err = s.submitRequest(req, nil)
	return err
}

type access struct {
	URL      string       `json:"url"`
	SessData *SessionData `json:"session_data"`
	Sender   string       `json:"sender"`
}

func (s *Service) getAssetAccess(acct account.Account, bitmarkId string) (*access, error) {
	req, _ := s.newSignedAPIRequest("GET", fmt.Sprintf("/v1/bitmarks/%s/asset", bitmarkId), nil, acct, "downloadAsset", bitmarkId)

	var result access
	if _, err := s.submitRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get the asset access: %s", err.Error())
	}

	return &result, nil
}

func (s *Service) getAssetContent(url string) (string, []byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var filename string
	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	name, ok := params["filename"]
	if err == nil && ok {
		filename = name
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return filename, data, nil
}

func (s *Service) addSessionData(acct account.Account, bitmarkId, receiver string, data *SessionData) error {
	body := toJSONRequestBody(map[string]interface{}{
		"bitmark_id":   bitmarkId,
		"owner":        receiver,
		"session_data": data,
	})
	req, _ := s.newSignedAPIRequest("POST", "/v2/session", body, acct, "updateSession", data.String())

	if _, err := s.submitRequest(req, nil); err != nil {
		return fmt.Errorf("failed to add session data for %s: %s", receiver, err.Error())
	}
	return nil
}

func (s *Service) registerEncPubkey(acct account.Account) error {
	encrKey := getEncrKey(acct)
	signature := hex.EncodeToString(acct.Sign(encrKey.PublicKeyBytes()))
	body := toJSONRequestBody(map[string]interface{}{
		"encryption_pubkey": fmt.Sprintf("%064x", encrKey.PublicKeyBytes()),
		"signature":         signature,
	})
	req, _ := s.newAPIRequest("POST", fmt.Sprintf("/v1/encryption_keys/%s", acct.AccountNumber()), body)

	if _, err := s.submitRequest(req, nil); err != nil {
		return fmt.Errorf("failed to register the encyrption public key for %s: %s", acct.AccountNumber(), err.Error())
	}
	return nil
}

func (s *Service) getEncPubkey(acctNo string) ([]byte, error) {
	req, _ := s.newKeyRequest("GET", fmt.Sprintf("/%s", acctNo), nil)

	var result struct {
		Key string `json:"encryption_pubkey"`
	}
	if _, err := s.submitRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get the encyrption public key for %s: %s", acctNo, err.Error())
	}

	return hex.DecodeString(result.Key)
}

func toJSONRequestBody(data map[string]interface{}) io.Reader {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(data)
	return body
}

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (se *ServiceError) Error() string {
	return fmt.Sprintf("[%d] %s", se.Code, se.Message)
}
