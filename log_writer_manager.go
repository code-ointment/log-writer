package logwriter

/*
* LogWriters register with the manager module on create.
* Calling flush causes the caller to wait until any LogWriter work is done.
 */
import (
	"sync"
)

var managerLock sync.Mutex
var writers []LogWriterInterface

/*
* Used by log writer interface implementors
 */
func Register(intf LogWriterInterface) {

	managerLock.Lock()
	defer managerLock.Unlock()

	writers = append(writers, intf)
}

/*
* Wait for writer to finish.  Typically this means waiting on gzipping to
* complete.
 */
func Flush() {
	for _, w := range writers {
		w.Close()
	}
}
