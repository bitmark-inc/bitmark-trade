package asset

import "time"

type Asset struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Metadata    map[string]string `json:"metadata"`
	Fingerprint string            `json:"fingerprint"`
	Registrant  string            `json:"registrant"`
	Status      string            `json:"status"`
	BlockNumber int               `json:"block_number"`
	Sequence    int               `json:"offset"`
	CreatedAt   time.Time         `json:"created_at"`
}
