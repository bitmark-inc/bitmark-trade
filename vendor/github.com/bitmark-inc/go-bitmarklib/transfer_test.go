package bitmarklib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTransfer(t *testing.T) {
	transfer, err := NewTransfer("6776599a5fd4f2ade1ca87ee5fffd0295bb69b1969ffab1ec042a5f71ef74209", "fqN6WnjUaekfrqBvvmsjVskoqXnhJ632xJPHzdSgReC6bhZGuP", true)
	assert.NoError(t, err)
	assert.NotNil(t, transfer)
	assert.EqualValues(t, []byte{}, transfer.Signature)
}

func TestTransferSign(t *testing.T) {
	owner, err := NewKeyPair(true, ED25519)
	assert.NoError(t, err)

	transfer, err := NewTransfer("6776599a5fd4f2ade1ca87ee5fffd0295bb69b1969ffab1ec042a5f71ef74209", "fqN6WnjUaekfrqBvvmsjVskoqXnhJ632xJPHzdSgReC6bhZGuP", true)
	err = transfer.Sign(owner)
	assert.NoError(t, err)
	assert.NotNil(t, transfer.Signature)
}
