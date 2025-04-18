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
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

/*
A transact represents a financial transaction in CSV format.
All but one of the fields in a transaction are mandatory, and must be non-zero.
*/
type transact struct {
	amount    float64
	date      string
	memo      string
	otherAcct string // optional, can be empty string
	thisAcct  string
}

const zero = 0.00

var (
	errAmount   = errors.New("amount cannot be zero")
	errMemo     = errors.New("memo cannot be empty string")
	errNFields  = errors.New("wrong number of fields")
	errThisAcct = errors.New("this account cannot be empty string")
)

/*
ParseAmount returns the amount of this transaction and nil.
It assumes the configuration is valid.
The amount is parsed from:
 1. amount field or
 2. credit field if debit field is an empty string or
 3. debit field if credit field is an empty string

If it fails to parse a value, parseAmount returns an error.
*/
func parseAmount(fields []string, cfg config) (float64, error) {
	if cfg.amountI != 0 {
		return parseFloat64(fields[cfg.amountI])
	}

	crt, dbt := fields[cfg.creditI], fields[cfg.debitI]
	if dbt == "" {
		return parseFloat64(crt)
	}

	amt, err := parseFloat64(dbt)
	if err != nil {
		return amt, err
	}

	const minus1 = -1.00

	amt = math.Abs(amt) * minus1

	return amt, nil
}

/*
ParseDate returns the date of this transaction and nil.
It assumes the configuration and its date format are valid.
If it fails to parse a date, parseDate returns an error.
*/
func parseDate(fields []string, cfg config) (string, error) {
	val, err := time.Parse(cfg.dateFormat, fields[cfg.dateI])
	if err != nil {
		return "", fmt.Errorf("parseDate: %w", err)
	}

	return val.Format(time.DateOnly), nil
}

/*
ParseFloat64 returns the float64 value parsed from the string and nil.
If it fails to parse a value, parseFloat64 returns an error.
*/
func parseFloat64(float string) (float64, error) {
	val, err := strconv.ParseFloat(float, 64)
	if err != nil {
		return zero, fmt.Errorf("parseFloat64: %w", err)
	}

	return val, nil
}

// String returns the transaction in the standard CSV format.
func (trn *transact) string() string {
	amt := strconv.FormatFloat(trn.amount, 'f', -1, 64)
	flds := []string{trn.date, trn.thisAcct, trn.otherAcct, trn.memo, amt}

	const sep = ","

	return strings.Join(flds, sep)
}

/*
Transact parses the transaction from the fields, according to the configuration, and returns nil.
It assumes the configuration and its date format are valid.
If transact fails to parse a transaction, it returns the first error.
*/
func (trn *transact) transact(fields []string, cfg config) error {
	if len(fields) != int(cfg.nFields) {
		return errNFields
	}

	/*
		Prepend fields with an empty string.
		The value of an optional field, whose field index is zero,
		will then be empty string.
	*/
	flds := slices.Insert(fields, 0, "")

	var err error

	trn.date, err = parseDate(flds, cfg)
	if err != nil {
		return err
	}

	trn.amount, err = parseAmount(flds, cfg)
	if err != nil {
		return err
	} else if trn.amount == zero {
		return errAmount
	}

	trn.memo = flds[cfg.memoI]
	if trn.memo == "" {
		return errMemo
	}

	trn.otherAcct = flds[cfg.otherAcctI]

	switch {
	case cfg.thisAcct != "":
		trn.thisAcct = cfg.thisAcct
	case flds[cfg.thisAcctI] != "":
		trn.thisAcct = flds[cfg.thisAcctI]
	default:
		return errThisAcct
	}

	return nil
}
