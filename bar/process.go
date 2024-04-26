package bar

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

func Process(r io.Reader) chan Bar {
	updates := make(chan Bar)

	go func() {
		s := bufio.NewScanner(r)
		s.Scan()
		s.Scan()
		for s.Scan() {
			bar := Bar{}
			err := json.Unmarshal(s.Bytes()[1:], &bar)
			if err != nil {
				log.Fatalf("error reading from bar: %v", err)
			}
			bar.Sort()
			updates <- bar
		}
	}()

	return updates
}
