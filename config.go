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
)

const (
	// The inclusive limits for the number of fields in an input CSV record.
	minNFields = 3 // date, memo and amount
	maxNFields = 20
	nIndexes   = 7 // number of field indexes in config
)

/*
A config configures cbas2trn.
Each field in the configuration is either mandatory or optional.
An optional field can have a zero value, but a mandatory field cannot.
*/
type config struct {
	// NFields is the number of fields in an input CSV record, and it is mandatory.
	nFields uint8
	/*
		The indexes of fields in an input CSV record.
		If an index is zero, this record does not contain that field.
	*/
	amountI    uint8 // optional, but if zero then creditI and debitI must be non-zero
	creditI    uint8 // optional, see amountI
	dateI      uint8 // mandatory
	debitI     uint8 // optional, see amountI
	memoI      uint8 // or description, mandatory
	otherAcctI uint8 // optional
	thisAcctI  uint8 // optional, see thisAcct
	/*
		DateFormat is the format of the date field in an input CSV record.
		It is mandatory and Go style e.g. "02/01/2006"
	*/
	dateFormat string
	/*
		ThisAcct is the name of the account that the input CSV record belongs to.
		It is optional, but if it is empty string then thisAcctI must be non-zero.
	*/
	thisAcct string
}

var (
	errAmountOpt    = errors.New("amount field index, or credit and debit indexes cannot both be zero")
	errDateI        = errors.New("date field index cannot be zero")
	errDateFormat   = errors.New("date format in input CSV record must be Go style e.g. \"02/01/2006\"")
	errIndexUnique  = errors.New("field indexes cannot share a non-zero value")
	errIndexRange   = errors.New("field index is out of range")
	errMemoI        = errors.New("memo field index cannot be zero")
	errNFieldsRange = errors.New("number of fields in input CSV record is out of range")
	errThisAcctOpt  = errors.New("this account and this account index " +
		"cannot be empty string and zero respectively")
)

/*
AreIndexesValid returns nil if all field indexes are valid.
It assumes the number of fields in an input CSV record nFields is in range.
All indexes must be <= nFields, and all non-zero indexes must be unique.
If not, areIndexesValid returns the first error.
*/
func (cfg *config) areIndexesValid() error {
	inxs := [nIndexes]uint8{cfg.amountI, cfg.creditI, cfg.dateI, cfg.debitI, cfg.memoI, cfg.otherAcctI, cfg.thisAcctI}

	var inUse [maxNFields + 1]bool

	for _, val := range inxs {
		switch {
		case cfg.nFields < val:
			return errIndexRange
		case val == 0:
			// input CSV record does not contain this field
		case inUse[val]:
			return errIndexUnique
		default:
			inUse[val] = true
		}
	}

	return nil
}

/*
AreOptionsValid returns nil if the combination of options are valid.
If not, areOptionsValid returns the first error.
*/
func (cfg *config) areOptionsValid() error {
	if cfg.thisAcct == "" && cfg.thisAcctI == 0 {
		return errThisAcctOpt
	}

	if (cfg.amountI == 0) && (cfg.creditI == 0 || cfg.debitI == 0) {
		return errAmountOpt
	}

	return nil
}

/*
IsValid returns nil if this configuration is valid.
If not, isValid returns the first error.
*/
func (cfg *config) isValid() error {
	if cfg.dateFormat == "" {
		// The date format should be validated here, but how?
		return errDateFormat
	}

	if cfg.nFields < minNFields || maxNFields < cfg.nFields {
		return errNFieldsRange
	}

	err := cfg.areIndexesValid()
	if err != nil {
		return err
	}

	if cfg.dateI == 0 {
		return errDateI
	}

	if cfg.memoI == 0 {
		return errMemoI
	}

	err = cfg.areOptionsValid()
	if err != nil {
		return err
	}

	return nil
}
