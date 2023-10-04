package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type PSchain struct {
	NodeID          string
	IFPSleader      int
	CurrentBlock    PSBlock
	MaybeNext       PSBlock
	Blocks          []PSBlock
	BlockQueue      chan PSBlock
	BlockQueue_send chan PSBlock
	VoteQueue1      chan Vote1
	VoteQueue2      chan Vote2
	VoteQueue3      chan Vote3
	VoteQueue_send1 chan Vote1
	VoteQueue_send2 chan Vote2
	VoteQueue_send3 chan Vote3
	//Vote4NextBlock       sync.Map
	Vote4NextBlock_vote1 sync.Map
	//Vote4NextBlock_vote2 sync.Map
	TxQueue            chan *Transaction
	Len                int
	StateChain         *StateChain
	Txpool             TransactionSlice
	Txpending          TransactionSlice
	Request_wait_proof RequestSlice
}

type Blockvotestatus struct {
	Num    int
	Nodeid []string
	Block  *PSBlock
}

func (bc *PSchain) AdPSBlock(b PSBlock) {
	bc.CurrentBlock = b
	bc.Blocks = append(bc.Blocks, b)
	bc.Len += 1
	PSchainlen += 1
}

func NewPSchain(node string, IFPSleader int) *PSchain {
	return &PSchain{
		node,
		IFPSleader,
		NewPSGenesisBlock(),
		NewPSGenesisBlock(),
		[]PSBlock{NewPSGenesisBlock()},
		make(chan PSBlock, 100),
		make(chan PSBlock, 100),
		make(chan Vote1, 100),
		make(chan Vote2, 100),
		make(chan Vote3, 100),
		make(chan Vote1, 100),
		make(chan Vote2, 100),
		make(chan Vote3, 100),
		sync.Map{},
		//sync.Map{},
		//sync.Map{},
		make(chan *Transaction, 100),
		0,
		NewStateChain(),
		NewEmptyTransactionSlice(),
		NewEmptyTransactionSlice(),
		NewEmptyRequestSlice(),
	}
}

func (bc *PSchain) Run(node *Node, cc *CSchain) {

	for {
		select {
		case tr := <-bc.TxQueue:
			if bc.Txpool.Exists(*tr) {
				continue
			}
			if bc.Txpending.Exists(*tr) {
				continue
			}
			if !tr.VerifyTransaction() {
				fmt.Println("Recieved non valid transaction", tr)
				continue
			}
			if tr.CheckCrossShard() {
				bc.Txpending.AddTransaction(*tr)
				temp_request := Newrequest(tr.Header.From, tr.Header.To, tr.Payload)
				temp_byte, _ := temp_request.MarshalBinary()
				temp_msg := NewMessage(node.ID, 3, temp_byte)

				ps, _ := tr.FindShardFrom()
				Nodeinps := getNodeInPS(nodes, ps, psnum)
				for i := 0; i < len(Nodeinps); i++ {
					sendMsgToIP(temp_msg, Nodeinps[i].Addr)
				}

			}
			bc.Txpool.AddTransaction(*tr)
		case temp_psblock := <-bc.BlockQueue:
			fmt.Println(bc.NodeID, "Now in:", (bc.Len)+1, "epoch")
			if temp_psblock.Epoch > int16(bc.Len) {
				if _, ok := bc.Vote4NextBlock_vote1.Load(temp_psblock.BlockHeader.Hash); ok {
				} else {
					var temp []string
					bc.Vote4NextBlock_vote1.Store(temp_psblock.BlockHeader.Hash, Blockvotestatus{1, temp, &temp_psblock})
					//bc.Vote4NextBlock_vote2.Store(temp_psblock.BlockHeader.Hash, Blockvotestatus{1, temp, &temp_psblock})
					//bc.Vote4NextBlock.Store(temp_psblock.BlockHeader.Hash, Blockvotestatus{1, temp, &temp_psblock})
					fmt.Println(bc.NodeID, "-----: Get block", temp_psblock.Epoch)
					node.timerExpired1 = false
					bc.MaybeNext = temp_psblock
					node.globalTimer1 = time.AfterFunc(time.Duration(T1)*time.Second, func() {
						node.timerExpired1 = true
						node.timerExpired2 = false
						if node.addps {
							//fmt.Println(bc.NodeID, "'s epoch", temp_psblock.Epoch, "add process block in time-bound-1 successfully!")
						} else {
							bc.GenerateBlock(node)
							//fmt.Println(node.ID, "再次尝试出块,此时vote summary是", node.VoteSummary)
							if node.addps == false {
								//fmt.Println(bc.NodeID, "'s epoch", temp_psblock.Epoch, "failed to add block in time-bound-1.")
								bc.BlockQueue_send <- temp_psblock
							}
						}
					})
					//fmt.Println(bc.NodeID, "Turn on time-bound-1 countdown")
					if temp_psblock.Origin != bc.NodeID && !contains(node.Advepoch, int(temp_psblock.Epoch)) {
						temp_txslice := new(TransactionSlice)
						err := temp_txslice.UnmarshalBinary(temp_psblock.Data)
						if err != nil {
							fmt.Println(err)
						}
						for i := 0; i < temp_txslice.Len(); i++ {
							data := *temp_txslice // Dereference the pointer to access the TransactionSlice value
							temp_Transaction := data[i]
							if temp_Transaction.VerifyTransaction() == false {
								//generate false vote
							}
						}
						newvote := Vote1{"", temp_psblock.Epoch, temp_psblock.PSindex, true, temp_psblock.BlockHeader.Hash, [32]byte{}}
						bc.VoteQueue_send1 <- newvote
					}
				}
			}
		case temp_vote := <-bc.VoteQueue1:
			if node.CSLeader == 1 && node.ifvotesummary == false && node.VotesummaryRound == bc.Len {
				node.ifvotesummary = true
				node.VoteSummary = NewVoteSummary(node.Addr, int16(bc.Len)+1, node.CSID)
				SendAfter(delta, node, bc, cc)
			}
			if temp_vote.Epoch == int16(bc.Len)+1 {
				if nowvoteRaw, ok := bc.Vote4NextBlock_vote1.Load(temp_vote.BlockHash); ok {
					nowvote, ok := nowvoteRaw.(Blockvotestatus)
					if !ok {
						// If the type assertion fails, the error is handled
						log.Printf("Error: Vote4NextBlock value of key %v has wrong type\n", temp_vote.BlockHash)
						continue
					}
					if stringInSlice(temp_vote.Origin, nowvote.Nodeid) {
						// Origin already exists in the Nodeid of the Block
					} else {
						nowvote.Nodeid = append(nowvote.Nodeid, temp_vote.Origin)
						bc.Vote4NextBlock_vote1.Store(temp_vote.BlockHash, Blockvotestatus{nowvote.Num + 1, nowvote.Nodeid, nowvote.Block})
						//fmt.Println(bc.NodeID, "--------:recevied vote_for_process_block from", temp_vote.Origin)
						//bc.Check_vote1(node, temp_vote)
						// if node.CSLeader == 1 && node.ifvotesummary == true && node.VotesummaryRound == bc.Len {
						// 	if nowvotesummary, ok := node.VoteSummary.AgreeMap.Load(temp_vote.PSindex); ok {
						// 		node.VoteSummary.AgreeMap.Store(temp_vote.PSindex, Votelist{nowvotesummary.Num + 1, append(nowvotesummary.Nodelist, temp_vote.Origin)})
						// 	} else {
						// 		node.VoteSummary.AgreeMap.Store(temp_vote.PSindex, Votelist{nowvotesummary.Num + 1, append(nowvotesummary.Nodelist, temp_vote.Origin)})
						// 	}
						// }
					}
				} else {
					//fmt.Println(bc.NodeID, "Stuck here1, please restart (server concurrency is poor),now", bc.Len, "receive", temp_vote.Epoch)
					go bc.insertvote_delay(temp_vote)
				}
			} else if temp_vote.Epoch > int16(bc.Len)+1 {
				//fmt.Println(bc.NodeID, "receive", temp_vote.Origin, "epoch", temp_vote.Epoch, "'s vote")
				go bc.insertvote_delay(temp_vote)
			} else {
			}
		}
	}
}

// func (bc *PSchain) Check_vote1(node *Node, temp_vote Vote1) {
// 	bc.Vote4NextBlock_vote1.Range(func(key, value interface{}) bool {
// 		blockvotestatus, ok := value.(Blockvotestatus)
// 		if !ok {
// 			return false
// 		}
// 		if blockvotestatus.Num > (psnum-2) && blockvotestatus.Block.Epoch == int16(bc.Len)+1 {
// 			newvote2 := Vote2{"", temp_vote.Epoch, temp_vote.PSindex, true, temp_vote.BlockHash, [32]byte{}}
// 			bc.VoteQueue_send2 <- newvote2
// 		}
// 		return true
// 	})
// }

func (bc *PSchain) GenerateBlock(node *Node) {
	node.VoteSummary.AgreeMap = VoteSummary_thisepoch.AgreeMap
	//fmt.Println(node.ID, "'s vote summary:", node.VoteSummary.AgreeMap)
	if Vote, ok := node.VoteSummary.AgreeMap.Load(node.PSID); ok {
		if Vote.Num > (psnum - 2) {
			bc.Blocks = append(bc.Blocks, bc.MaybeNext)
			bc.CurrentBlock = bc.MaybeNext
			bc.Len += 1
			PSchainlen += 1
			bc.Vote4NextBlock_vote1.Delete(bc.CurrentBlock.BlockHeader.Hash)
			node.addps = true
			fmt.Println(bc.NodeID, "--------:Add Block", bc.Len)
			err := os.MkdirAll("./data/"+bc.NodeID+"/psblock", os.ModePerm)
			err = os.MkdirAll("./data/"+bc.NodeID+"/state", os.ModePerm)
			if err != nil {
			}
			tempstate := NewState(bc.CurrentBlock.Epoch, bc.CurrentBlock.Hash)
			bc.StateChain.Add(tempstate)
			SavePSBlockToFile(bc.CurrentBlock, "./data/"+bc.NodeID+"/psblock/"+strconv.Itoa(bc.Len))
			SaveStateToFile(bc.StateChain.LastState(), "./data/"+bc.NodeID+"/state/"+strconv.Itoa(bc.Len))
		}
		return
	}
}

func (bc *PSchain) insertvote_delay(temp_vote Vote1) {
	time.Sleep(1 * time.Second)
	bc.VoteQueue1 <- temp_vote
}
func (bc *PSchain) insertblock_delay(temp_block PSBlock) {
	time.Sleep(1 * time.Second)
	bc.BlockQueue <- temp_block
}
