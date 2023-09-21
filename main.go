package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

const num = 6   // number of nodes totally
const psnum = 6 // the shard size of process shard
const csnum = 6 // the shard size of control shard
var adv = [2]int{0, 0}
var T1 = 15 // T1:time-bound-1
// lamda
const lamda = 35 // T2 = (pontrol shard number - suceess pontrol shard)*(lamda)
const delta = 10

//const tau = 41

var random = false // if shuffle the node by drand number
var Start time.Time
var keypair *Keypair

const psincs = csnum / psnum

var nodes []Node

func main() {
	// Generate the key pair
	keypair, _ = GenerateECDSAKeyPair("private_key.pem", "public_key.pem")
	// Gernerate the nodes
	node, _ := generateNodes(num)
	// TODO: Replace it with Practical asynchronous distributed key generation

	getDrandResult()
	readRandomnessFromFile("./data/drand.json")

	advlist := generateadvepcoh(adv[:], num, psnum, csnum)
	localIP, err := GetLocalIP()
	if err != nil {
		fmt.Println("err when get Get local ip")
	}
	publicIP, err := GetPublicIP()
	if err != nil {
		fmt.Println("err when get Get public ip")
	}
	for _, tempnode := range node {
		if IsPrefix(tempnode.Addr, localIP) || IsPrefix(tempnode.Addr, publicIP) || IsPrefix(tempnode.Addr, "localhost") || IsPrefix(tempnode.Addr, "127.0.0.1") {
			tempnode.bootstrap(random)
			for epoch, epochadvlist := range advlist {
				nodeid, _ := strconv.Atoi(tempnode.ID[4:])
				if contains(epochadvlist, nodeid) {
					tempnode.Advepoch = append(tempnode.Advepoch, epoch+1)
				}
			}
		}
	}

	for _, tempnode := range node {
		if IsPrefix(tempnode.Addr, localIP) || IsPrefix(tempnode.Addr, publicIP) || IsPrefix(tempnode.Addr, "localhost") || IsPrefix(tempnode.Addr, "127.0.0.1") {
			tempnode.bootstrap(random)
			go startNode(tempnode)
		}
	}

	for {
		time.Sleep(2 * time.Second)
	}

}

func (bc *PSchain) AssignMessage(msg Message, psid int, cc *CSchain, node Node) {
	//1:PSblock 2:Vote1 3：request 4:proof 5:Vote2 6: Vote3 7:Vote summary 8：CSblock
	switch msg.Type {

	//case 1 代表收到的是 process shard block
	case 1:
		temp_block, err := DecodePSBlock(msg.Data)
		if err != nil {
			log.Printf("Error decoding PSBlock: %s", err.Error())
			return
		}

		if temp_block == nil {
			log.Printf("Decoded PSBlock is nil")
			return
		}

		if temp_block.PSindex == psid {
			bc.BlockQueue <- *temp_block
		} else {
			cc.BlockQueue <- *temp_block
		}

	//case 2 代表收到的是对区块的投票信息
	case 2:
		temp_vote, err := DecodeVote1(msg.Data)
		if err != nil {
			log.Printf("Error decoding Vote: %s", err.Error())
			return
		}

		if temp_vote == nil {
			log.Printf("Decoded Vote is nil")
			return
		}

		if temp_vote.PSindex == psid && stringInSlice(temp_vote.Origin, node.PSPeerID) {
			bc.VoteQueue1 <- *temp_vote
		} else {
			cc.VoteQueue1 <- *temp_vote
		}

	// case 3 代表的是收到cross shard tx查询的request
	case 3:
		temp_request := Newrequest_empty()
		_, err := temp_request.UnmarshalBinary(msg.Data)
		if err != nil {
			log.Printf("Error decoding Request: %s", err.Error())
			return
		}
		if temp_request == nil {
			log.Printf("Decoded Request is nil")
			return
		}

		temp_tx := NewTransaction(temp_request.Header.From, temp_request.Header.To, temp_request.Payload)
		if (!bc.Txpool.Exists(*temp_tx)) && temp_tx.VerifyTransaction() {
			bc.Txpool.AddTransaction(*temp_tx)
			bc.Request_wait_proof.AddRequest(*temp_request)
		}

	//case 4 代表收到的收到证明cross-shard tx 的 proof
	case 4:
		temp_proof := Newproof_empty()
		_, err := temp_proof.UnmarshalBinary(msg.Data)
		if err != nil {
			log.Printf("Error decoding Proof: %s", err.Error())
			return
		}

		if temp_proof == nil {
			log.Printf("Decoded Proof is nil")
			return
		}
		temp_tx := NewTransaction(temp_proof.Header.From, temp_proof.Header.To, temp_proof.Payload)
		if bc.Txpending.Exists(*temp_tx) {
			if (!bc.Txpool.Exists(*temp_tx)) && temp_tx.VerifyTransaction() {
				bc.Txpool.AddTransaction(*temp_tx)
				bc.Txpending.DeleteTransaction(*temp_tx)
			}
		}
	// case 5 是 Byzantine agreement的第一轮投票
	case 5:
		temp_vote2, err := DecodeVote2(msg.Data)
		if err != nil {
			log.Printf("Error decoding Vote: %s", err.Error())
			return
		}

		if temp_vote2 == nil {
			log.Printf("Decoded Vote is nil")
			return
		}

		if temp_vote2.CSindex == node.CSID {
			cc.VoteQueue2 <- *temp_vote2
		}
	// case 6 是 Byzantine agreement的第二轮投票
	case 6:
		temp_vote3, err := DecodeVote3(msg.Data)
		if err != nil {
			log.Printf("Error decoding Vote: %s", err.Error())
			return
		}

		if temp_vote3 == nil {
			log.Printf("Decoded Vote is nil")
			return
		}

		if temp_vote3.CSindex == node.CSID {
			cc.VoteQueue3 <- *temp_vote3
		}
	// case 7 是 leader 收集提出的要进行Byzantine agreement的投票信息汇总
	case 7:
		temp_votesummary, err := DecodeVoteSummary(msg.Data)
		if err != nil {
			log.Printf("Error decoding Vote summary: %s", err.Error())
			return
		}
		if temp_votesummary == nil {
			log.Printf("Decoded Votesummary is nil")
			return
		}
		cc.VoteSummaryQueue <- *temp_votesummary

	}
}
