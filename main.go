package main

// convert go test report JSON into chrome tracing JSON

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"
)

type testevent struct {
	Time    *time.Time `json:",omitempty"`
	Action  string
	Package string   `json:",omitempty"`
	Test    string   `json:",omitempty"`
	Elapsed *float64 `json:",omitempty"`
	Output  string   `json:",omitempty"`
}

type traceevent struct {
	Ts   int64  `json:"ts"`
	Pid  string `json:"pid"`
	Tid  string `json:"tid"`
	Ph   string `json:"ph"`
	Name string `json:"name,omitempty"`
	Cat  string `json:"cat,omitempty"`
}

func main() {
	pkg := make(map[string][]string)
	start := make(map[string]*time.Time)
	end := make(map[string]*time.Time)
	result := make(map[string]string)
	min := time.Now()

	var tr io.Reader
	if len(os.Args) > 1 {
		report := os.Args[1]
		f, err := os.Open(report)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		tr = f
	} else {
		tr = os.Stdin
	}
	dec := json.NewDecoder(tr)

	for {
		var e testevent
		if err := dec.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return
		}

		action := e.Action
		switch action {
		case "output":
			continue
		case "run":
			test := e.Test
			start[test] = e.Time
			if e.Time.Before(min) {
				min = *e.Time
			}
			p := e.Package
			if p == "" {
				p = "main"
			}
			pkg[p] = append(pkg[p], test)
		case "pass", "fail", "skip":
			test := e.Test
			if test == "" {
				continue
			}
			end[test] = e.Time
			if _, ok := start[test]; !ok {
				panic("missing start")
			}
			result[test] = action
		case "pause", "cont":
			continue
		default:
			fmt.Printf("Unknown action: %s", action)
			continue
		}
	}

	events := []traceevent{}
	for p, tests := range pkg {
		sort.Strings(tests)
                sort.SliceStable(tests, func(i, j int) bool {
			return start[tests[i]].Before(*start[tests[j]])
		})
		for _, test := range tests {
			res := result[test]
			if res == "" {
				continue
			}
			//duration := end[test].Sub(*start[test])
			//fmt.Printf("%s %s %s %s %s\n", p, test, start[test], end[test], duration)
			b := start[test].Sub(min).Microseconds()
			e := end[test].Sub(min).Microseconds()
			events = append(events, traceevent{Ts: b, Pid: p, Tid: test, Ph: "B", Name: test, Cat: res})
			events = append(events, traceevent{Ts: e, Pid: p, Tid: test, Ph: "E"})
		}
	}

	enc := json.NewEncoder(os.Stdout)
	doc := map[string]interface{}{"displayTimeUnit": "ms", "traceEvents": events}
	if err := enc.Encode(&doc); err != nil {
		log.Println(err)
	}
}
