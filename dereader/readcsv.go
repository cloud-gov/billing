package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
)

func ReadCSV() {
	f, err := os.Open("./dereader/03/23/cg-billing-export-cur-hourly-csv-00008.csv")
	check(err, ErrFileOpen)
	defer (func() { check(f.Close(), ErrFileClose) })()

	r := csv.NewReader(f)
	r.ReuseRecord = true

	// "line_item_usage_type"
	// "line_item_usage_end_date"
	// "line_item_usage_start_date"
	// "tags" JSON -> "resourceTags/user:Instance GUID"
	// "pricing_unit" "GB-Mo"
	// "line_item_usage_amount"
	ks := map[string]int{
		"tags":                       -1,
		"pricing_unit":               -1,
		"line_item_usage_type":       -1,
		"line_item_usage_amount":     -1,
		"line_item_usage_end_date":   -1,
		"line_item_usage_start_date": -1,
	}

	names, err := r.Read()
	check(err, ErrReading)
	names = slices.Clone(names)

	kks := slices.Collect(maps.Keys(ks))
	uns := len(kks)
	for i, n := range names {
		if slices.Contains(kks, n) {
			ks[n] = i
			uns--
		}
		if uns < 1 {
			break
		}
	}

	// now := time.Now().UTC()

	// lazy timeout
	for range int(1e8) {
		l, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		check(err, ErrReading)

		// end, err := time.Parse(time.RFC3339, l[ks["line_item_usage_end_date"]])
		// check(err)

		ut := l[ks["line_item_usage_type"]]
		if strings.Contains(ut, "TimedStorage-ByteHrs") {
			m := make(map[string]string, len(l))
			for i, n := range names {
				m[n] = l[i]
			}

			j, err := json.Marshal(m)
			check(err)

			fmt.Println(string(j))
		}
	}
}
