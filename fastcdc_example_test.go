package fastcdc_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/jotfs/fastcdc-go"
)

func Example_basic() {
	data := make([]byte, 10*1024*1024)
	rand.Seed(4542)
	rand.Read(data)
	rd := bytes.NewReader(data)

	chunker, err := fastcdc.NewChunker(rd, fastcdc.Options{
		AverageSize: 1024 * 1024, // target 1 MiB average chunk size
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%-32s  %s\n", "CHECKSUM", "CHUNK SIZE")

	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%x  %d\n", md5.Sum(chunk.Data), chunk.Length)
	}

	// Output:
	// CHECKSUM                          CHUNK SIZE
	// 42efc2d41f145f1523b613b5e5c244c6  1056261
	// 7f8a30eda030a8723a207a5c467db177  1391596
	// 36a604a2eb2dc5a13caf55689e9b820d  1192712
	// 9d08132b57b7c4c7c0cfdd326936a416  1071442
	// d989c788b682423912a7df97b6805d2f  1230898
	// 2ca33ba7b0ba2dab2dfeca01a1daa925  1107414
	// 5c2120e71cd568c02c4c06cb4fe46919  1197108
	// 0cfc83a7674298ac90e0e0ed7a045699  1521253
	// 9dcdf5bb1fbee15d8d86fd21a0f2c3d9  717076
}
