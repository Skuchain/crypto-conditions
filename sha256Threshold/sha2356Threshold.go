// Generates and parses Ed25519-Sha256 Crypto Conditions
package Ed25519Sha256

import (
	"bytes"
	"encoding/base64"
	"errors"
	"sort"
	"strings"

	"github.com/jtremback/crypto-conditions/encoding"
)

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

type WeightedString struct {
	Weight uint64
	String string
}

type WeightedStrings []WeightedString

func (a WeightedStrings) Len() int      { return len(a) }
func (a WeightedStrings) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a WeightedStrings) Less(i, j int) bool {
	// Sort lexicographically if the lengths are equal
	if len(a[i].String) == len(a[j].String) {
		return a[i].String < a[j].String
	}

	// Sort by length otherwise
	return len(a[i].String) < len(a[j].String)
}

type Fulfillment struct {
	Threshold       uint64
	SubFulfillments WeightedStrings
	SubConditions   WeightedStrings
}

func (wss WeightedStrings) MakeVarray() []byte {
	b := [][]byte{}

	for _, ws := range wss {
		b = append(b, encoding.MakeVarbyte(bytes.Join([][]byte{
			encoding.MakeUvarint(ws.Weight),
			encoding.MakeVarbyte([]byte(ws.String)),
		}, []byte{})))
	}

	return bytes.Join(b, []byte{})
}

func ParseWeightedStrings(b []byte) WeightedStrings {
	bs := encoding.ParseVarray(b)
	ws := WeightedStrings{}

	for _, b := range bs {
		w, b := encoding.GetUvarint(b)
		s, _ := encoding.GetVarbyte(b)

		ws = append(ws, WeightedString{
			Weight: w,
			String: string(s),
		})
	}

	return ws
}

func ParseFulfillments(b) []Fulfillment {
	fuls := []Fulfillment{}

	wss := ParseWeightedStrings(b)

	for _, ws := range wss {
		append()
	}
}

func ParseConditions(b) []Condition {

}

func (ful Fulfillment) Serialize(privkey []byte) string {
	sort.Sort(WeightedStrings(ful.SubFulfillments))
	sort.Sort(WeightedStrings(ful.SubConditions))
	payload := base64.URLEncoding.EncodeToString(bytes.Join([][]byte{
		encoding.MakeUvarint(ful.Threshold),
		encoding.MakeVarbyte(ful.SubConditions.MakeVarray()),
		encoding.MakeUvarint(ful.Threshold),
	}, []byte{}))

	return "cf:1:8:" + payload
}

// Parses Fulfillment out of the Crypto Conditions string format,
// and checks it for validity.
func ParseFulfillment(s string) (*Fulfillment, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 5 {
		return nil, errors.New("parsing error")
	}

	if parts[0] != "cf" {
		return nil, errors.New("fulfillments must start with \"cf\"")
	}

	if parts[1] != "1" {
		return nil, errors.New("must be protocol version 1")
	}

	if parts[2] != "4" {
		return nil, errors.New("not a sha256Threshold condition")
	}

	b, err := base64.URLEncoding.DecodeString(parts[3])
	if err != nil {
		return nil, errors.New("parsing error")
	}

	th, b := encoding.GetUvarint(b)

	fuls, b := encoding.GetVarbyte(b)

	conds, b := encoding.GetVarbyte(b)

	ful := &Fulfillment{
		Threshold:       threshold,
		SubFulfillments: ParseWeightedStrings(fulfillments),
		SubConditions:   ParseWeightedStrings(conditions),
	}

	// TODO: iterate through SubConditions and SubFulfillments, parse and verify

	return ful, nil
}

// // Turns an in-memory Fulfillment to an in-memory Condition. DynamicMessage and Signature
// // are discarded if present.
// func (ful Fulfillment) Condition() string {
// 	hash := sha256.Sum256(bytes.Join([][]byte{
// 		encoding.MakeVarbyte(ful.PublicKey),
// 		encoding.MakeVarbyte(ful.MessageId),
// 		encoding.MakeVarbyte(ful.FixedMessage),
// 	}, []byte{}))

// 	return "cc:1:8:" + base64.URLEncoding.EncodeToString(hash[:]) + ":" + strconv.FormatUint(ful.MaxFulfillmentLength, 10)
// }

// type Condition struct {
// 	PublicKey            []byte
// 	MessageId            []byte
// 	FixedMessage         []byte
// 	MaxFulfillmentLength uint64
// }

// // Serializes to the Crypto Conditions string format.
// func (cond Condition) Serialize() string {
// 	hash := sha256.Sum256(bytes.Join([][]byte{
// 		encoding.MakeVarbyte(cond.PublicKey),
// 		encoding.MakeVarbyte(cond.MessageId),
// 		encoding.MakeVarbyte(cond.FixedMessage),
// 	}, []byte{}))

// 	return "cc:1:8:" + base64.URLEncoding.EncodeToString(hash[:]) + ":" + strconv.FormatUint(cond.MaxFulfillmentLength, 10)
// }
