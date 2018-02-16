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
var delegate_count = 4
var delegates []*Delegate


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

		if (*accounts).accounts[strconv.Itoa(i)].Balance == 1000{

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
		delegates = append(delegates, delegate)
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
		CreateAccount( strconv.Itoa(i), 10000000900)
	}
}


func auditVotes ( votes *VoteCounter ) {

	var total int = 0
	var yes_votes int = 0
	var no_votes int = 0
	var mixed_concensus int = 0
	var all_agreed int = 0

	var received_all_votes int = 0
	var missing_votes  int = 0

	for key, _ := range (*votes).votes {

		v := *votes
		v1 := v.votes[key]
		c := v1.VoteCount

		fmt.Println(c)
		//
		if (*votes).votes[key].VoteCount != delegate_count {
			missing_votes++
		} else {
			received_all_votes++
		}

		// analyze votes
		yes := 0
		no := 0
		for j := 0 ; j < delegate_count ; j++ {

			if (*votes).votes[key].VoteYesNo != nil  && (*votes).votes[key].VoteYesNo[j] == true {
				yes++
			} else {
				no++
			}

		}
		if yes == delegate_count {
			all_agreed++
		} else {
			mixed_concensus++
		}

		// TODO:  Needs logic here
		if yes > no {
			yes_votes++
		} else {
			no_votes++
		}

		total++

	}

	fmt.Println( "Total Transactions %d", total)
	fmt.Println("Total Mix Concensus %d", mixed_concensus)
	fmt.Println ( "Total that Agreed %d", all_agreed)
	fmt.Println("Total Votes Rec %d", received_all_votes)
	fmt.Println( "Total Missing Votes %d", missing_votes)

}

func TestSequencedSlam ( t *testing.T ) {


	var account_count = 1000
	var transaction_count = 1000

	var c chan Transaction = make(chan Transaction)

	createAccounts(account_count)

	voteCounter := setup(delegate_count,c )

	start := StartTimer()
	for i := 1; i < transaction_count; i++ {
		go runSome(voteCounter, delegate_count, c, i)
//		time.Sleep(time.Second * 5)

	}

	time.Sleep(time.Second * 1)

	duration := duration(start)
	fmt.Println(duration)

	fmt.Println( "Total %d", voteCounter.TotalPendingBlocks)
	fmt.Println( "Completed %d", voteCounter.CompletedBlocks)

	/*for {
		if voteCounter.TotalPendingBlocks >= transaction_count-1 {
			break
		}
		time.Sleep(time.Millisecond * 1 )

	}


	time.Sleep(time.Second * 5 )
*/



	//auditVotes(voteCounter)



}