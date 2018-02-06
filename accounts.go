package main

import (
	"sync"
	"time"
	"fmt"
)

type singleton struct {
	accounts map[string]*Account
}

var instance *singleton
var once sync.Once

func GetAccount(member string) *Account {
	for _, acct := range getAccounts().accounts {
		if(acct.Name == member) {
			return acct
		}
	}
	panic(fmt.Sprintf("Account for %s not found", member))

}

//need to do something better for transaction id
func CreateAccount(newMember string, initialBalance int) {
	new_wallet := Transaction{0, "dl", newMember, initialBalance, time.Now(), 0, 1}
	txs := make([]Transaction, 0)
	txs = append(txs, new_wallet)
	acct := Account{"value", newMember, initialBalance, txs}
	getAccounts().accounts[newMember] = &acct
}


func getAccounts() *singleton {
	once.Do(func() {
		instance = &singleton{}
		instance.accounts = make(map[string]*Account, 0)
	})
	return instance
}