package main

import (
	"fmt"
	"time"
	"math/rand"
	 log "github.com/sirupsen/logrus"
	 "os"

)

var names = []string{"Bob", "Chris", "Greg", "Muhammad", "Nicolae", "Zane"}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.ErrorLevel)

	log.Info("Dapos POC Starting")
}


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
	var c chan Transaction = make(chan Transaction)
	var voteChannel chan Vote = make(chan Vote)

	var voteCounter= NewVoteCounter(voteChannel)
	voteCounter.Start()
	for i := 0; i < delegate_count; i++ {
		delegate := NewDelegate(i, delegate_count, c, voteChannel)
		go delegate.Start()
	}

	for i := 1; i < 100; i++ {
		runSome(voteCounter, delegate_count, c,i)
//		time.Sleep(time.Second * 5)
		fmt.Println( i)
	}
	for _, acct := range getAccounts().accounts {
		printAccounts(acct)
	}
}

func runSome(voteCounter *VoteCounter, delegateCount int, c chan Transaction, transactionId int) {

	from := getRandomAccount(nil)
	to := getRandomAccount(from)
	amount := GetRandomNumber(20)


	log.WithFields( log.Fields { "From " : from.Name, "To " : to.Name, "Amount " : amount}).Info ("Transaction receipt")

	//input things forever
	//first transaction is 1 because geesis block is 0
	var delay bool = false
	//generate new transaction from
	new_transaction := Transaction{transactionId, from.Name, to.Name, amount, time.Now(), (delegateCount + 1), 0}
	voteCounter.AddVoting(new_transaction, delegateCount)
	sendTransaction(new_transaction, c,  delay)

	//increment transaction id so no repeats.
	//in actual code this would be transaction hash

	log.WithFields( log.Fields {"From " : from.Name, "From Balance " : from.Balance, "To " : to.Name, "To Balance" : to.Balance}).Info ("Balances")
}


func sendTransaction(transaction Transaction, c chan Transaction, delay bool) {

	log.WithFields( log.Fields { "Transaction ID" : transaction.Id, "Transaction From" : transaction.From, "Transaction To" : transaction.To, "Transaction Value" : transaction.Value}).Info ("Sending Transaction")

	if delay {
		log.Info ("delay true")
//		fmt.Println("delay true")
		time.Sleep(time.Second * 10)
	}
	c <- transaction
}

func printAccounts(acct *Account) {

	log.WithFields( log.Fields { "Account Name" : acct.Name, "Balance": acct.Balance}).Info ("New Balances")

	var balance int
	for _, t := range acct.Transactions {
		if t.From == "dl" {
			balance = t.Value
		} else if t.From == acct.Name {
			balance -= t.Value
		} else {
			balance += t.Value
		}

		log.WithFields( log.Fields { "Transaction ID" : t.Id, "From " :t.From, "To" : t.To, "Value" : t.Value, "Balance" : balance} ).Info ("Transaction:")
	}
}