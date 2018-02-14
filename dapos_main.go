package main

import (
	"fmt"
	"time"
	"math/rand"
	"log"
)

var names = []string{"Bob", "Chris", "Greg", "Muhammad", "Nicolae", "Zane"}

func GetRandomNumber(boundary int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(boundary)
}

func getRandomAccount(acct *Account) *Account {
	accounts := getAccounts().accounts
	randomNum := GetRandomNumber(len(accounts))
	if acct == nil {
		return accounts[names[randomNum]]
	}
	account := accounts[names[randomNum]]
	if(account.Name == acct.Name) {
		return getRandomAccount(acct)
	}
	return account
 }

func main() {
	fmt.Println("DAPoS Simulation!")

	CreateAccount("Bob", 100)
	CreateAccount("Chris", 100)
	CreateAccount("Greg", 100)
	CreateAccount("Muhammad", 100)
	CreateAccount("Nicolae", 100)
	CreateAccount("Zane", 100)

	//number of delegates
	var delegate_count int = 4

	//create delegates
	var c_node chan Transaction = make(chan Transaction)
	var c_delegate chan Transaction = make (chan Transaction)
	var voteChannel chan Vote = make(chan Vote)

	var voteCounter= NewVoteCounter(voteChannel)
	voteCounter.Start()
	for i := 0; i < delegate_count; i++ {
		delegate := NewDelegate(i, delegate_count, c_node, c_delegate, voteChannel)
		go delegate.StartNodeDaemon()
		go delegate.StartDelegateDaemon()
	}

	for i := 1; i < 15; i++ {
		runSome(voteCounter, delegate_count, c_node, c_delegate,i)
//		time.Sleep(time.Second * 5)
		fmt.Println( i)
	}
	for _, acct := range getAccounts().accounts {
		printAccounts(acct)
	}
}

func runSome(voteCounter *VoteCounter, delegateCount int, c_node chan Transaction, c_delegate chan Transaction, transactionId int) {
	from := getRandomAccount(nil)
	to := getRandomAccount(from)
	amount := GetRandomNumber(20)
	fmt.Printf("Transaction from %s : to %s for amount %d\n\n", from.Name, to.Name, amount)

	//input things forever
	//first transaction is 1 because geesis block is 0
	var delay bool = false
	//generate new transaction from
	new_transaction := Transaction{transactionId, from.Name, to.Name, amount, time.Now(), (delegateCount + 1), 0}
	voteCounter.AddVoting(new_transaction, delegateCount)
	sendTransaction(new_transaction, c_node, c_delegate, delay)

	//increment transaction id so no repeats.
	//in actual code this would be transaction hash
	log.Printf("\nAccount %s now has : %d and Account %s now has %d\n", from.Name, from.Balance, to.Name, to.Balance)
}


func sendTransaction(transaction Transaction, c_node chan Transaction, c_delegate chan Transaction, delay bool) {
	log.Printf("Sending Transaction: %d : %s : %s : %d", transaction.Id, transaction.From, transaction.To, transaction.Value)
	fmt.Println("SendTransaction")
	if delay {
		fmt.Println("delay true")
		time.Sleep(time.Second * 10)
	}
	c_node <- transaction
}

func printAccounts(acct *Account) {
	log.Printf("Account %s now has : %d \n", acct.Name, acct.Balance)
	var balance int
	for _, t := range acct.Transactions {
		if t.From == "dl" {
			balance = t.Value
		} else if t.From == acct.Name {
			balance -= t.Value
		} else {
			balance += t.Value
		}
		log.Printf("\tTransaction: id=%d  from %s to %s value=%d (balance = %d)", t.Id, t.From, t.To, t.Value, balance)
	}
}