package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

var payouts = []Payout{
	Payout{
		Amount:               999999,
		LifetimePayoutAmount: 0,
		PaymentMethod: PaymentMethod{
			Name: "jo",
		},
	}, Payout{
		Amount:               1000000,
		LifetimePayoutAmount: 100000,
		PaymentMethod: PaymentMethod{
			Name: "joe",
		},
	}, Payout{
		Amount:               1000001,
		LifetimePayoutAmount: 100000000,
		PaymentMethod: PaymentMethod{
			Name: "jo",
		},
	},
}

type PayoutInput struct {
	Account Account
	Payout  Payout
	Result  Result
}

type Result struct {
	IsKYCNameEqualToPMNameFlagged   bool
	IsPayoutAmountTooHighFlagged    bool
	IsFirstPayoutFlagged            bool
	ManualReviewPayoutAmountFlagged bool
	IsVerifiedFlagged               bool
}

type Account struct {
	Name     string
	Verified bool
}

type Payout struct {
	Amount               int
	LifetimePayoutAmount int
	PaymentMethod        PaymentMethod
}

type PaymentMethod struct {
	Name string
}

func main() {
	rules := `
rule IsKYCNameEqualToPMName  "KYC Name Equal to Payment Method Name" {
	when
		Payout.PaymentMethod.Name != Account.Name
	then
		Result.IsKYCNameEqualToPMNameFlagged = true;
		Retract("IsKYCNameEqualToPMName");
}

rule IsPayoutAmountTooHigh "Payout Amount Too High" salience 10 {
when
	Payout.Amount > 1000000
then
	Result.IsPayoutAmountTooHighFlagged = true;
	Retract("IsPayoutAmountTooHigh");
}

rule IsFirstPayout "First Payout" salience 1 {
when
	Payout.LifetimePayoutAmount == 0
then
	Result.IsFirstPayoutFlagged = true;
	Retract("IsFirstPayout");
}

rule ManualReviewPayoutAmount "Manual Review Payout" salience 2 {
when
	Payout.LifetimePayoutAmount > 500000 &&
	Payout.LifetimePayoutAmount < 1000000
then
	Result.ManualReviewPayoutAmountFlagged = true;
	Retract("ManualReviewPayoutAmount");
}

rule IsVerified "Verified" salience 5 {
when
	Account.Verified == false
then
	Result.IsVerifiedFlagged;
	Retract("IsVerified");
}
`
	lib := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(lib)
	err := ruleBuilder.BuildRuleFromResource("PayoutRules", "0.1.1", pkg.NewBytesResource([]byte(rules)))
	if err != nil {
		log.Fatal(err)
	}

	payoutKnowledgeBase, err := lib.NewKnowledgeBaseInstance("PayoutRules", "0.1.1")
	if err != nil {
		log.Fatal(err)
	}

	//eng := engine.NewGruleEngine() - defaults to 5000 cycles
	eng := &engine.GruleEngine{MaxCycle: 500}

	for _, payout := range payouts {
		now := time.Now()
		input := PayoutInput{
			Account: Account{
				Name: "jo",
			},
			Payout: payout,
			Result: Result{},
		}

		dataCtx := ast.NewDataContext()
		err := dataCtx.Add("Result", &input.Result)
		if err != nil {
			log.Fatal(err)
		}

		err = dataCtx.Add("Account", &input.Account)
		if err != nil {
			log.Fatal(err)
		}

		err = dataCtx.Add("Payout", &input.Payout)
		if err != nil {
			log.Fatal(err)
		}

		err = eng.Execute(dataCtx, payoutKnowledgeBase)
		if err != nil {
			log.Fatal(err)
		}

		printOutcome(now, input)
	}
}

func printOutcome(now time.Time, input PayoutInput) {
	fmt.Println("\n--------------OutCome------------------------")
	fmt.Println("time elapse:", time.Since(now))
	fmt.Println(fmt.Sprintf("Result:%+v. Account:%+v. Payout:%+v\n", input.Result, input.Account, input.Payout))
}
