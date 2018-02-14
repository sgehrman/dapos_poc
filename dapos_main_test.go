package main


import "testing"
import "strconv"
//import "time"


var voteChannel chan Vote = make(chan Vote)
var voteCounter= NewVoteCounter(voteChannel)



func TestAccounts ( t *testing.T ) {

	for i := 0; i < 10 ; i++ {
		CreateAccount( strconv.Itoa(i), 100)
	}
//	for i := 0; i < 10 ; i++ {
//		CreateAccount( strconv.Itoa(i), 100)
//	}

	accounts := getAccounts()

	for i := 0; i < 10 ; i++ {

		if (*accounts).accounts[strconv.Itoa(i)].Balance == 100{

		} else {
			t.Error ("Balance was not recorded correctly")
		}

	}

}


func setup ( delegate_count int, c chan Transaction) *VoteCounter {

	//create delegates
//	var c chan Transaction = make(chan Transaction)

	voteCounter.Start()
	for i := 0; i < delegate_count; i++ {
		delegate := NewDelegate(i, delegate_count, c, voteChannel)
		go delegate.Start()
	}

	return voteCounter
}


func createAccounts ( num int){
	names = names[:0]

	for i := 0; i < num ; i++ {
		// update global names
		// TODO: change all this when we move away from POC
		names = append ( names, strconv.Itoa (i))
		CreateAccount( strconv.Itoa(i), 100)
	}
}


func TestSequencedSlam ( t *testing.T ) {

	var delegate_count = 4
	var account_count = 1000

	var c chan Transaction = make(chan Transaction)

	createAccounts(account_count)

	voteCounter := setup(delegate_count,c )

	for i := 1; i < 100; i++ {
		runSome(voteCounter, delegate_count, c, i)
//		time.Sleep(time.Second * 5)

	}
	for _, acct := range getAccounts().accounts {
		printAccounts(acct)
	}




}