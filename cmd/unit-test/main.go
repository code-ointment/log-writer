package main

/*
* Not really a unit-test yet..
 */
import (
	"fmt"

	logwriter "github.com/code-ointment/log-writer"
	logfile "github.com/code-ointment/log-writer/logfile"
)

func main() {

	//lw := logfile.NewLogFileWriter("./tmp/lw.log", 5, 1024)
	lw := logfile.NewLogFileWriter("./tmp/lw.log", 5, 1024)
	for i := 0; i < 1024; i++ {
		msg := fmt.Sprintf("Message number %d\n", i)
		lw.Write([]byte(msg))
	}
	logwriter.Flush()
}
