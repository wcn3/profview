package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/wcn3/profview/perftools_profiles"

	"github.com/golang/protobuf/proto"
)

const (
	overhead = "Waiting"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s profile_file", os.Args[0])
		os.Exit(1)
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Couldn't read file %s: %v", os.Args[1], err)
	}
	msg := perftools_profiles.Profile{}
	if err := proto.Unmarshal(b, &msg); err != nil {
		// Maybe it's gzipped?
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			log.Fatalf("Couldn't create gzip reader: %v", err)
		}
		b, err = ioutil.ReadAll(r)
		if err != nil {
			log.Fatalf("Couldn't read all from memory: %v", err)
		}
		if err = proto.Unmarshal(b, &msg); err != nil {
			log.Fatalf("Couldn't unmarshal proto message: %v", err)
		}
	}

	// Attribute the slack time to overhead
	tot := int64(10e6 * len(msg.GetSample()))
	odur := msg.DurationNanos - tot

	// In order to create a synthetic entry, we need to create the following
	// entities.

	// A string table entry to which to refer.
	msg.StringTable = append(msg.StringTable, overhead)

	// A Mapping that points to the string table.
	mid := uint64(len(msg.GetMapping()) + 1)
	mapping := perftools_profiles.Mapping{
		Id:       mid,
		Filename: int64(len(msg.GetStringTable()) - 1),
	}
	msg.Mapping = append(msg.GetMapping(), &mapping)

	// A Location that refers to the Mapping.
	lid := uint64(len(msg.GetLocation()) + 1)
	loc := perftools_profiles.Location{
		Id:        lid,
		MappingId: mid,
	}
	msg.Location = append(msg.GetLocation(), &loc)

	// And a sample that has the slack time referring to the Location
	// created.
	samp := perftools_profiles.Sample{
		LocationId: []uint64{lid},
		Value:      []int64{odur / 1e9, odur},
	}
	msg.Sample = append(msg.GetSample(), &samp)

	// Write out the updated message.
	var buf bytes.Buffer

	w := gzip.NewWriter(&buf)

	up, err := proto.Marshal(&msg)
	if err != nil {
		log.Fatalf("Couldn't marshal proto: %v", err)
	}

	_, err = w.Write(up)
	if err != nil {
		log.Fatalf("Couldn't write proto: %v", err)
	}
	w.Close()

	err = ioutil.WriteFile(os.Args[2], buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Couldn't write file to disk: %v", err)
	}
}
