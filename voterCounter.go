package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Votes struct {
	TotalVotesSoFar []bool
	IsAheadOfTime   bool
}

type Vote struct {
	TransactionId int
	YesNo         bool
}

func (node *Node) StartVoteCounting() {
	fmt.Println("StartVoteCounting()")
	go func() {
		for {
			select {
			case vote := <-node.VoteChannel:

				// we have received a vote
				theVotesForTx := node.AllVotes[vote.TransactionId]

				if theVotesForTx == nil {

					// We got opinion/vote about a TX but we did not get the actual TX yet

					node.AllVotes[vote.TransactionId] = &Votes{
						TotalVotesSoFar: []bool{},
						IsAheadOfTime:   true,
					}

					theVotesForTx = node.AllVotes[vote.TransactionId]
				} else {

					// We got opinion/vote about an existing TX

					theVotesForTx.IsAheadOfTime = false
				}

				theVotesForTx.TotalVotesSoFar = append(theVotesForTx.TotalVotesSoFar, vote.YesNo)

				if !theVotesForTx.IsAheadOfTime {
					var tx = node.TxFromChainById[vote.TransactionId]

					if tx != nil {
						node.VoteCounterProcessTx(tx)
					}
				}
			}
		}
	}()
}

func (node *Node) VoteCounterProcessTx(tx *Transaction) {

	theVotesForTx := node.AllVotes[tx.Id]
	if theVotesForTx != nil {
		fmt.Println(fmt.Sprintf("GotVote()       | Tx_%d(%s -> %s) | %s", tx.Id, tx.From, tx.To, node.Wallet.Id))

		var totalDelegates = len(tx.CurrentValidators)

		if totalDelegates == len(theVotesForTx.TotalVotesSoFar) {
			if theVotesForTx.isValid(totalDelegates) {
				updateAccounts(node.TxFromChainById[tx.Id])

				fromAcct := (*getNodes()[string(tx.From)]).Wallet
				toAcct := (*getNodes()[string(tx.To)]).Wallet
				fmt.Println(fmt.Sprintf("ProcessedVote() | Tx_%d(%s -> %s) | %s (%d, %d)", tx.Id, tx.From, tx.To, node.Wallet.Id, fromAcct.Balance, toAcct.Balance))
			} else {
				fmt.Println("Vote Failed")
			}
		}
	}
}

func (v Votes) isValid(totalDelegates int) bool {
	var positiveCount = 0
	var negativeCount = 0
	for _, value := range v.TotalVotesSoFar {
		if value {
			positiveCount++
		} else {
			negativeCount++
		}
	}

	if positiveCount == totalDelegates {
		return true
	}

	return false
}

func updateAccounts(t *Transaction) {
	log.Printf("Update Accounts: %d", t.Id)
	fromAcct := (*getNodes()[string(t.From)]).Wallet
	toAcct := (*getNodes()[string(t.To)]).Wallet
	fromAcct.Balance -= t.Value
	toAcct.Balance += t.Value
	// fromAcct.Transactions = append(fromAcct.Transactions, t)
	// toAcct.Transactions = append(toAcct.Transactions, t)
}
