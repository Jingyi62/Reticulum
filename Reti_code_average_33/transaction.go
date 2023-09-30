package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"time"
)

type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
}

type TransactionHeader struct {
	From          []byte
	To            []byte
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

func NewTransaction(from, to, payload []byte) *Transaction {

	t := Transaction{Header: TransactionHeader{From: from, To: to}, Payload: payload}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}

func (t *Transaction) Hash() []byte {

	headerBytes, _ := t.Header.MarshalBinary()
	return SHA256(headerBytes)
}

func (t *Transaction) Sign(keypair *Keypair) []byte {

	s, _ := keypair.Sign(t.Hash())

	return s
}

func (t *Transaction) VerifyTransaction() bool {

	headerHash := t.Hash()
	payloadHash := SHA256(t.Payload)

	return reflect.DeepEqual(payloadHash, t.Header.PayloadHash) && SignatureVerify(t.Header.From, t.Signature, headerHash)
}
func (t *Transaction) CheckCrossShard() bool {
	// ToDo: cross-shard rule
	if t.Header.From == nil {
		return true
	}
	return false
}
func (t *Transaction) MarshalBinary() ([]byte, error) {

	headerBytes, _ := t.Header.MarshalBinary()

	return append(append(headerBytes, FitBytesInto(t.Signature, 80)...), t.Payload...), nil
}
func (t *Transaction) UnmarshalBinary(d []byte) ([]byte, error) {

	buf := bytes.NewBuffer(d)

	if len(d) < 80+80 {
		return nil, errors.New("Insuficient bytes for unmarshalling transaction")
	}

	header := &TransactionHeader{}
	if err := header.UnmarshalBinary(buf.Next(80)); err != nil {
		return nil, err
	}

	t.Header = *header

	t.Signature = StripByte(buf.Next(80), 0)
	t.Payload = buf.Next(int(t.Header.PayloadLength))

	return buf.Next(MaxInt), nil

}

func (th *TransactionHeader) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.Write(FitBytesInto(th.From, 80))
	buf.Write(FitBytesInto(th.To, 80))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil

}
func (th *TransactionHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	th.From = StripByte(buf.Next(80), 0)
	th.To = StripByte(buf.Next(80), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}

type TransactionSlice []Transaction

func NewEmptyTransactionSlice() TransactionSlice {
	return make(TransactionSlice, 0)
}

func (slice TransactionSlice) Len() int {

	return len(slice)
}

func (slice TransactionSlice) Exists(tr Transaction) bool {

	for _, t := range slice {
		if reflect.DeepEqual(t.Signature, tr.Signature) {
			return true
		}
	}
	return false
}
func (slice TransactionSlice) DeleteTransaction(t Transaction) TransactionSlice {
	index := -1

	for i, tr := range slice {
		if reflect.DeepEqual(t.Signature, tr.Signature) {
			index = i
			break
		}
	}

	if index == -1 {
		return slice
	}

	newSlice := make(TransactionSlice, 0, len(slice)-1)
	newSlice = append(newSlice, slice[:index]...)
	newSlice = append(newSlice, slice[index+1:]...)

	return newSlice
}

func (slice TransactionSlice) AddTransaction(t Transaction) TransactionSlice {

	// Inserted sorted by timestamp
	for i, tr := range slice {
		if tr.Header.Timestamp >= t.Header.Timestamp {
			return append(append(slice[:i], t), slice[i:]...)
		}
	}

	return append(slice, t)
}

func (slice *TransactionSlice) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	for _, t := range *slice {

		bs, err := t.MarshalBinary()

		if err != nil {
			return nil, err
		}

		buf.Write(bs)
	}

	return buf.Bytes(), nil
}

func (slice *TransactionSlice) UnmarshalBinary(d []byte) error {

	remaining := d

	for len(remaining) > 80+80 {
		t := new(Transaction)
		rem, err := t.UnmarshalBinary(remaining)

		if err != nil {
			return err
		}
		(*slice) = append((*slice), *t)
		remaining = rem
	}
	return nil
}
func CreateTransaction(to []byte, txt string) *Transaction {

	t := NewTransaction(keypair.Public, to, []byte(txt))
	t.Signature = t.Sign(keypair)

	return t
}
