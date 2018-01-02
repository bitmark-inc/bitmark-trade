package bitmarksdk

type transaction struct {
	TxId string `json:"txId"`
}

type assetAccessMeta struct {
	URL      string       `json:"url"`
	SessData *SessionData `json:"session_data"`
}

type accessByOwnership struct {
	assetAccessMeta
	Sender string `json:"sender"`
}

type accessByRenting struct {
	assetAccessMeta
	AssetId  string `json:"asset_id"`
	Owner    string `json:"owner"`
	Duration uint   `json:"duration"`
	ExpTime  int64  `json:"expiration_time"`
}
