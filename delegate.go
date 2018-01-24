package main

import (
	"fmt"
	"time"
)

//block struct to keep the linked list
type block struct {
	prev_block *block
	next_block *block
	transaction transaction
}

func delegate(id int, nodes int, initial_state map[string]int, c chan transaction, test chan string) {

	//Genesis block
	//All delegates are instantiated with teh same genesis block
	premine_wallet_a0 := transaction{0, "dl", "a0", 10, time.Now(), 0, nodes}
	genesis_block := new(block)
	genesis_block.transaction = premine_wallet_a0
	//set current block pointer to genesis block
	current_block := genesis_block


	//announce yor existance
	fmt.Printf("Delegate %d\n", id)


	//listen for transactions forever
	for {
		msg := <- c
		//if transaction came from non-delegate node (new)
		if msg.delegate_id > nodes {

			//check to see if transaction is valid
			valid := process_transaction(genesis_block, msg)
			if valid {
				fmt.Printf("delegate %d: received valid transaction %d from an ndn \n", id, msg.id)
				//save the transaction to the chain
				new_block := block{nil, nil, msg}
				current_block.next_block = &new_block
				current_block = &new_block

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

			//TODO: process unseen transactions from other delegates
			//if transaction came from another delegate, check to see if it's been seen before then process it
			if !seen_transaction(msg.id, genesis_block) {
				valid := process_transaction(genesis_block, msg)
				if valid {
					fmt.Printf("delegate %d: received valid transaction %d from delegate %d \n", id, msg.id, msg.delegate_id)
					//save the transaction to the chain
					new_block := block{nil, nil, msg}
					current_block.next_block = &new_block
					current_block = &new_block

					//set the delegate id to current id and broadcast the valid transaction to other nodes
					msg.delegate_id = id
					for i := 0; i < nodes-1; i++ {
						c <- msg
					}

				} else {
					fmt.Printf("delegate %d: received invalid transaction %d from delegate %d \n", id, msg.id, msg.delegate_id)
				}
			}
		}
	}

}

func seen_transaction(id int, genesis_block *block) bool {
	pointer_block := genesis_block
	for pointer_block != nil {
		if pointer_block.transaction.id == id {
			return true
		}
		pointer_block = pointer_block.next_block
	}
	return false
}

//processing the transaction consists of:
//adding the transaction in it's proper place in the time-sorted linked list
//checking that transaction and the ones following it for validity
//return true if transaction is valid
func process_transaction(current_block *block, msg transaction) bool {
	//don't validate transactions on 0 or less
	if msg.value <= 0 {
		return false
	}

	//new balance maping
	new_balances := make(map[string]int)
	pointer_block := current_block

	//iterate until new_balances matches up with blockchain state at time of transaction
	for {
		if pointer_block.next_block == nil { //pointer_block is end of the chain
			fmt.Println("there are no more blocks after pointer_block")
			new_balances[pointer_block.transaction.to] += pointer_block.transaction.value

			if pointer_block.transaction.from != "dl" { //don't set a negative balance for premined transfers
				new_balances[pointer_block.transaction.from] -= pointer_block.transaction.value
			}
			break
		} else {
			fmt.Println("there are still more blocks after pointer_block")
			//break if msg goes after pointer_block, but before pointer_block.next_block
			if msg.time.After(pointer_block.transaction.time) && msg.time.Before(pointer_block.next_block.transaction.time) {

				fmt.Println("msg goes between tx %d and tx %d \n", pointer_block.transaction.id, pointer_block.next_block.transaction.id)
				new_balances[pointer_block.transaction.to] += pointer_block.transaction.value

				if pointer_block.transaction.from != "dl" { //don't set a negative balance for premined transfers
					new_balances[pointer_block.transaction.from] -= pointer_block.transaction.value
				}
				break
			} else {
				//if msg.time doesn't follow pointer_block, update new_balance and iterate
				new_balances[pointer_block.transaction.to] += pointer_block.transaction.value

				if pointer_block.transaction.from != "dl" { //don't set a negative balance for premined transfers
					new_balances[pointer_block.transaction.from] -= pointer_block.transaction.value
				}

				pointer_block = pointer_block.next_block
			}
		}
	}

	fmt.Printf("current block is %d \n", pointer_block.transaction.id)
	fmt.Printf("new_balances[%s]=%d \n", msg.from, new_balances[msg.from])
	//is new transaction valid?
	if new_balances[msg.from] >= msg.value {

		//if so make a new block and add it to the chain
		new_valid_block := block{pointer_block, pointer_block.next_block, msg}
		pointer_block.next_block = &new_valid_block
		if new_valid_block.next_block != nil {
			new_valid_block.next_block.prev_block = &new_valid_block
		}

		//update new_balances to reflect the recent addition
		new_balances[msg.from] = new_balances[msg.from] - msg.value
		new_balances[msg.to] = new_balances[msg.to] + msg.value

		//and then check following blocks for validity
		//using revoked_blocks to keep track of now invalid transactions
		pointer_block = &new_valid_block
		for pointer_block.next_block != nil {
			//is the following transaction valid?
			if new_balances[pointer_block.transaction.from] >= pointer_block.transaction.value {
				fmt.Printf("yay! transaction %d is still valid \n", pointer_block.transaction.id)
				//yay! transaction is valid - update balances and continue onto next block
				new_balances[pointer_block.transaction.from] -= pointer_block.transaction.value
				new_balances[pointer_block.transaction.to] += pointer_block.transaction.value

				if pointer_block.next_block == nil {
					break
				} else {
					pointer_block = pointer_block.next_block
				}
			}
			if new_balances[pointer_block.transaction.from] < pointer_block.transaction.value {
				fmt.Printf("turns out transaction %d is invalid \n", pointer_block.transaction.id)
				//oh no! this previously believed valid block is actually invaid! D:
				//remove block from list and keep going
				pointer_block.prev_block.next_block = pointer_block.next_block
				pointer_block.next_block.prev_block = pointer_block.prev_block
			}
		}
		//when finished, broadcast awesome new block and potentially broken transactions
		return true

	} else { //if new transaction is not valid, drop that bitch like it's hot
		return false
	}

}