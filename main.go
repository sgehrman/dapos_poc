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
	validators int
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
	//first transaction is 1 because geesis block is 0
	transaction_id := 1
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
		new_transaction := transaction{transaction_id, from, to, value, time.Now(), (delegate_count+1), 0}

		//increment transaction id so no repeats.
		//in actual code this would be transaction hash
		transaction_id++

		c <- new_transaction
	}
}