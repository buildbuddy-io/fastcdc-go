package v2020

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"testing"
)

// TestCutSekien16kChunks matches https://github.com/nlfiedler/fastcdc-rs/blob/master/src/v2020/mod.rs#L903
func TestCutSekien16kChunks(t *testing.T) {
	data, err := os.ReadFile("../testdata/SekienAkashita.jpg")
	if err != nil {
		t.Skipf("test file not found: %v", err)
	}

	chunker, err := NewChunker(bytes.NewReader(data), Options{
		AverageSize:   16384,
		MinSize:       4096,
		MaxSize:       65535,
		Normalization: 1,
	})
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	expected := []struct {
		hash   uint64
		length int
	}{
		{17968276318003433923, 21325},
		{8197189939299398838, 17140},
		{13019990849178155730, 28084},
		{4509236223063678303, 18217},
		{2504464741100432583, 24700},
	}

	var chunks []struct {
		hash   uint64
		length int
	}

	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("error reading chunk: %v", err)
		}
		chunks = append(chunks, struct {
			hash   uint64
			length int
		}{chunk.Fingerprint, chunk.Length})
	}

	if len(chunks) != len(expected) {
		t.Errorf("expected %d chunks, got %d", len(expected), len(chunks))
		for i, c := range chunks {
			t.Logf("  %d: length=%d hash=%d", i, c.length, c.hash)
		}
		return
	}

	for i, e := range expected {
		if chunks[i].length != e.length {
			t.Errorf("chunk %d: expected length %d, got %d", i, e.length, chunks[i].length)
		}
		if chunks[i].hash != e.hash {
			t.Errorf("chunk %d: expected hash %d, got %d", i, e.hash, chunks[i].hash)
		}
	}
}

// TestCutSekien16kChunksSeed666 matches https://github.com/nlfiedler/fastcdc-rs/blob/master/src/v2020/mod.rs#L928
func TestCutSekien16kChunksSeed666(t *testing.T) {
	data, err := os.ReadFile("../testdata/SekienAkashita.jpg")
	if err != nil {
		t.Skipf("test file not found: %v", err)
	}

	chunker, err := NewChunker(bytes.NewReader(data), Options{
		AverageSize:   16384,
		MinSize:       4096,
		MaxSize:       65535,
		Normalization: 1,
		Seed:          666,
	})
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	expected := []struct {
		hash   uint64
		length int
	}{
		{9312357714466240148, 10605},
		{226910853333574584, 55745},
		{12271755243986371352, 11346},
		{14153975939352546047, 5883},
		{5890158701071314778, 11586},
		{8981594897574481255, 14301},
	}

	var chunks []struct {
		hash   uint64
		length int
	}

	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("error reading chunk: %v", err)
		}
		chunks = append(chunks, struct {
			hash   uint64
			length int
		}{chunk.Fingerprint, chunk.Length})
	}

	if len(chunks) != len(expected) {
		t.Errorf("expected %d chunks, got %d", len(expected), len(chunks))
		for i, c := range chunks {
			t.Logf("  %d: length=%d hash=%d", i, c.length, c.hash)
		}
		return
	}

	for i, e := range expected {
		if chunks[i].length != e.length {
			t.Errorf("chunk %d: expected length %d, got %d", i, e.length, chunks[i].length)
		}
		if chunks[i].hash != e.hash {
			t.Errorf("chunk %d: expected hash %d, got %d", i, e.hash, chunks[i].hash)
		}
	}
}

func TestChunkingRandom(t *testing.T) {
	data := randBytes(1e6, 63)
	chunker, err := NewChunker(bytes.NewReader(data), Options{
		AverageSize: 1024,
	})
	if err != nil {
		t.Fatal(err)
	}

	var prevOffset int
	var prevLength int
	allData := make([]byte, 0)
	for i := 0; ; i++ {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		offset := prevOffset + prevLength
		if offset != chunk.Offset {
			t.Errorf("chunk %d: Offset should be %d not %d", i, offset, chunk.Offset)
		}
		if chunk.Length != len(chunk.Data) {
			t.Errorf("chunk %d: Length %d does not match len(Data) %d", i, chunk.Length, len(chunk.Data))
		}

		allData = append(allData, chunk.Data...)

		prevOffset = chunk.Offset
		prevLength = chunk.Length
	}
	if !bytes.Equal(allData, data) {
		t.Error("data does not match")
	}
}

func TestMinSize(t *testing.T) {
	data := randBytes(10, 51)
	chunker, err := NewChunker(bytes.NewReader(data), Options{
		AverageSize:          1024,
		DisableNormalization: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err := chunker.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, c.Data) {
		t.Error("data not equal")
	}
	if c.Length != len(data) {
		t.Errorf("invalid length %d", c.Length)
	}

	_, err = chunker.Next()
	if err != io.EOF {
		t.Error("expected io.EOF error")
	}
}

func TestCutAllZeros(t *testing.T) {
	data := make([]byte, 10240)

	chunker, err := NewChunker(bytes.NewReader(data), Options{
		AverageSize:   256,
		MinSize:       64,
		MaxSize:       1024,
		Normalization: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	var totalLength int
	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if chunk.Length != 1024 {
			t.Errorf("expected chunk length 1024 for all zeros, got %d", chunk.Length)
		}
		totalLength += chunk.Length
	}

	if totalLength != 10240 {
		t.Errorf("expected total length 10240, got %d", totalLength)
	}
}

type benchSpec struct {
	size int
	name string
}

var benchSizes = []benchSpec{
	{1 << 10, "1k"},
	{4 << 10, "4k"},
	{16 << 10, "16k"},
	{32 << 10, "32k"},
	{64 << 10, "64k"},
	{128 << 10, "128k"},
	{256 << 10, "256k"},
	{512 << 10, "512k"},
	{1 << 20, "1M"},
	{4 << 20, "4M"},
	{16 << 20, "16M"},
	{32 << 20, "32M"},
	{64 << 20, "64M"},
	{128 << 20, "128M"},
	{512 << 20, "512M"},
	{1 << 30, "1G"},
}

func BenchmarkFastCDCSize(b *testing.B) {
	for _, s := range benchSizes {
		s := s
		b.Run(s.name, func(b *testing.B) {
			benchmarkFastCDCSize(b, s.size)
		})
	}
}

func benchmarkFastCDCSize(b *testing.B, size int) {
	rng := rand.New(rand.NewSource(1))
	data := make([]byte, size)
	rng.Read(data)

	r := bytes.NewReader(data)
	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	cnkr, err := NewChunker(r, Options{
		AverageSize: 1 * miB,
	})
	if err != nil {
		b.Fatal(err)
	}

	var res uint64
	var nchks int64

	for i := 0; i < b.N; i++ {
		r.Reset(data)
		cnkr.Reset(r)

		for {
			chunk, err := cnkr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
			res = res + uint64(len(chunk.Data))
			nchks++
		}
	}
	b.ReportMetric(float64(nchks)/float64(b.N), "chunks")
}

func randBytes(n int, seed int64) []byte {
	b := make([]byte, n)
	rnd := rand.New(rand.NewSource(seed))
	rnd.Read(b)
	return b
}
