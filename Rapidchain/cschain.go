package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type CSchain struct {
	NodeID                 string
	CurrentBlock           CSBlock
	CurrentVoteSummary     VoteSummary
	Blocks                 []CSBlock
	VoteSummarys           []VoteSummary
	BlockQueue             chan PSBlock
	Votetable              map[int][]string
	VoteQueue1             chan Vote1
	VoteQueue2             chan Vote2
	VoteQueue3             chan Vote3
	VoteSummaryQueue       chan VoteSummary
	VoteQueue_send1        chan Vote1
	VoteQueue_send2        chan Vote2
	VoteQueue_send3        chan Vote3
	Vote4votesummary_vote2 sync.Map
	Vote4votesummary_vote3 sync.Map
	VoteSummaryQueue_send  chan VoteSummary
	Len                    int
}

type votesummarystatus struct {
	Num         int
	Nodeid      []string
	Votesummary *VoteSummary
}

func NewCSchain(node string) *CSchain {
	return &CSchain{
		node,
		NewCSGenesisBlock(),
		NewVoteSummary("", 0, 0),
		[]CSBlock{NewCSGenesisBlock()},
		[]VoteSummary{NewVoteSummary("", 0, 0)},
		make(chan PSBlock, 100),
		map[int][]string{},
		make(chan Vote1, 100),
		make(chan Vote2, 100),
		make(chan Vote3, 100),
		make(chan VoteSummary, 50),
		make(chan Vote1, 100),
		make(chan Vote2, 100),
		make(chan Vote3, 100),
		sync.Map{},
		sync.Map{},
		make(chan VoteSummary, 50),
		0,
	}
}

func SetTime2(cc *CSchain, bc *PSchain, node *Node) {
	for {
		if (node.timerExpired1 == true) && node.timerExpired2 == false {
			count := 0

			node.VoteSummary = VoteSummary_thisepoch
			for i := 0; i < psincs; i++ {
				if node.VoteSummary.AgreeMap.Data[i].Value.Num >= (psnum) {
					count++
				}
			}

			node.T2 = (psincs - count + 1) * lamda
			//fmt.Println(node.ID, "Now begin the T2", node.T2)
			Endepoch(cc, bc, node)

		}
	}

}

func Endepoch(cc *CSchain, bc *PSchain, node *Node) {
	done := make(chan bool)
	node.globalTimer2 = time.AfterFunc(time.Duration(node.T2)*time.Second, func() {
		node.timerExpired2 = true
		if node.addps == false {
			vote := cc.Votetable[node.PSID]
			if len(vote) > (csnum/2)-2 {
				bc.Blocks = append(bc.Blocks, bc.MaybeNext)
				bc.Len += 1
				PSchainlen += 1
				//fmt.Println(bc.NodeID, "--------after control shard's vote : Add Block", bc.Len)
				err := os.MkdirAll("./data/"+bc.NodeID+"/psblock", os.ModePerm)
				err = os.MkdirAll("./data/"+bc.NodeID+"/state", os.ModePerm)
				if err != nil {
				}
				tempstate := NewState(bc.CurrentBlock.Epoch, bc.CurrentBlock.Hash)
				bc.StateChain.Add(tempstate)
				SavePSBlockToFile(bc.CurrentBlock, "./data/"+bc.NodeID+"/psblock/"+strconv.Itoa(bc.Len))
				SaveStateToFile(bc.StateChain.LastState(), "./data/"+bc.NodeID+"/state/"+strconv.Itoa(bc.Len))

			}
		}
		done <- true

	})
	<-done
	node.addps = false
	cc.Add(node)
	node.ifvotesummary = false
	node.ifsendvote3 = false
	node.ifsavevotesummary = false
	node.VoteSummary = VoteSummary{}
}

func (cc *CSchain) Add(node *Node) {
	temp_block := NewCSBlock(cc.CurrentBlock.Epoch, cc.CurrentBlock.Origin, cc.CurrentBlock.PSBlock, cc.Votetable, cc.CurrentBlock.PrevHash)
	cc.Blocks = append(cc.Blocks, temp_block)
	err := os.MkdirAll("./data/"+cc.NodeID+"/csblock", os.ModePerm)
	if err != nil {
	}
	cc.Len += 1

	SaveCSBlockToFile(temp_block, "./data/"+cc.NodeID+"/csblock/"+strconv.Itoa(cc.Len))
	CSchainlen += 1
	cc.Votetable = make(map[int][]string)
	cc.CurrentBlock = NewCSBlock(temp_block.Epoch+1, cc.NodeID, []PSBlock{}, map[int][]string{}, cc.CurrentBlock.BlockHeader.Hash)
	elapsed := time.Since(Start)
	err = saveElapsedTimeToFile(node, elapsed, "./data/"+cc.NodeID+"/csblock/"+strconv.Itoa(cc.Len)+"_time")
	if err != nil {
		fmt.Println("save file error:", err)
		return
	}

}

func (cc *CSchain) Run(bc *PSchain, node *Node) {
	go SetTime2(cc, bc, node)
	for {
		select {
		case temp_block := <-cc.BlockQueue:
			if temp_block.Epoch == int16(cc.Len)+1 {
				// add to current block's ps array
				cc.CurrentBlock.PSBlock = append(cc.CurrentBlock.PSBlock, temp_block)
				//fmt.Println(cc.NodeID, "-----: Receive from ps", temp_block.PSindex, "block", temp_block.Epoch)
				//产生一个投票
				if !(contains(node.Advepoch, int(temp_block.Epoch))) {
					newvote := Vote1{"", temp_block.Epoch, temp_block.PSindex, true, temp_block.BlockHeader.Hash, [32]byte{}}
					cc.VoteQueue_send1 <- newvote
				}
			}
		case temp_vote := <-cc.VoteQueue1:
			if temp_vote.Epoch == int16(cc.Len)+1 {
				nowvoteRaw := cc.Votetable[temp_vote.PSindex]
				if stringInSlice(temp_vote.Origin, nowvoteRaw) {
					// Origin has already existed in this Block's Nodeid
				} else {
					nowvoteRaw = append(nowvoteRaw, temp_vote.Origin)
					cc.Votetable[temp_vote.PSindex] = nowvoteRaw
					//fmt.Println(cc.NodeID, "--------Control shard :recevied ps", temp_vote.PSindex, "vote from", temp_vote.Origin, "epoch", temp_vote.Epoch)
					if node.CSLeader == 1 && node.VotesummaryRound == bc.Len && node.ifvotesummary == false {
						node.ifvotesummary = true
						node.VoteSummary = NewVoteSummary(node.Addr, int16(bc.Len)+1, node.CSID)
						SendAfter(delta, node, bc, cc)
					}
					// if node.CSLeader == 1 {
					// 	if node.ifvotesummary == true && node.VotesummaryRound == bc.Len {
					// 		if nowvotesummary, ok := node.VoteSummary.AgreeMap.Load(temp_vote.PSindex); ok {
					// 			node.VoteSummary.AgreeMap.Store(temp_vote.PSindex, Votelist{nowvotesummary.Num + 1, append(nowvotesummary.Nodelist, temp_vote.Origin)})
					// 		} else {
					// 			var temp []string
					// 			node.VoteSummary.AgreeMap.Store(temp_vote.PSindex, Votelist{1, append(temp, temp_vote.Origin)})
					// 		}
					// 	}
					// }
				}
			}
		case temp_vote_summary := <-cc.VoteSummaryQueue:
			if temp_vote_summary.Epoch == int16(cc.Len)+1 {
				cc.CurrentVoteSummary = temp_vote_summary
				fmt.Println(cc.NodeID, "-----: Receive the votesummary of cs", temp_vote_summary.CSindex, "epoch", temp_vote_summary.Epoch)
				node.ifvotesummary = true
				var temp []string
				cc.Vote4votesummary_vote2.Store(temp_vote_summary.Signature, votesummarystatus{1, temp, &temp_vote_summary})
				cc.Vote4votesummary_vote3.Store(temp_vote_summary.Signature, votesummarystatus{0, temp, &temp_vote_summary})
				//产生一个投票
				if !(contains(node.Advepoch, int(temp_vote_summary.Epoch))) {
					newvote := Vote2{node.ID, temp_vote_summary.Epoch, temp_vote_summary.CSindex, true, temp_vote_summary.Signature, [32]byte{}}
					cc.VoteQueue_send2 <- newvote
				}

			}
		case temp_vote2 := <-cc.VoteQueue2:
			if temp_vote2.Epoch == int16(bc.Len)+1 && node.ifsendvote3 == false {
				if nowvoteRaw, ok := cc.Vote4votesummary_vote2.Load(temp_vote2.VoteSummaryHash); ok {
					nowvote, ok := nowvoteRaw.(votesummarystatus)
					if !ok {
						// If the type assertion fails, the error is handled
						log.Printf("Error: Vote4NextBlock value of key %v has wrong type\n", temp_vote2.VoteSummaryHash)
						continue
					}
					if stringInSlice(temp_vote2.Origin, nowvote.Nodeid) {
						// Origin already exists in the Nodeid of the Block
					} else {
						nowvote.Nodeid = append(nowvote.Nodeid, temp_vote2.Origin)
						cc.Vote4votesummary_vote2.Store(temp_vote2.VoteSummaryHash, votesummarystatus{nowvote.Num + 1, nowvote.Nodeid, nowvote.Votesummary})
						fmt.Println(bc.NodeID, "--------:recevied vote_2 of votesummary from", temp_vote2.Origin)
						cc.Check_vote2(node, temp_vote2)
					}
				} else {
					//fmt.Println(bc.NodeID, "Stuck here2, please restart (server concurrency is poor)")
					//go cc.insertvote_delay2(temp_vote2)
				}
			} else if temp_vote2.Epoch > int16(bc.Len)+1 {
				//fmt.Println(bc.NodeID, "receive", temp_vote2.Origin, "epoch", temp_vote2.Epoch, "'s vote2")
				go cc.insertvote_delay2(temp_vote2)
			} else {

			}
		case temp_vote3 := <-cc.VoteQueue3:
			if temp_vote3.Epoch == int16(bc.Len)+1 && node.ifsavevotesummary == false {
				if nowvoteRaw, ok := cc.Vote4votesummary_vote3.Load(temp_vote3.VoteSummaryHash); ok {
					nowvote, ok := nowvoteRaw.(votesummarystatus)
					if !ok {
						// If the type assertion fails, the error is handled
						log.Printf("Error: Vote4NextBlock value of key %v has wrong type\n", temp_vote3.VoteSummaryHash)
						continue
					}
					if stringInSlice(temp_vote3.Origin, nowvote.Nodeid) {
						// Origin already exists in the Nodeid of the Block
					} else {
						nowvote.Nodeid = append(nowvote.Nodeid, temp_vote3.Origin)
						cc.Vote4votesummary_vote3.Store(temp_vote3.VoteSummaryHash, votesummarystatus{nowvote.Num + 1, nowvote.Nodeid, nowvote.Votesummary})
						//fmt.Println(bc.NodeID, "--------:recevied vote_final of votesummary from", temp_vote3.Origin)
						if (cc.Save_votesummary(node)) && (node.addps == false) {
							bc.GenerateBlock(node)
						}

					}
				} else {
					//fmt.Println(bc.NodeID, "Stuck here3, please restart (server concurrency is poor)")
					go cc.insertvote_delay3(temp_vote3)
				}
			} else if temp_vote3.Epoch > int16(bc.Len)+1 {
				//fmt.Println(bc.NodeID, "receive", temp_vote3.Origin, "epoch", temp_vote3.Epoch, "'s vote")
				go cc.insertvote_delay3(temp_vote3)
			} else {

			}
		}

	}
}

func (cc *CSchain) Save_votesummary(node *Node) bool {
	cc.Vote4votesummary_vote3.Range(func(key, value interface{}) bool {
		blockvotestatus, ok := value.(votesummarystatus)
		if !ok {
			return false
		}
		if blockvotestatus.Num >= (csnum/2) && blockvotestatus.Votesummary.Epoch == int16(cc.Len)+1 {
			cc.VoteSummarys = append(cc.VoteSummarys, *blockvotestatus.Votesummary)
			cc.CurrentVoteSummary = *blockvotestatus.Votesummary
			node.VoteSummary = *blockvotestatus.Votesummary
			cc.Vote4votesummary_vote2.Delete(key)
			cc.Vote4votesummary_vote3.Delete(key)
			//fmt.Println(cc.NodeID, "--------:Votesummary", node.VoteSummary.Epoch, "reached consensus")
			err := os.MkdirAll("./data/"+cc.NodeID+"/Votesummary", os.ModePerm)
			if err != nil {
			}
			SaveVotesummaryToFile(cc.CurrentVoteSummary, "./data/"+cc.NodeID+"/Votesummary/"+strconv.Itoa(int(node.VoteSummary.Epoch)))
			//fmt.Println(cc.NodeID, "--------:Votesummary", node.VoteSummary.Epoch, "is saved")
			node.PSchain.GenerateBlock(node)
			node.ifsavevotesummary = true
			//node.VoteSummary = VoteSummary{}
			return true
		}
		return false
	})
	return false
}

func (cc *CSchain) ccinsertvote_delay(temp_vote Vote1) {
	time.Sleep(1 * time.Second)
	cc.VoteQueue1 <- temp_vote
}

func (cc *CSchain) ccinsertblock_delay(temp_block PSBlock) {
	time.Sleep(1 * time.Second)
	cc.BlockQueue <- temp_block
}

func saveElapsedTimeToFile(node *Node, elapsed time.Duration, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "Generate Block use time: %d\n", node.T2+T1)
	if err != nil {
		return err
	}

	return nil
}

func (cc *CSchain) Check_vote2(node *Node, temp_vote Vote2) {
	cc.Vote4votesummary_vote2.Range(func(key, value interface{}) bool {
		blockvotestatus, ok := value.(votesummarystatus)
		if !ok {
			return false
		}
		if blockvotestatus.Num > (csnum/2) && blockvotestatus.Votesummary.Epoch == int16(cc.Len)+1 && node.ifsendvote3 == false {
			if !(contains(node.Advepoch, int(temp_vote.Epoch))) {
				newvote := Vote3{node.ID, temp_vote.Epoch, temp_vote.CSindex, true, temp_vote.VoteSummaryHash, [32]byte{}}
				cc.VoteQueue_send3 <- newvote
				node.ifsendvote3 = true
			}
			return true
		}
		return false
	})
}

func (cc *CSchain) insertvote_delay2(temp_vote2 Vote2) {
	time.Sleep(1 * time.Second)
	cc.VoteQueue2 <- temp_vote2
}
func (cc *CSchain) insertvote_delay3(temp_vote3 Vote3) {
	time.Sleep(1 * time.Second)
	cc.VoteQueue3 <- temp_vote3
}
