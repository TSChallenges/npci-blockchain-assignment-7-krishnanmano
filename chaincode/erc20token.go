package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Token represents the structure of the token state
type Token struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply int    `json:"totalSupply"`
}

// TokenContract provides functions for managing the token
type TokenContract struct {
	contractapi.Contract
}

// Balance reflects the balance of a user
type Balance struct {
	Balance int `json:"balance"`
}

type Approval struct {
	Owner     string `json:"owner"`
	Spender   string `json:"spender"`
	Allowance int    `json:"allowance"`
}

// TODO: InitLedger for token initialization
func (t *TokenContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("user", "UserA")
	if err != nil {
		return fmt.Errorf("unauthorized operation: user should be admin")
	}

	token := Token{
		Name:        "BNB-Token",
		Symbol:      "BNB",
		Decimals:    18,
		TotalSupply: 200000000,
	}
	tokenAsBytes, _ := json.Marshal(token)

	err = ctx.GetStub().PutState("token", tokenAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	// Initialize admin balance
	admin := Balance{Balance: token.TotalSupply}
	adminAsBytes, _ := json.Marshal(admin)
	return ctx.GetStub().PutState("UserA", adminAsBytes)
}

// TODO: MintTokens for minting tokens
func (t *TokenContract) MintTokens(ctx contractapi.TransactionContextInterface, amount int) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("user", "UserA")
	if err != nil {
		return fmt.Errorf("unauthorized operation: user should be admin")
	}

	adminAsBytes, err := ctx.GetStub().GetState("UserA")
	if err != nil {
		return err
	}

	var admin Balance
	json.Unmarshal(adminAsBytes, &admin)

	admin.Balance += amount
	updatedAdminAsBytes, _ := json.Marshal(admin)
	return ctx.GetStub().PutState("UserA", updatedAdminAsBytes)
}

// TODO: TransferTokens for transferring tokens
func (t *TokenContract) TransferTokens(ctx contractapi.TransactionContextInterface, from string, to string, amount int) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("user", from)
	if err != nil {
		return fmt.Errorf("unauthorized operation: calling user should be sender")
	}

	fromAsBytes, _ := ctx.GetStub().GetState(from)
	toAsBytes, _ := ctx.GetStub().GetState(to)

	var fromBalance, toBalance Balance
	json.Unmarshal(fromAsBytes, &fromBalance)
	json.Unmarshal(toAsBytes, &toBalance)

	if fromBalance.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	fromBalance.Balance -= amount
	toBalance.Balance += amount

	updatedFromAsBytes, _ := json.Marshal(fromBalance)
	updatedToAsBytes, _ := json.Marshal(toBalance)
	ctx.GetStub().PutState(from, updatedFromAsBytes)
	ctx.GetStub().PutState(to, updatedToAsBytes)

	return nil
}

// TODO: GetBalance to check the balance
func (t *TokenContract) GetBalance(ctx contractapi.TransactionContextInterface, owner string) (int, error) {
	ownerAsBytes, _ := ctx.GetStub().GetState(owner)

	var ownerBalance Balance
	json.Unmarshal(ownerAsBytes, &ownerBalance)

	return ownerBalance.Balance, nil
}

// TODO: ApproveSpender for approving spending
func (t *TokenContract) ApproveSpender(ctx contractapi.TransactionContextInterface, owner string, spender string, amount int) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("user", owner)
	if err != nil {
		return fmt.Errorf("unauthorized operation: calling user should be owner")
	}

	// Create an approval object
	approval := Approval{
		Owner:     owner,
		Spender:   spender,
		Allowance: amount,
	}

	approvalAsBytes, err := json.Marshal(approval)
	if err != nil {
		return err
	}

	// Use a composite key to uniquely identify the allowance for a spender by an owner
	approvalKey, err := ctx.GetStub().CreateCompositeKey("approval", []string{owner, spender})
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(approvalKey, approvalAsBytes)
}

// TODO: TransferFrom for transferring from approved spenders
func (t *TokenContract) TransferFrom(ctx contractapi.TransactionContextInterface, owner string, spender string, recipient string, amount int) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("user", spender)
	if err != nil {
		return fmt.Errorf("unauthorized operation: calling user should be spender")
	}

	// Retrieve the approval to check if the spender is allowed to spend tokens
	approvalKey, err := ctx.GetStub().CreateCompositeKey("approval", []string{owner, spender})
	if err != nil {
		return err
	}

	approvalAsBytes, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return err
	}

	var approval Approval
	if err := json.Unmarshal(approvalAsBytes, &approval); err != nil {
		return err
	}

	if approval.Allowance < amount {
		return fmt.Errorf("spender does not have enough allowance")
	}

	// Decrease the allowance
	approval.Allowance -= amount
	updatedApprovalAsBytes, err := json.Marshal(approval)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState(approvalKey, updatedApprovalAsBytes)

	// Proceed with transferring tokens from owner to recipient
	ownerAsBytes, err := ctx.GetStub().GetState(owner)
	if err != nil {
		return err
	}

	recipientAsBytes, err := ctx.GetStub().GetState(recipient)
	if err != nil {
		return err
	}

	var ownerBalance, recipientBalance Balance
	json.Unmarshal(ownerAsBytes, &ownerBalance)
	json.Unmarshal(recipientAsBytes, &recipientBalance)

	if ownerBalance.Balance < amount {
		return fmt.Errorf("owner does not have enough balance")
	}

	// Transfer the tokens
	ownerBalance.Balance -= amount
	recipientBalance.Balance += amount

	updatedOwnerAsBytes, _ := json.Marshal(ownerBalance)
	updatedRecipientAsBytes, _ := json.Marshal(recipientBalance)
	ctx.GetStub().PutState(owner, updatedOwnerAsBytes)
	ctx.GetStub().PutState(recipient, updatedRecipientAsBytes)

	return nil
}

// TODO: BurnTokens for burning tokens
func (t *TokenContract) BurnTokens(ctx contractapi.TransactionContextInterface, amount int) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("user", "UserA")
	if err != nil {
		return fmt.Errorf("unauthorized operation: user should be admin")
	}

	adminAsBytes, err := ctx.GetStub().GetState("UserA")
	if err != nil {
		return err
	}

	var admin Balance
	json.Unmarshal(adminAsBytes, &admin)

	admin.Balance -= amount
	updatedAdminAsBytes, _ := json.Marshal(admin)
	return ctx.GetStub().PutState("UserA", updatedAdminAsBytes)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&TokenContract{})
	if err != nil {
		fmt.Printf("Error creating token chaincode: %v", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting token chaincode: %v", err)
	}
}
