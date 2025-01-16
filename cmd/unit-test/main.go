package main

/*
* Not really a unit-test yet..
 */
import (
	"fmt"

	logwriter "github.com/code-ointment/log-writer"
	"github.com/code-ointment/log-writer/log_file"
)

func main() {

	lw := log_file.NewLogFileWriter("./tmp/lw.log", 5, 1024)

	for i := 0; i < 1024; i++ {
		msg := fmt.Sprintf("Message number %d\n", i)
		lw.Write([]byte(msg))
	}
	logwriter.Flush()
}
