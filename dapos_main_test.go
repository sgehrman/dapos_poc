package main


import "testing"
import "strconv"
import "time"
import (
	"fmt"
)
//import "time"


var voteChannel chan Vote = make(chan Vote)
var voteCounter= NewVoteCounter(voteChannel)


type Timer struct {
	state	bool
	start 	int64
	duration int64
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}


func StartTimer ( ) *Timer {
	timer := new ( Timer )

	(*timer).start = makeTimestamp()
	(*timer).state = true


	return timer
}

func duration ( timer *Timer ) int64 {

	end := makeTimestamp ()

	// leave it to the functions above to control time.  we are just a pee ons
	//	(*timer).state = false
	(*timer).duration = end - (*timer).start

	return (*timer).duration
}


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

	start := StartTimer()
	for i := 1; i < 100000; i++ {
		go runSome(voteCounter, delegate_count, c, i)
//		time.Sleep(time.Second * 5)

	}

	duration := duration(start)

	fmt.Println(duration)
	for _, acct := range getAccounts().accounts {
		printAccounts(acct)
	}




}