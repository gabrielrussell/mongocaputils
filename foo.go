package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gabrielrussell/mgo"
	"github.com/gabrielrussell/mgo/bson"
)

func main() {
	mgo.SetDebug(true)
	mgo.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
	session, err := mgo.Dial("localhost")
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	var r bson.M
	db := session.DB("test")

	//	err = db.C("bar").Find(nil).One(&r)
	//	if err != nil {
	//		fmt.Printf("%v", err)
	//		os.Exit(1)
	//	}
	//	fmt.Printf("%#v\n", r)

	data, reply, err := db.QueryOp(&mgo.QueryOp{Collection: "test.bar", Limit: 1})
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%v, %#v\n", data, reply)
	for _, d := range data {
		err = bson.Unmarshal(d, &r)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(1)
		}
		fmt.Printf("%#v\n", r)
	}
}
