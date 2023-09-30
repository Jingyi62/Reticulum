package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"time"
)

type Request struct {
	Header    RequestHeader
	Signature []byte
	Payload   []byte
}

type RequestHeader struct {
	From          []byte
	To            []byte
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

type Proof struct {
	Header    ProofHeader
	Signature []byte
	Payload   []byte
}

type ProofHeader struct {
	From          []byte
	To            []byte
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}
type RequestSlice []Request

func Newrequest(from, to, payload []byte) *Request {

	t := Request{Header: RequestHeader{From: from, To: to}, Payload: payload}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}
func Newrequest_empty() *Request {

	t := Request{Header: RequestHeader{}}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}
func Newproof_empty() *Proof {

	t := Proof{Header: ProofHeader{}}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}
func Newproof(from, to, payload []byte) *Proof {

	t := Proof{Header: ProofHeader{From: from, To: to}, Payload: payload}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}
func (t *Request) MarshalBinary() ([]byte, error) {

	headerBytes, _ := t.Header.MarshalBinary()

	return append(append(headerBytes, FitBytesInto(t.Signature, 80)...), t.Payload...), nil
}
func (t *Request) UnmarshalBinary(d []byte) ([]byte, error) {

	buf := bytes.NewBuffer(d)

	if len(d) < 80+80 {
		return nil, errors.New("Insuficient bytes for unmarshalling transaction")
	}

	header := &RequestHeader{}
	if err := header.UnmarshalBinary(buf.Next(80)); err != nil {
		return nil, err
	}

	t.Header = *header

	t.Signature = StripByte(buf.Next(80), 0)
	t.Payload = buf.Next(int(t.Header.PayloadLength))

	return buf.Next(MaxInt), nil
}

func (th *RequestHeader) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.Write(FitBytesInto(th.From, 80))
	buf.Write(FitBytesInto(th.To, 80))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil

}
func (th *RequestHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	th.From = StripByte(buf.Next(80), 0)
	th.To = StripByte(buf.Next(80), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}

func (t *Proof) MarshalBinary() ([]byte, error) {

	headerBytes, _ := t.Header.MarshalBinary()

	return append(append(headerBytes, FitBytesInto(t.Signature, 80)...), t.Payload...), nil
}
func (t *Proof) UnmarshalBinary(d []byte) ([]byte, error) {

	buf := bytes.NewBuffer(d)

	if len(d) < 80+80 {
		return nil, errors.New("Insuficient bytes for unmarshalling transaction")
	}

	header := &ProofHeader{}
	if err := header.UnmarshalBinary(buf.Next(80)); err != nil {
		return nil, err
	}

	t.Header = *header

	t.Signature = StripByte(buf.Next(80), 0)
	t.Payload = buf.Next(int(t.Header.PayloadLength))

	return buf.Next(MaxInt), nil
	
}

func (th *ProofHeader) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.Write(FitBytesInto(th.From, 80))
	buf.Write(FitBytesInto(th.To, 80))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil

}
func (th *ProofHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	th.From = StripByte(buf.Next(80), 0)
	th.To = StripByte(buf.Next(80), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}

// send request to which shard
func (t *Transaction) FindShardFrom() (ps int, cs int) {
	// cross-shard assign algorithm
	return ps, cs
}

// send proof to which shard
func (t *Transaction) FindShardTo() (ps int, cs int) {
	// cross-shard assign algorithm
	return ps, cs
}

func NewEmptyRequestSlice() RequestSlice {
	return make(RequestSlice, 0)
}
func (slice RequestSlice) AddRequest(t Request) RequestSlice {

	// Inserted sorted by timestamp
	for i, tr := range slice {
		if tr.Header.Timestamp >= t.Header.Timestamp {
			return append(append(slice[:i], t), slice[i:]...)
		}
	}

	return append(slice, t)
}
func (slice RequestSlice) Exists(tr Transaction) bool {

	for _, t := range slice {
		if reflect.DeepEqual(t.Signature, tr.Signature) {
			return true
		}
	}
	return false
}

func (slice RequestSlice) DeleteRequest(t Transaction) RequestSlice {
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

	newSlice := make(RequestSlice, 0, len(slice)-1)
	newSlice = append(newSlice, slice[:index]...)
	newSlice = append(newSlice, slice[index+1:]...)

	return newSlice
}
