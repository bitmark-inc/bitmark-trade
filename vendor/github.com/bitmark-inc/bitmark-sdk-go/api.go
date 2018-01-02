package bitmarksdk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type APIRequest struct {
	*http.Request
}

func (r APIRequest) Sign(acct *Account, action, resource string) {
	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts := []string{
		action,
		resource,
		acct.AccountNumber(),
		ts,
	}
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.AuthKey.Sign([]byte(message)))

	r.Header.Add("requester", acct.AccountNumber())
	r.Header.Add("timestamp", ts)
	r.Header.Add("signature", sig)
}

func newAPIRequest(method, url string, body io.Reader) (*APIRequest, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return &APIRequest{r}, nil
}

type APIClient struct {
	client      *http.Client
	apiServer   string
	assetServer string
}

func NewAPIClient(network Network, client *http.Client) *APIClient {
	api := &APIClient{
		client: client,
	}

	switch network {
	case Testnet:
		api.apiServer = "api.test.bitmark.com"
		api.assetServer = "assets.test.bitmark.com"
	case Livenet:
		api.apiServer = "api.bitmark.com"
		api.assetServer = "assets.bitmark.com"
	}

	return api
}

func (api *APIClient) submitRequest(req *APIRequest, reply interface{}) ([]byte, error) {
	resp, err := api.client.Do(req.Request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		var m struct {
			Message string `json:"message"`
		}
		if e := json.Unmarshal(data, &m); e != nil {
			return nil, errors.New(string(data))
		}
		return nil, errors.New(m.Message)
	}

	if reply != nil {
		err = json.Unmarshal(data, reply)
		if err != nil {
			return nil, fmt.Errorf("json decode error: %s, data: %s", err.Error(), string(data))
		}
	}

	return data, nil
}

// [ASSET] - upload a asset file; if private asset, encryption needs to be applied
func (api *APIClient) uploadAsset(acct *Account, af *AssetFile) error {
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

	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v1/assets",
	}
	req, err := newAPIRequest("POST", u.String(), body)

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	req.Sign(acct, "uploadAsset", af.Id())

	_, err = api.submitRequest(req, nil)
	return err
}

// [ASSET] - get the access information of an asset
//
// -> public:  w/o session data
// -> private: w/  session data
func (api *APIClient) getAssetAccess(acct *Account, bitmarkId string) (*accessByOwnership, error) {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   fmt.Sprintf("/v1/bitmarks/%s/asset", bitmarkId),
	}

	req, _ := newAPIRequest("GET", u.String(), nil)
	req.Sign(acct, "downloadAsset", bitmarkId)

	var result accessByOwnership
	if _, err := api.submitRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// [ASSET] - get the asset file content
//
// -> public:  plaintext
// -> private: ciphertext
func (api *APIClient) getAssetContent(url string) (string, []byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := api.client.Do(req)
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

// [ASSET] - add the session data for the bitmark receiver
func (api *APIClient) addSessionData(acct *Account, bitmarkId, receiver string, data *SessionData) error {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v2/session",
	}
	body := toJSONRequestBody(map[string]interface{}{
		"bitmark_id":   bitmarkId,
		"owner":        receiver,
		"session_data": data,
	})
	req, _ := newAPIRequest("POST", u.String(), body)
	req.Sign(acct, "updateSession", data.String())

	_, err := api.submitRequest(req, nil)
	return err
}

// [ENCRYPTION PUBLIC KEY] - register encryption public key of an account
func (api *APIClient) registerEncPubkey(acct *Account) error {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   fmt.Sprintf("/v1/encryption_keys/%s", acct.AccountNumber()),
	}
	signature := hex.EncodeToString(acct.AuthKey.Sign(acct.EncrKey.PublicKeyBytes()))
	body := toJSONRequestBody(map[string]interface{}{
		"encryption_pubkey": fmt.Sprintf("%064x", acct.EncrKey.PublicKeyBytes()),
		"signature":         signature,
	})
	req, _ := newAPIRequest("POST", u.String(), body)

	_, err := api.submitRequest(req, nil)
	return err
}

// [ENCRYPTION PUBLIC KEY] - get the encryption public key of an account
func (api *APIClient) getEncPubkey(acctNo string) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("key.%s", api.assetServer),
		Path:   fmt.Sprintf("/%s", acctNo),
	}
	req, _ := newAPIRequest("GET", u.String(), nil)

	var result struct {
		Key string `json:"encryption_pubkey"`
	}
	if _, err := api.submitRequest(req, &result); err != nil {
		return nil, err
	}

	return hex.DecodeString(result.Key)
}

// [TRANSACTION] issue bitmarks
func (api *APIClient) issue(asset *AssetRecord, issues []*IssueRecord) ([]string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v1/issue",
	}

	b := map[string]interface{}{
		"issues": issues,
	}
	if asset != nil {
		b["assets"] = []*AssetRecord{asset}
	}
	body := toJSONRequestBody(b)
	req, _ := newAPIRequest("POST", u.String(), body)

	result := make([]transaction, 0)
	if _, err := api.submitRequest(req, &result); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, b := range result {
		bitmarkIds = append(bitmarkIds, b.TxId)
	}

	return bitmarkIds, nil
}

// [TRANSACTION] transfer a bitmark
func (api *APIClient) transfer(record *TransferRecord) (string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v1/transfer",
	}
	body := toJSONRequestBody(map[string]interface{}{
		"transfer": record,
	})
	req, _ := newAPIRequest("POST", u.String(), body)

	result := make([]transaction, 0)
	if _, err := api.submitRequest(req, &result); err != nil {
		return "", err
	}

	return result[0].TxId, nil
}

// [REGISTRY] query a bitmark
func (api *APIClient) getBitmark(bitmarkId string) (*Bitmark, error) {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v1/bitmarks/" + bitmarkId,
	}
	req, _ := newAPIRequest("GET", u.String(), nil)

	var result struct {
		Bitmark *Bitmark
	}
	_, err := api.submitRequest(req, &result)
	return result.Bitmark, err
}

func (api *APIClient) getAsset(assetId string) (*Asset, error) {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v1/assets/" + assetId,
	}
	req, _ := newAPIRequest("GET", u.String(), nil)

	var result struct {
		Asset Asset
	}
	_, err := api.submitRequest(req, &result)
	if err != nil && err.Error() == "Not Found" {
		return nil, nil
	}
	return &result.Asset, err
}

func (api *APIClient) updateLease(acct *Account, bitmarkId, receiver string, days uint, data *SessionData) error {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v2/leases/" + bitmarkId,
	}
	u = url.URL{
		Scheme: "http",
		Host:   "0.0.0.0:8087",
		Path:   "/v2/leases/" + bitmarkId,
	}
	body := toJSONRequestBody(map[string]interface{}{
		"renter":       receiver,
		"days":         days,
		"session_data": data,
	})
	req, _ := newAPIRequest("PUT", u.String(), body)
	req.Sign(acct, "updateLease", bitmarkId)

	_, err := api.submitRequest(req, nil)
	return err
}

func (api *APIClient) listLeases(acct *Account) ([]*accessByRenting, error) {
	u := url.URL{
		Scheme: "https",
		Host:   api.apiServer,
		Path:   "/v2/leases",
	}
	u = url.URL{
		Scheme: "http",
		Host:   "0.0.0.0:8087",
		Path:   "/v2/leases",
	}
	req, _ := newAPIRequest("GET", u.String(), nil)
	req.Sign(acct, "listLeases", "")
	var result struct {
		AssetsAccess []*accessByRenting
	}
	_, err := api.submitRequest(req, &result)

	return result.AssetsAccess, err
}

func toJSONRequestBody(data map[string]interface{}) io.Reader {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(data)
	return body
}
