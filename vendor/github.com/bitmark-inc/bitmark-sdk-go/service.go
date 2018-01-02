package bitmarksdk

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
)

type Service struct {
	client      *http.Client
	apiEndpoint string
	keyEndpoint string
}

func (s *Service) newAPIRequest(method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, s.apiEndpoint+path, body)
}

func (s *Service) newSignedAPIRequest(method, path string, body io.Reader, acct *Account, parts ...string) (*http.Request, error) {
	req, err := http.NewRequest(method, s.apiEndpoint+path, body)
	if err != nil {
		return nil, err
	}

	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts = append(parts, acct.AccountNumber(), ts)
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.AuthKey.Sign([]byte(message)))

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

func (s *Service) uploadAsset(acct *Account, af *AssetFile) error {
	body := new(bytes.Buffer)

	bodyWriter := multipart.NewWriter(body)
	bodyWriter.WriteField("asset_id", af.Id())
	bodyWriter.WriteField("accessibility", string(af.Accessibility))

	fileWriter, err := bodyWriter.CreateFormFile("file", af.Name)
	if err != nil {
		return err
	}

	switch af.Accessibility {
	case Public:
		if _, e := fileWriter.Write(af.Content); e != nil {
			return err
		}
	case Private:
		dataKey, e := NewDataKey()
		if e != nil {
			return err
		}
		encryptedContent, e := dataKey.Encrypt(af.Content)
		if e != nil {
			return err
		}
		sessData, e := createSessionData(acct, dataKey, acct.EncrKey.PublicKeyBytes())
		if e != nil {
			return err
		}
		if _, e := fileWriter.Write(encryptedContent); e != nil {
			return err
		}
		bodyWriter.WriteField("session_data", sessData.String())
	}

	err = bodyWriter.Close()
	if err != nil {
		return err
	}

	req, _ := s.newSignedAPIRequest("POST", "/v1/assets", body, acct, "uploadAsset", af.Id())
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	_, err = s.submitRequest(req, nil)
	return err
}

func (s *Service) getAssetAccess(acct *Account, bitmarkId string) (*accessByOwnership, error) {
	req, _ := s.newSignedAPIRequest("GET", fmt.Sprintf("/v1/bitmarks/%s/asset", bitmarkId), nil, acct, "downloadAsset", bitmarkId)

	var result accessByOwnership
	if _, err := s.submitRequest(req, &result); err != nil {
		return nil, err
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

	_, params, _ := mime.ParseMediaType(resp.Header["Content-Disposition"][0])
	filename := params["filename"]

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return filename, data, nil
}

func (s *Service) createIssueTx(asset *AssetRecord, issues []*IssueRecord) ([]string, error) {
	b := map[string]interface{}{
		"issues": issues,
	}
	if asset != nil {
		b["assets"] = []*AssetRecord{asset}
	}
	body := toJSONRequestBody(b)
	req, _ := s.newAPIRequest("POST", "/v1/issue", body)

	result := make([]transaction, 0)
	if _, err := s.submitRequest(req, &result); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, b := range result {
		bitmarkIds = append(bitmarkIds, b.TxId)
	}
	return bitmarkIds, nil
}

func (s *Service) createTransferTx(record *TransferRecord) (string, error) {
	body := toJSONRequestBody(map[string]interface{}{
		"transfer": record,
	})
	req, _ := s.newAPIRequest("POST", "/v1/transfer", body)

	result := make([]transaction, 0)
	if _, err := s.submitRequest(req, &result); err != nil {
		return "", err
	}

	return result[0].TxId, nil
}

func (s *Service) createCountersignTransferTx(record *CountersignedTransferRecord) (string, error) {
	body := toJSONRequestBody(map[string]interface{}{
		"transfer": record,
	})
	req, _ := s.newAPIRequest("POST", "/v1/transfer", body)

	result := make([]transaction, 0)
	if _, err := s.submitRequest(req, &result); err != nil {
		return "", err
	}

	return result[0].TxId, nil
}

func (s *Service) addSessionData(acct *Account, bitmarkId, receiver string, data *SessionData) error {
	body := toJSONRequestBody(map[string]interface{}{
		"bitmark_id":   bitmarkId,
		"owner":        receiver,
		"session_data": data,
	})
	req, _ := s.newSignedAPIRequest("POST", "/v2/session", body, acct, "updateSession", data.String())

	_, err := s.submitRequest(req, nil)
	return err
}

func (s *Service) registerEncPubkey(acct *Account) error {
	signature := hex.EncodeToString(acct.AuthKey.Sign(acct.EncrKey.PublicKeyBytes()))
	body := toJSONRequestBody(map[string]interface{}{
		"encryption_pubkey": fmt.Sprintf("%064x", acct.EncrKey.PublicKeyBytes()),
		"signature":         signature,
	})
	req, _ := s.newAPIRequest("POST", fmt.Sprintf("/v1/encryption_keys/%s", acct.AccountNumber()), body)

	_, err := s.submitRequest(req, nil)
	return err
}

func (s *Service) getEncPubkey(acctNo string) ([]byte, error) {
	req, _ := s.newKeyRequest("GET", fmt.Sprintf("/%s", acctNo), nil)

	var result struct {
		Key string `json:"encryption_pubkey"`
	}
	if _, err := s.submitRequest(req, &result); err != nil {
		return nil, err
	}

	return hex.DecodeString(result.Key)
}

func (s *Service) getBitmark(bitmarkId string) (*Bitmark, error) {
	req, _ := s.newAPIRequest("GET", "/v1/bitmarks/"+bitmarkId, nil)

	var result struct {
		Bitmark *Bitmark
	}
	_, err := s.submitRequest(req, &result)
	return result.Bitmark, err
}

func toJSONReqBody(data map[string]interface{}) io.Reader {
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
