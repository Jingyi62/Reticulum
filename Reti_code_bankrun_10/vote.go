package main

import (
	"bytes"
	"encoding/gob"
)

/*
Vote1: process nodes 对 process block 进行的投票，发送给control nodes
Vote2: control nodes 对 process block 投票结果的 第一次投票（Byzantine Agreement的第二阶段，prepare）
Vote3: control nodes 对 process block 投票结果的 第一次投票的确认收到的投票（Byzantine Agreement的第三阶段，commit)
Vote Summary： process shard 对 process block 投票结果的summary
*/
type Vote1 struct {
	Origin    string
	Epoch     int16
	PSindex   int
	Agree     bool
	BlockHash [32]byte
	Signature [32]byte
}

type KeyValuePair struct {
	PSindex int
	Value   Votelist
}
type Votelist struct {
	Num      int
	Nodelist []string
}

type SyncMapWrapper struct {
	Data []KeyValuePair
}
type VoteSummary struct {
	Origin    string
	CSindex   int
	Epoch     int16
	AgreeMap  SyncMapWrapper
	Signature [32]byte
}
type VoteSummaryMap struct {
	Num    int
	Nodeid []string
}

type Vote2 struct {
	Origin          string
	Epoch           int16
	CSindex         int
	Agree           bool
	VoteSummaryHash [32]byte
	Signature       [32]byte
}

type Vote3 struct {
	Origin          string
	Epoch           int16
	CSindex         int
	Generateblock   bool
	VoteSummaryHash [32]byte
	Signature       [32]byte
}

func (s *SyncMapWrapper) Load(PS int) (Votelist, bool) {
	for _, kv := range s.Data {
		if kv.PSindex == PS {
			return kv.Value, true
		}
	}
	return Votelist{}, false
}

func (s *SyncMapWrapper) Store(PS int, value Votelist) {
	for i, kv := range s.Data {
		if kv.PSindex == PS {
			s.Data[i].Value = value
			return
		}
	}
	s.Data = append(s.Data, KeyValuePair{PSindex: PS, Value: value})
}

func NewVote1(Org string, epoch int16, PSindex int, agree bool, BlockSign [32]byte) Vote1 {
	var sign [32]byte
	return Vote1{Org, epoch, PSindex, agree, BlockSign, sign}
}

func NewVoteSummary(Org string, epoch int16, Csindex int) VoteSummary {
	var sign [32]byte
	return VoteSummary{Org, Csindex, epoch, SyncMapWrapper{}, sign}
}

func EncodeVoteSummary(v *VoteSummary) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeVoteSummary(data []byte) (*VoteSummary, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var v VoteSummary
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func EncodeVote1(v *Vote1) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeVote1(data []byte) (*Vote1, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var v Vote1
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func NewVote2(Org string, epoch int16, PSindex int, agree bool, BlockSign [32]byte) Vote2 {
	var sign [32]byte
	return Vote2{Org, epoch, PSindex, agree, BlockSign, sign}
}

func EncodeVote2(v *Vote2) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeVote2(data []byte) (*Vote2, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var v Vote2
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func NewVote3(Org string, epoch int16, PSindex int, agree bool, BlockSign [32]byte) Vote3 {
	var sign [32]byte
	return Vote3{Org, epoch, PSindex, agree, BlockSign, sign}
}

func EncodeVote3(v *Vote3) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeVote3(data []byte) (*Vote3, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var v Vote3
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
