package main

import (
	"fmt"
	"reflect"
	"time"
)

type transaction struct {
	id int
	from string
	to string
	value int
	time time.Time
	delegate_id int
}

func delegate(id int, nodes int, initial_state map[string]int, c chan transaction, test chan string) {

	//create local balance variable for this delegate
	balance := make(map[string]int)

	for k,v := range initial_state {
		balance[k] = v
	}

	transaction_validity := make(map[int]string)

	//announce yor existance
	fmt.Printf("Delegate %d\n", id)

	//listen for transactions forever
	for {
		msg := <- c
		//if transaction came from non-delegate node (new)
		if msg.delegate_id > nodes {

			//check to see if transaction is valid
			valid := process_transaction(balance, msg.from, msg.to, msg.value)
			if valid {
				transaction_validity[msg.id] = "valid"
				fmt.Printf("delegate %d: received valid transaction %d from an ndn \n", id, msg.id)

				//set the delegate id to current id and broadcast the valid transaction to other nodes
				msg.delegate_id = id
				for i := 0; i < nodes-1; i++ {
					c <- msg
				}

			} else {
				fmt.Printf("delegate %d: received invalid transaction %d from an ndn \n", id, msg.id)
			}
			time.Sleep(time.Millisecond)

		//transactions from delegates should be reevaluated
		} else {
			fmt.Printf("delegate %d: received transaction %d from delegate %d \n", id, msg.id, msg.delegate_id)

		}
	}

}

func process_transaction(balance map[string]int, from string, to string, value int) bool {
	if balance[from] >= value {
		//if transaction is valid, update existing balances
		balance[from] = balance[from] - value
		balance[to] = balance[to] + value
		return true
	} else {
		return false
	}
}

func main() {
	fmt.Println("DAPoS Simulation!")

	//initialize a0 w some money
	balance := make(map[string]int)
	balance["a0"] = 5

	//Prompt for munber of delegates
	fmt.Println("How many elected delegates?")
	var delegate_count int
	fmt.Scanf("%d", &delegate_count)
	fmt.Printf("%d elected delegates \n", delegate_count)
	fmt.Println(reflect.TypeOf(delegate_count))

	//create delegates
	var c chan transaction = make(chan transaction)
	var test chan string = make(chan string)

	for i := 0; i < delegate_count; i++ {
		go delegate(i, delegate_count, balance, c, test)
	}

	//input things forever
	transaction_id := 0
	var from string
	var to string
	var value int
	for {
		//take in
		fmt.Println("transaction from?")
		fmt.Scanf("%s", &from)
		fmt.Println("transaction to?")
		fmt.Scanf("%s", &to)
		fmt.Println("transaction value?")
		fmt.Scanf("%d", &value)

		//generate new transaction from
		new_transaction := transaction{transaction_id, from, to, value, time.Now(), (delegate_count+1)}

		//increment transaction id so no repeats.
		//in actual code this would be transaction hash
		transaction_id++

		c <- new_transaction
	}
}