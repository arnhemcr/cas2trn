/*
Copyright (C) 2025 Andrew Flint.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

/*
Cas2trn translates financial transactions from an arbitrary comma-separated values (CSV) format to the standard format.
The program's name stands for CSV account statement to transactions,
and it allows transactions from statements in different formats to be combined.
For more information see:

	cas2trn -help
*/
package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

const pgmName = "cas2trn" // see also pgmTitle

// Main runs cas2trn.
func main() {
	log.SetPrefix(pgmName + ": ")
	log.SetFlags(0)

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	var rdr *csv.Reader

	if 0 < flag.NArg() {
		for _, stmt := range flag.Args() {
			var file *os.File

			file, err = os.Open(stmt)
			if err != nil {
				log.Fatal(err)
			}

			rdr = csv.NewReader(file)
			err = translateStatement(rdr, cfg)
		}
	} else {
		rdr = csv.NewReader(os.Stdin)
		err = translateStatement(rdr, cfg)
	}

	if err != nil {
		log.Fatal(err)
	}
}

/*
Parseconfig returns the configuration for cas2trn and nil.
The configuration is parsed from flags.
If the configuration is not valid, parseConfig returns the first error.
*/
func parseConfig() (config, error) {
	flag.Usage = usage

	var help bool

	flag.BoolVar(&help, "help", false, "write this help text then exit")

	var cfg config

	var nFlds uint

	flag.UintVar(&nFlds, "nfields", 0, "number of fields in a CSV record, mandatory")

	var vals [nIndexes]uint

	flag.UintVar(&vals[0], "amounti", 0, "amount field index, "+
		"optional but if zero then crediti and debiti must be non-zero")
	flag.UintVar(&vals[1], "crediti", 0, "credit field index, optional see amounti")
	flag.UintVar(&vals[2], "datei", 0, "date field index, mandatory")
	flag.UintVar(&vals[3], "debiti", 0, "debit field index, optional see amounti")
	flag.UintVar(&vals[4], "memoi", 0, "memo or description field index, mandatory")
	flag.UintVar(&vals[5], "otheraccti", 0, "other account number or name field index, optional")
	flag.UintVar(&vals[6], "thisaccti", 0, "this account number or name field index, optional see thisacct")

	flag.StringVar(&cfg.dateFormat, "dateformat", "", "date format, mandatory and Go style e.g. \"02/01/2006\"")
	flag.StringVar(&cfg.thisAcct, "thisacct", "", "this account number or name, "+
		"optional but if empty string then thisaccti must be non-zero")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	cfg.nFields, cfg.amountI = ui2ui8(nFlds), ui2ui8(vals[0])
	cfg.creditI, cfg.dateI = ui2ui8(vals[1]), ui2ui8(vals[2])
	cfg.debitI, cfg.memoI = ui2ui8(vals[3]), ui2ui8(vals[4])
	cfg.otherAcctI, cfg.thisAcctI = ui2ui8(vals[5]), ui2ui8(vals[6])

	err := cfg.isValid()
	if err != nil {
		return cfg, fmt.Errorf("cfg.isValid: %w", err)
	}

	return cfg, nil
}

/*
TranslateStatement translates financial transactions in an account statement
from an arbitrary CSV format to the standard format and returns nil.
It reads each transaction, and parses it according to the cas2trn ration.
If it fails to read the statement, translateStatement returns an error.
If it fails to parse a transaction,
translateStatement writes an error to standard error and continues.
If it successfully parses a transaction,
translateStatement writes it in the standard format to standard output and continues.
*/
func translateStatement(reader *csv.Reader, cfg config) error {
	// Disable number of fields per record check; it is done in transact.transact() instead.
	reader.FieldsPerRecord = -1

	for {
		flds, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return fmt.Errorf("reader.Read(): %w", err)
		}

		var trn transact

		err = trn.transact(flds, cfg)
		if err != nil {
			lineN, _ := reader.FieldPos(0)
			fmt.Fprintln(os.Stderr,
				fmt.Errorf("%v: %w on line %v", pgmName, err, lineN))

			continue
		}

		fmt.Fprintln(os.Stdout, trn.string())
	}
}

/*
Ui2ui8 returns a value converted from uint to uint8.
If value is too large for a uint8, ui2ui8 returns zero.
*/
func ui2ui8(value uint) uint8 {
	if value <= math.MaxUint8 {
		return uint8(value)
	}

	return 0
}

// Prints usage for cas2trn.
func usage() {
	const pgmTitle = "Cas2trn"

	fmt.Fprintf(os.Stderr, "usage: %v [flags] [file names]\n", pgmName)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%v %v\n", pgmTitle,
		"translates financial transactions from an arbitrary comma-separated values (CSV) format to the standard format.")
	fmt.Fprintf(os.Stderr,
		`The program's name stands for CSV account statement to transactions, 
and it allows transactions from statements in different formats to be combined.
If the names of statement files are not given, cas2trn reads transactions from standard input.

The standard transaction format, written as a CSV record to standard output, contains the following fields:
 * date in ISO 8601 format, which is sortable, e.g. "2006-01-02"
 * this account number or name
 * other account number or name, optional
 * memo or description
 * amount

Parsing the arbitrary input transaction format is configured by flags.
Fields in the CSV records are linked to those in transactions by field indexes.
An index of zero means these records do not contain that field.
The flags are:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
For example, consider input transaction "24/12/2019,Brumby's,6.50,,330.04". 
It does not contain this account. 
It has debit and credit fields, instead of an amount, which are followed by a balance.
The flags to translate this transaction would be
"-thisacct=PCUS1 -nfields=5 -datei=1 -dateformat=02/01/2006 -memoi=2 -debiti=3 -crediti=4",
which would output "2019-12-24,PCUS1,,Brumby's,-6.5".

If cas2trn fails to parse a CSV record as a transaction, it prints an error on standard error then continues.
Errors about failing to parse header lines can be ignored.
`)
}
