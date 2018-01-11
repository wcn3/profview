package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/wcn3/profview/perftools_profiles"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s profile_file", os.Args[0])
		os.Exit(1)
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Couldn't read file %s: %v", os.Args[1], err)
	}
	msg := perftools_profiles.Profile{}
	if err := proto.Unmarshal(b, &msg); err != nil {
		log.Fatalf("Couldn't unmarshal profile: %v", err)
	}
	spew.Dump(msg)
}
