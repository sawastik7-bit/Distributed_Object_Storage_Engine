package chunker

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

const ChunkSize = 500 // 

// describes one piece of a file
type ChunkMeta struct {
	ID       string
	Index    int
	Size     int
	CheckSum string
}

// the chunk pairs the metadata with the actual bytes
type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

func Split(r io.Reader) ([]Chunk, error) {
	var chunks []Chunk
	buf := make([]byte, ChunkSize)
	index := 0

	for {
		n, err := io.ReadFull(r, buf)

		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])

			sum := sha256.Sum256(data)
			checksum := hex.EncodeToString(sum[:])

			chunks = append(chunks, Chunk{
				Meta: ChunkMeta{
					ID:       checksum,
					Index:    index,
					Size:     n,
					CheckSum: checksum,
				},
				Data: data,
			})
			index++
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("chunker: read failed: %w", err)
		}
	}

	return chunks, nil
}

// Verify recomputes the checksum of data and compares it to expected.
// Used on the download path to catch corrupted chunks.
func Verify(data []byte, expectedChecksum string) bool {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) == expectedChecksum
}