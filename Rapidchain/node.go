package main

import (
	"log"
	"time"
)

var VoteSummary_thisepoch VoteSummary

type Node struct {
	ID                string
	Addr              string
	PSID              int
	CSID              int
	PSchain           *PSchain
	CSchain           *CSchain
	PSPeerID          []string
	CSPeerID          []string
	PSPeerAddr        []string
	CSPeerAddr        []string
	PSLeader          int
	CSLeader          int
	globalTimer2      *time.Timer
	timerExpired2     bool
	globalTimer1      *time.Timer
	timerExpired1     bool
	addps             bool
	T2                int
	Advepoch          []int
	ifvotesummary     bool
	VotesummaryRound  int
	VoteSummary       VoteSummary
	GetVotesummary    bool
	ifsendvote3       bool
	ifsavevotesummary bool
}

func startNode(node *Node) {
	n := SetupNetwork(*node)
	node.PSchain = NewPSchain(node.ID, node.PSLeader)
	node.CSchain = NewCSchain(node.ID)
	node.ifvotesummary = false
	node.ifsendvote3 = false
	node.ifsavevotesummary = false
	go n.Run()
	go node.PSchain.Run(node, node.CSchain)
	go node.CSchain.Run(node.PSchain, node)
	time.Sleep(2 * time.Second)
	if node.PSLeader == 1 {
		go producePSBlocks(node, n, node.PSchain)
	}
	for {
		select {
		//处理接收到的消息
		case message := <-n.IncomingMessages:
			// Processing incoming messages
			//log.Printf("%s Received message from: %+v", node.ID, message.SenderID)
			node.PSchain.AssignMessage(message, node.PSID, node.CSchain, *node)

		//发送不同类型的消息
		case newvote := <-node.PSchain.VoteQueue_send1:
			// 发送vote, 这里是对process shard的投票
			newvote.Origin = node.ID
			new_vote_byte, err := EncodeVote1(&newvote)
			if err != nil {
				log.Printf("error happened vote to bytes")
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 2, new_vote_byte)
			n.CSBroadcastMessages <- NewMessage(node.ID, 2, new_vote_byte)
		case newvote2 := <-node.PSchain.VoteQueue_send2:
			// Send vote
			newvote2.Origin = node.ID
			new_vote_byte, err := EncodeVote2(&newvote2)
			if err != nil {
				log.Printf("error happened vote to bytes")
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 5, new_vote_byte)
			n.CSBroadcastMessages <- NewMessage(node.ID, 5, new_vote_byte)
		case newvote3 := <-node.PSchain.VoteQueue_send3:
			// Send vote
			newvote3.Origin = node.ID
			new_vote_byte, err := EncodeVote3(&newvote3)
			if err != nil {
				log.Printf("error happened vote to bytes")
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 6, new_vote_byte)
			n.CSBroadcastMessages <- NewMessage(node.ID, 6, new_vote_byte)
		case blocktocs := <-node.PSchain.BlockQueue_send:
			if node.PSLeader == 1 {
				blocktocs_byte, err := EncodePSBlock(&blocktocs)
				if err != nil {
					log.Printf("error happened blocktocs to bytes")
				}
				n.CSBroadcastMessages <- NewMessage(node.ID, 1, blocktocs_byte)
			}
		case newvote := <-node.CSchain.VoteQueue_send1:
			newvote.Origin = node.ID
			new_vote_byte, err := EncodeVote1(&newvote)
			if err != nil {
				log.Printf("error happened vote to bytes")
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 2, new_vote_byte)
			n.CSBroadcastMessages <- NewMessage(node.ID, 2, new_vote_byte)
		case newvote2 := <-node.CSchain.VoteQueue_send2:
			newvote2.Origin = node.ID
			new_vote_byte2, err := EncodeVote2(&newvote2)
			if err != nil {
				log.Printf("error happened vote to bytes")
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 5, new_vote_byte2)
			n.CSBroadcastMessages <- NewMessage(node.ID, 5, new_vote_byte2)
		case NewVote3 := <-node.CSchain.VoteQueue_send3:
			NewVote3.Origin = node.ID
			new_vote_byte3, err := EncodeVote3(&NewVote3)
			if err != nil {
				log.Printf("error happened vote to bytes")
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 6, new_vote_byte3)
			n.CSBroadcastMessages <- NewMessage(node.ID, 6, new_vote_byte3)
		case newVoteSummary := <-node.CSchain.VoteSummaryQueue_send:
			newVoteSummary.Origin = node.ID
			newVoteSummary_byte, err := EncodeVoteSummary(&newVoteSummary)
			if err != nil {
				log.Printf("error happened vote summary to bytes", err)
			}
			n.PSBroadcastMessages <- NewMessage(node.ID, 7, newVoteSummary_byte)
			n.CSBroadcastMessages <- NewMessage(node.ID, 7, newVoteSummary_byte)
		}

	}
}

func producePSBlocks(node *Node, n *Network, bc *PSchain) {
	sentBlockepoch := 0
	// Generate the first block
	time.Sleep(20 * time.Second)
	//read transaction from ps.Txpool and delete them from ps.Txpool
	temptxpool := NewEmptyTransactionSlice()
	for i := 0; i < 100; i++ { //生成一百条交易
		temptxpool = append(temptxpool, *CreateTransaction(ShuffleBytes(keypair.Public), GenerateRandomNumber())) // 将空的Transaction结构体追加到TransactionSlice中
	}
	// 将每个block填充一百条交易
	tempBlock := NewPSBlock(1, node.ID, node.PSID, [32]byte{}, &temptxpool, []byte("node"))
	//查看被包含的交易是否是被其他shard require的
	//TODO 这个步骤应该后置
	for i := 0; i < len(temptxpool); i++ {
		if bc.Request_wait_proof.Exists(temptxpool[i]) {
			temp_proof := Newproof(bc.Txpool[i].Header.From, bc.Txpool[i].Header.To, bc.Txpool[i].Payload)
			bc.Request_wait_proof.DeleteRequest(bc.Txpool[i])
			temp_byte, _ := temp_proof.MarshalBinary()
			temp_msg := NewMessage(bc.NodeID, 4, temp_byte)
			ps, _ := temptxpool[i].FindShardTo()
			Nodeinps := getNodeInPS(nodes, ps, psnum)
			for i := 0; i < len(Nodeinps); i++ {
				sendMsgToIP(temp_msg, Nodeinps[i].Addr)
			}
			bc.Txpool.DeleteTransaction(bc.Txpool[i])
		}
	}
	if &tempBlock == nil {
		log.Printf("tempblock is nil")
	}

	tempBlockByte, err := EncodePSBlock(&tempBlock)

	if err != nil {
		log.Printf(err.Error())
		log.Printf("error happened block to bytes")
	}
	nodeMsg := NewMessage(node.ID, 1, tempBlockByte)
	Start = time.Now()
	// 把区块发送给ps中的全部节点
	n.PSBroadcastMessages <- nodeMsg
	// 自己也需要处理这个区块
	node.PSchain.AssignMessage(nodeMsg, node.PSID, node.CSchain, *node)

	for {
		if (PSchainlen / csnum) > sentBlockepoch {
			if (CSchainlen / csnum) > sentBlockepoch { //当前len 从0开始，一旦len大于sent block epoch意味着上一个block已经被大家接受
				if node.timerExpired1 && node.timerExpired2 {
					//两个timerExpired 说明上一个epoch结束
					node.timerExpired1 = false
					node.timerExpired2 = false
					// keep generate the block
					temptxpool := NewEmptyTransactionSlice()
					for i := 0; i < 100; i++ {
						temptxpool = append(temptxpool, *CreateTransaction(ShuffleBytes(keypair.Public), GenerateRandomNumber()))
					}
					tempBlock := NewPSBlock(int16(node.PSchain.Len)+1, node.ID, node.PSID, node.PSchain.CurrentBlock.BlockHeader.Hash, &temptxpool, []byte("node1"))
					tempBlockByte, err := EncodePSBlock(&tempBlock)
					if err != nil {
						log.Printf("error happened block to bytes")
					}
					for i := 0; i < len(temptxpool); i++ {
						if bc.Request_wait_proof.Exists(temptxpool[i]) {
							temp_proof := Newproof(bc.Txpool[i].Header.From, bc.Txpool[i].Header.To, bc.Txpool[i].Payload)
							bc.Request_wait_proof.DeleteRequest(bc.Txpool[i])
							temp_byte, _ := temp_proof.MarshalBinary()
							temp_msg := NewMessage(bc.NodeID, 4, temp_byte)
							ps, _ := temptxpool[i].FindShardTo()
							Nodeinps := getNodeInPS(nodes, ps, psnum)
							for i := 0; i < len(Nodeinps); i++ {
								sendMsgToIP(temp_msg, Nodeinps[i].Addr)
							}
							bc.Txpool.DeleteTransaction(bc.Txpool[i])
						}
					}
					nodeMsg := NewMessage(node.ID, 1, tempBlockByte)
					Start = time.Now()
					n.PSBroadcastMessages <- nodeMsg
					node.PSchain.AssignMessage(nodeMsg, node.PSID, node.CSchain, *node)
					sentBlockepoch++
				}
			} else {
				//fmt.Println("当前CSchainlen", CSchainlen)
				time.Sleep(1 * time.Second)
			}
		} else {
			//fmt.Println("当前PSchainlen", PSchainlen)
			time.Sleep(1 * time.Second)
		}
	}
}
