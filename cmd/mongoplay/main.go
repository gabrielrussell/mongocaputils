package main

import (
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"

	"github.com/gabrielrussell/mongocaputils"
	"github.com/gabrielrussell/mongocaputils/mongoproto"
)

var (
	messageFile   = flag.String("f", "-", "message file (or '-' for stdin)")
	packetBufSize = flag.Int("size", 1000, "size of packet buffer used for ordering within streams")
	verbose       = flag.Bool("v", false, "verbose output (to stderr)")
	host          = flag.String("h", "127.0.0.1:27017", "mongod host (127.0.0.1:27017)")
)

func newPlayOpChan(fileName string) (<-chan *mongocaputils.OpWithTime, error) {
	opFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	ch := make(chan *mongocaputils.OpWithTime)
	go func() {
		defer close(ch)
		for {
			buf, err := mongoproto.ReadDocument(opFile)
			if err != nil {
				fmt.Printf("ReadDocument: %v\n", err)
				if err == io.EOF {
					return
				}
				os.Exit(1)
			}
			var doc mongocaputils.OpWithTime
			err = bson.Unmarshal(buf, &doc)
			if err != nil {
				fmt.Printf("Unmarshal: %v\n", err)
				os.Exit(1)
			}
			ch <- &doc

		}
	}()
	return ch, nil
}

func newOpConnection(url string, realReplyChan chan<- int32) (chan<- *mongocaputils.OpWithTime, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	ch := make(chan *mongocaputils.OpWithTime)
	go func() {
		for op := range ch {
			err = op.Execute(session, realReplyChan)
			if err != nil {
				fmt.Printf("op.Execute %v", err)
			}
		}
	}()
	return ch, nil
}

type realReply struct {
	numberReturned int32
}

func newReplyManager() (chan<- *mongocaputils.OpWithTime, chan<- int32) {
	recordedReplyChan := make(chan *mongocaputils.OpWithTime)
	realReplyChan := make(chan int32)
	go func() {
		for {
			select {
			case <-realReplyChan:
			case <-recordedReplyChan:
				//XXX also pipe in that we expect a reply
			}
		}
	}()
	return recordedReplyChan, realReplyChan //XXX also pipe in that we expect a reply
}

func main() {
	flag.Parse()
	opChan, err := newPlayOpChan(*messageFile)
	if err != nil {
		fmt.Printf("newPlayOpChan: %v\n", err)
		os.Exit(1)
	}
	recordedReplyChan, realReplyChan := newReplyManager()
	sessions := make(map[string]chan<- *mongocaputils.OpWithTime)
	for op := range opChan {
		fmt.Printf("%v\n\n", op)
		if op.OpCode() == 1 {
			recordedReplyChan <- op
		} else {
			session, ok := sessions[op.Connection]
			if !ok {
				session, err = newOpConnection(*host, realReplyChan)
				if err != nil {
					fmt.Printf("newOpConnection: %v\n", err)
					os.Exit(1)
				}
				sessions[op.Connection] = session
			}
			session <- op
		}
	}
}
