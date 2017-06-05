package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var nonSpaceSeqs = regexp.MustCompile("[^\\s]+")

func main() {
	reader := bufio.NewReader(os.Stdin)

	rowIndex := 0
	for {
		rowIndex++
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// end of input
				return
			}
			panic(err)
		}

		// replace +/- character
		line = strings.Replace(line, "±", "", -1)

		fields := nonSpaceSeqs.FindAllString(line, 3)

		// remove test name
		name := strings.Split(fields[0], "/")[1]

		// shorter forms
		name = strings.Replace(name, "Websockets", "WS", -1)
		name = strings.Replace(name, "binary_", "", -1)
		name = strings.Replace(name, "_clients_suite", "", -1)
		name = strings.Replace(name, "-4", "", -1)
		name = strings.Replace(name, "_concurrency-safe", "\\nconc.", -1)
		name = strings.Replace(name, "_text", "\\nJSON", -1)

		name = strings.Replace(name, "_", " ", -1)

		// parse measure time (Y)
		value, err := time.ParseDuration(fields[1])
		if err != nil {
			panic(err)
		}
		
		// parse standard deviation (Y error)
		pcentS := fields[2]
		pcent, err := strconv.ParseInt(pcentS[:len(pcentS)-1], 10, 64)
		if err != nil {
			panic(err)
		}

		// convert percentage to flat value
		pcentNorm := time.Duration(time.Nanosecond * time.Duration(uint(float64(value.Nanoseconds()*pcent)/float64(100))))
		
		const simulatedWorkDelay = time.Millisecond * 500
		
		// remove offset from both
		pcentNorm -= simulatedWorkDelay
		value -= simulatedWorkDelay

		// re-format output for gnuplot
		fmt.Println(rowIndex, `"`+name+`"`, value.Seconds()*1000, pcentNorm.Seconds()*1000)
	}
}
