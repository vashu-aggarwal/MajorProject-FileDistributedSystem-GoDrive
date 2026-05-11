package pipeline

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"errors"
)

// huffmanNode is a single node in the Huffman tree.
type huffmanNode struct {
	symbol    byte
	frequency int
	left      *huffmanNode
	right     *huffmanNode
}

// --- Min-Heap Implementation for building the Huffman Tree ---

type minHeap []*huffmanNode

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].frequency < h[j].frequency }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(*huffmanNode)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	node := old[n-1]
	*h = old[:n-1]
	return node
}

// buildFrequencyTable counts how many times each byte appears in the data.
func buildFrequencyTable(data []byte) map[byte]int {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	return freq
}

// buildHuffmanTree constructs the Huffman tree from the frequency table.
func buildHuffmanTree(freq map[byte]int) *huffmanNode {
	h := &minHeap{}
	heap.Init(h)

	for symbol, count := range freq {
		heap.Push(h, &huffmanNode{symbol: symbol, frequency: count})
	}

	// Edge case: file has only one unique byte (e.g., "aaaa")
	if h.Len() == 1 {
		only := heap.Pop(h).(*huffmanNode)
		heap.Push(h, &huffmanNode{
			frequency: only.frequency,
			left:      only,
			right:     &huffmanNode{},
		})
	}

	for h.Len() > 1 {
		left := heap.Pop(h).(*huffmanNode)
		right := heap.Pop(h).(*huffmanNode)
		heap.Push(h, &huffmanNode{
			frequency: left.frequency + right.frequency,
			left:      left,
			right:     right,
		})
	}

	if h.Len() == 0 {
		return nil
	}
	return heap.Pop(h).(*huffmanNode)
}

// generateCodes recursively walks the Huffman tree and assigns binary codes.
func generateCodes(node *huffmanNode, prefix string, codeMap map[byte]string) {
	if node == nil {
		return
	}
	// It is a leaf node
	if node.left == nil && node.right == nil {
		codeMap[node.symbol] = prefix
		return
	}
	generateCodes(node.left, prefix+"0", codeMap)
	generateCodes(node.right, prefix+"1", codeMap)
}

// --- Serialization: Writing and Reading the Huffman Tree ---
// We serialize the tree into the compressed output so we can decompress later.
// Format: We use a pre-order traversal.
// - Leaf node: write 1 byte (0x01) followed by the symbol byte.
// - Internal node: write 1 byte (0x00).

func serializeTree(node *huffmanNode, buf *bytes.Buffer) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil {
		buf.WriteByte(0x01) // Marker: leaf
		buf.WriteByte(node.symbol)
		return
	}
	buf.WriteByte(0x00) // Marker: internal node
	serializeTree(node.left, buf)
	serializeTree(node.right, buf)
}

func deserializeTree(buf *bytes.Reader) (*huffmanNode, error) {
	marker, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if marker == 0x01 {
		symbol, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return &huffmanNode{symbol: symbol}, nil
	}
	left, err := deserializeTree(buf)
	if err != nil {
		return nil, err
	}
	right, err := deserializeTree(buf)
	if err != nil {
		return nil, err
	}
	return &huffmanNode{left: left, right: right}, nil
}

// --- Public API ---

// Compress takes any raw byte slice and returns a compressed byte slice.
// The output format is:
// [4 bytes: tree size] [N bytes: serialized tree] [1 byte: padding bits count] [compressed bits]
func Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("huffman: cannot compress empty data")
	}

	freq := buildFrequencyTable(data)
	root := buildHuffmanTree(freq)

	codeMap := make(map[byte]string)
	generateCodes(root, "", codeMap)

	// Serialize the tree
	var treeBuf bytes.Buffer
	serializeTree(root, &treeBuf)
	treeBytes := treeBuf.Bytes()

	// Build the bitstream from the data
	var bitStream bytes.Buffer
	var currentByte byte
	bitCount := 0

	for _, b := range data {
		code := codeMap[b]
		for _, bit := range code {
			currentByte <<= 1
			if bit == '1' {
				currentByte |= 1
			}
			bitCount++
			if bitCount == 8 {
				bitStream.WriteByte(currentByte)
				currentByte = 0
				bitCount = 0
			}
		}
	}

	// Handle the last partial byte and record how many padding bits were added
	paddingBits := byte(0)
	if bitCount > 0 {
		paddingBits = byte(8 - bitCount)
		currentByte <<= paddingBits
		bitStream.WriteByte(currentByte)
	}

	// Assemble the final output
	var output bytes.Buffer

	// Write tree size (4 bytes, big-endian)
	treeSize := uint32(len(treeBytes))
	if err := binary.Write(&output, binary.BigEndian, treeSize); err != nil {
		return nil, err
	}

	// Write the serialized tree
	output.Write(treeBytes)

	// Write padding bit count (1 byte)
	output.WriteByte(paddingBits)

	// Write the compressed bitstream
	output.Write(bitStream.Bytes())

	return output.Bytes(), nil
}

// Decompress takes a compressed byte slice and returns the original raw bytes.
func Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("huffman: cannot decompress empty data")
	}

	buf := bytes.NewReader(data)

	// Read tree size
	var treeSize uint32
	if err := binary.Read(buf, binary.BigEndian, &treeSize); err != nil {
		return nil, errors.New("huffman: failed to read tree size")
	}

	// Read the serialized tree bytes and reconstruct the tree
	treeBytes := make([]byte, treeSize)
	if _, err := buf.Read(treeBytes); err != nil {
		return nil, errors.New("huffman: failed to read tree data")
	}
	treeBuf := bytes.NewReader(treeBytes)
	root, err := deserializeTree(treeBuf)
	if err != nil {
		return nil, errors.New("huffman: failed to deserialize tree")
	}

	// Read padding bits count
	paddingBits, err := buf.ReadByte()
	if err != nil {
		return nil, errors.New("huffman: failed to read padding info")
	}

	// Read the remaining compressed bytes
	compressedBytes := make([]byte, buf.Len())
	if _, err := buf.Read(compressedBytes); err != nil {
		return nil, errors.New("huffman: failed to read compressed data")
	}

	// Decode the bitstream by walking the Huffman tree
	var output bytes.Buffer
	current := root

	totalBits := len(compressedBytes)*8 - int(paddingBits)
	bitsProcessed := 0

	for i, byteVal := range compressedBytes {
		for bit := 7; bit >= 0; bit-- {
			if bitsProcessed >= totalBits {
				break
			}
			// Stop before the padding bits of the last byte
			if i == len(compressedBytes)-1 && bit < int(paddingBits) {
				break
			}

			if (byteVal>>uint(bit))&1 == 0 {
				current = current.left
			} else {
				current = current.right
			}

			if current.left == nil && current.right == nil {
				output.WriteByte(current.symbol)
				current = root
			}
			bitsProcessed++
		}
	}

	return output.Bytes(), nil
}
