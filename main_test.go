/*
Copyright (C) 2025 Andrew Flint.

This file is part of cas2trn.

Cas2trn is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Cas2trn is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with cas2trn.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"testing"
)

func TestHappyConfig(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	err := cfg.isValid()
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil\n")
	}
}

func TestHappyTransactKBAmount(t *testing.T) {
	t.Parallel()

	// configure Kiwibank full CSV statement
	cfg := kbFull

	err := cfg.isValid()
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	// test credit transaction using amount field
	flds := []string{"ZZ-YYYY-XXXXXXX-WW", "29-12-2023", "Automatic Payment Rates MISS E MACD ;Ref: Rates MISS E MACD",
		"AP", "Rates", "E", "", "", "", "", "MISS E MACD", "AA-BBBB-CCCCCCC-DD", "162.00", "", "162.00", "1434.23"}

	var trn transact

	err = trn.transact(flds, cfg)
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	expectAmount := 162.00
	gotAmount := trn.amount

	if gotAmount != expectAmount {
		t.Fatalf("wrong amount: expected==%v, got==%v\n", expectAmount, gotAmount)
	}

	expect := "2023-12-29"
	got := trn.date

	if got != expect {
		t.Fatalf("wrong date: expected==%v, got==%v\n", expect, got)
	}

	expect = "Automatic Payment Rates MISS E MACD ;Ref: Rates MISS E MACD"
	got = trn.memo

	if got != expect {
		t.Fatalf("wrong memo: expected==%q, got==%q\n", expect, got)
	}

	expect = "AA-BBBB-CCCCCCC-DD"
	got = trn.otherAcct

	if got != expect {
		t.Fatalf("wrong that account: expected==%v, got==%v\n", expect, got)
	}

	expect = "ZZ-YYYY-XXXXXXX-WW"
	got = trn.thisAcct

	if got != expect {
		t.Fatalf("wrong this account: expected==%v, got==%v\n", expect, got)
	}
}

func TestHappyTransactMini(t *testing.T) {
	t.Parallel()

	cfg := mini

	err := cfg.isValid()
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	flds := []string{"2025-04-17", "A penny for your thoughts.", ".01"}

	var trn transact

	err = trn.transact(flds, cfg)
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	expect := "2025-04-17,Mini,,A penny for your thoughts.,0.01"
	got := trn.string()

	if got != expect {
		t.Fatalf("wrong String(): expected==%q, got==%q\n", expect, got)
	}
}

func TestHappyTransactPCUCredit(t *testing.T) {
	t.Parallel()

	cfg := pcu

	err := cfg.isValid()
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	// test credit transaction using the credit field
	flds := []string{"28/11/2019", "HealthAndLif eInsuranceAn dSubs ARNHEMCR BP", "", "123.00", "316.69"}

	var trn transact

	err = trn.transact(flds, cfg)
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	expect := "2019-11-28,Assets:Current:PCUS1,,HealthAndLif eInsuranceAn dSubs ARNHEMCR BP,123"
	got := trn.string()

	if got != expect {
		t.Fatalf("wrong String(): expected==%q, got==%q\n", expect, got)
	}
}

func TestHappyTransactPCUDebit(t *testing.T) {
	t.Parallel()

	cfg := pcu

	// test debit transaction using the debit field
	flds := []string{"07/01/2020", "554PHP 18832946 Best of Health", "16.92", "", "265.01"}

	var trn transact

	err := trn.transact(flds, cfg)
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	expect := "2020-01-07,Assets:Current:PCUS1,,554PHP 18832946 Best of Health,-16.92"
	got := trn.string()

	if got != expect {
		t.Fatalf("wrong String(): expected==%q, got==%q\n", expect, got)
	}

	// test debit transaction with a negative debit value!
	flds = []string{"07/01/2020", "554PHP 18832946 Best of Health", "-16.92", "", "265.01"}

	err = trn.transact(flds, cfg)
	if err != nil {
		t.Fatalf("wrong error: expected==nil, got!=nil")
	}

	got = trn.string()
	if got != expect {
		t.Fatalf("wrong String(): expected==%q, got==%q\n", expect, got)
	}
}

func TestUnhappyConfigIndexes(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	// field index cannot be out of range
	cfg.amountI = cfg.nFields + 1

	err := cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}

	cfg = kbFull

	// field indexes cannot share a non-zero value
	cfg.creditI, cfg.debitI = 1, 1

	err = cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}
}

func TestUnhappyConfigMandatory(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	// date field index cannot be zero
	cfg.dateI = 0

	err := cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}

	cfg = kbFull

	// memo field index cannot be zero
	cfg.memoI = 0

	err = cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}

	cfg = kbFull

	// date format cannot be empty string
	cfg.dateFormat = ""

	err = cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}

	// date format must be a Go date format
	cfg.dateFormat = "gibberish"

	err = cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}
}

func TestUnhappyConfigNFields(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	// number of fields cannot be out of range, too low
	cfg.nFields = minNFields - 1

	err := cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}

	// number of fields cannot be out of range, too high
	cfg.nFields = maxNFields + 1

	err = cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}
}

func TestUnhappyConfigOptional(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	// this account and this account field index cannot be empty string and zero respectively
	cfg.thisAcct, cfg.thisAcctI = "", 0

	err := cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}

	cfg = kbFull

	// if amount field index is zero then both credit and debit indexes must be non-zero
	cfg.amountI, cfg.creditI, cfg.debitI = 0, 1, 0

	err = cfg.isValid()
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil\n")
	}
}

func TestUnhappyTransactAmount(t *testing.T) {
	t.Parallel()

	cfg := pcu

	var trn transact

	// either credit or debit must have a value not both
	flds := []string{"07/01/2020", "554PHP 18832946 Best of Health", "16.92", "-16.92", "265.01"}

	err := trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}

	// either credit or debit must have a value not neither
	flds = []string{"07/01/2020", "554PHP 18832946 Best of Health", "", "", "265.01"}

	err = trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}
}

func TestUnhappyTransactDate(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	// date field cannot have different format from date format
	flds := []string{"ZZ-YYYY-XXXXXXX-WW", "29/12/2023", "Automatic Payment Rates MISS E MACD ;Ref: Rates MISS E MACD",
		"AP", "Rates", "E", "", "", "", "", "MISS E MACD", "AA-BBBB-CCCCCCC-DD", "162.00", "", "162.00", "1434.23"}

	var trn transact

	err := trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}

	// date format cannot be gibberish!
	cfg.dateFormat = "gibberish"

	err = trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}
}

func TestUnhappyTransactMemo(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	flds := []string{"ZZ-YYYY-XXXXXXX-WW", "29-12-2023", "",
		"AP", "Rates", "E", "", "", "", "", "MISS E MACD", "AA-BBBB-CCCCCCC-DD", "162.00", "", "162.00", "1434.23"}

	var trn transact

	err := trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}
}

func TestUnhappyTransactNFields(t *testing.T) {
	t.Parallel()

	cfg := pcu

	flds := []string{"28/11/2019", "HealthAndLif eInsuranceAn dSubs ARNHEMCR BP", "123.00", "316.69"}

	var trn transact

	err := trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}
}

func TestUnhappyTransactThisAcct(t *testing.T) {
	t.Parallel()

	cfg := kbFull

	// as this account field index is non-zero, this account field cannot be empty string
	flds := []string{"", "29-12-2023", "Automatic Payment Rates MISS E MACD ;Ref: Rates MISS E MACD",
		"AP", "Rates", "E", "", "", "", "", "MISS E MACD", "AA-BBBB-CCCCCCC-DD", "162.00", "", "", "1434.23"}

	var trn transact

	err := trn.transact(flds, cfg)
	if err == nil {
		t.Fatalf("wrong error: expected!=nil, got==nil")
	}
}

var kbFull = config{ // for Kiwibank full CSV statement
	nFields: 16,
	amountI: 15, creditI: 13, dateI: 2, debitI: 14,
	memoI: 3, otherAcctI: 12, thisAcctI: 1,
	dateFormat: "02-01-2006", thisAcct: "",
}

var mini = config{ // for minimal CSV statement
	nFields: 3,
	amountI: 3, creditI: 0, dateI: 1, debitI: 0,
	memoI: 2, otherAcctI: 0, thisAcctI: 0,
	dateFormat: "2006-01-02", thisAcct: "Mini",
}

var pcu = config{ // for Police Credit Union account CSV statement
	nFields: 5,
	amountI: 0, creditI: 4, dateI: 1, debitI: 3,
	memoI: 2, otherAcctI: 0, thisAcctI: 0,
	dateFormat: "02/01/2006", thisAcct: "Assets:Current:PCUS1",
}
