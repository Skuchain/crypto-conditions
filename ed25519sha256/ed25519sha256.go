package Ed25519Sha256

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/agl/ed25519"
	"github.com/jtremback/crypto-conditions/encoding"
)

var MaxFulfillmentLength uint64 = 9999999

func sliceTo64Byte(slice []byte) *[64]byte {
	if len(slice) == 64 {
		var array [64]byte
		copy(array[:], slice[:64])
		return &array
	}
	return &[64]byte{}
}

func sliceTo32Byte(slice []byte) *[32]byte {
	if len(slice) == 32 {
		var array [32]byte
		copy(array[:], slice[:32])
		return &array
	}
	return &[32]byte{}
}

type Fulfillment struct {
	PublicKey    []byte
	MessageId    []byte
	FixedMessage []byte
	// MaxDynamicMessageLength uint64
	DynamicMessage       []byte
	Signature            []byte
	MaxFulfillmentLength uint64
}

func (ful Fulfillment) Serialize(pubkey []byte, privkey [64]byte) string {
	payload := base64.URLEncoding.EncodeToString(bytes.Join([][]byte{
		encoding.MakeVarbyte(ful.PublicKey),
		encoding.MakeVarbyte(ful.MessageId),
		encoding.MakeVarbyte(ful.FixedMessage),
		// encoding.MakeVaruint(ful.MaxDynamicMessageLength),
		encoding.MakeVarbyte(ful.DynamicMessage),
		encoding.MakeVarbyte(ed25519.Sign(&privkey, append(ful.FixedMessage, ful.DynamicMessage...))[:]),
	}, []byte{}))

	return "cf:1:8:" + payload
}

func DeserializeFulfillment(ful string) (*Fulfillment, error) {
	parts := strings.Split(ful, ":")
	if len(parts) != 5 {
		return nil, errors.New("parsing error")
	}

	if parts[0] != "cf" {
		return nil, errors.New("fulfillments must start with \"cf\"")
	}

	if parts[1] != "1" {
		return nil, errors.New("must be protocol version 1")
	}

	if parts[2] != "8" {
		return nil, errors.New("not an Ed25519Sha256 condition")
	}

	b, err := base64.URLEncoding.DecodeString(parts[3])
	if err != nil {
		return nil, errors.New("parsing error")
	}

	// Get PublicKey
	length, offset := binary.Uvarint(b)
	pubkey, b := b[offset:][:length], b[offset:][length:]

	// Get MessageId
	length, offset = binary.Uvarint(b)
	messageId, b := b[offset:][:length], b[offset:][length:]

	// Get FixedMessage
	length, offset = binary.Uvarint(b)
	fixedMessage, b := b[offset:][:length], b[offset:][length:]

	// // Get MaxDynamicMessageLength
	// maxDynamicMessageLength, offset := binary.Uvarint(b)
	// b = b[offset:][length:]

	// Get DynamicMessage
	length, offset = binary.Uvarint(b)
	dynamicMessage, b := b[offset:][:length], b[offset:][length:]

	// Get Signature
	length, offset = binary.Uvarint(b)
	signature, b := b[offset:][:length], b[offset:][length:]

	// Check signature
	fullMessage := append(fixedMessage, dynamicMessage...)
	if !ed25519.Verify(sliceTo32Byte(pubkey), fullMessage, sliceTo64Byte(signature)) {
		return nil, errors.New("signature not valid")
	}

	// Get MaxFulfillmentLength
	maxFulfillmentLength, err := strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		return nil, errors.New("invalid maxFulfillmentLength")
	}

	return &Fulfillment{
		PublicKey:    pubkey,
		MessageId:    messageId,
		FixedMessage: fixedMessage,
		// MaxDynamicMessageLength: maxDynamicMessageLength,
		DynamicMessage:       dynamicMessage,
		Signature:            signature,
		MaxFulfillmentLength: maxFulfillmentLength,
	}, nil
}

func (ful Fulfillment) Condition() (string, error) {

	// payload := base64.URLEncoding.EncodeToString(

	hash := sha256.Sum256(bytes.Join([][]byte{
		encoding.MakeVarbyte(ful.PublicKey),
		encoding.MakeVarbyte(ful.MessageId),
		encoding.MakeVarbyte(ful.FixedMessage),
		// encoding.MakeVaruint(ful.MaxDynamicMessageLength),
	}, []byte{}))

	return "cc:1:8:" + base64.URLEncoding.EncodeToString(hash[:]) + ":" + fmt.Sprintf("%d", MaxFulfillmentLength), nil
}

type Condition struct {
	PublicKey    []byte
	MessageId    []byte
	FixedMessage []byte
	// MaxDynamicMessageLength uint64
	MaxFulfillmentLength uint64
}

func (cond Condition) Serialize(pubkey []byte, privkey [64]byte) string {
	payload := base64.URLEncoding.EncodeToString(bytes.Join([][]byte{
		encoding.MakeVarbyte(cond.PublicKey),
		encoding.MakeVarbyte(cond.MessageId),
		encoding.MakeVarbyte(cond.FixedMessage),
	}, []byte{}))

	return "cc:1:8:" + payload + fmt.Sprintf("%d", cond.MaxFulfillmentLength)
}
