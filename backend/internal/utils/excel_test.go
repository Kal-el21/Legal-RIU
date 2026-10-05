package utils

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadCSV(t *testing.T) {
	want := [][]string{{"code", "name"}, {"D01", "Divisi Legal, Kepatuhan"}}

	cases := map[string]string{
		"koma":            "code,name\nD01,\"Divisi Legal, Kepatuhan\"\n",
		"titik koma":      "code;name\r\nD01;Divisi Legal, Kepatuhan\r\n",
		"BOM UTF-8":       "\xEF\xBB\xBFcode,name\nD01,\"Divisi Legal, Kepatuhan\"\n",
		"BOM, titik koma": "\xEF\xBB\xBFcode;name\nD01;Divisi Legal, Kepatuhan\n",
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := readCSV(strings.NewReader(input))
			if err != nil {
				t.Fatalf("readCSV error: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
