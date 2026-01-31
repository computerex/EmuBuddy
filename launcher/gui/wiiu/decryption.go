package wiiu

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var wiiUCommonKey = []byte{0xD7, 0xB0, 0x04, 0x02, 0x65, 0x9B, 0xA2, 0xAB, 0xD2, 0xCB, 0x0D, 0xB2, 0x7F, 0xA2, 0xB6, 0x56}
var wiiCommonKeys = map[byte][]byte{
	0: {0xEB, 0xE4, 0x2A, 0x22, 0x5E, 0x85, 0x93, 0xE4, 0x48, 0xD9, 0xC5, 0x45, 0x73, 0x81, 0xAA, 0xF7},
	1: {0x63, 0xB8, 0x2B, 0xB4, 0xF4, 0x61, 0x4E, 0x2E, 0x13, 0xF2, 0xFE, 0xFB, 0xBA, 0x4C, 0x9B, 0x7E},
	2: {0x30, 0xBF, 0xC7, 0x6E, 0x7C, 0x19, 0xAF, 0xBB, 0x23, 0x16, 0x33, 0x30, 0xCE, 0xD7, 0xC2, 0x8D},
}

const (
	BLOCK_SIZE        = 0x8000
	BLOCK_SIZE_HASHED = 0x10000
	HASH_BLOCK_SIZE   = 0xFC00
	HASHES_SIZE       = 0x0400
	MAX_LEVELS        = 0x10
	READ_SIZE         = 8 * 1024 * 1024
	MAX_FST_SIZE      = 200 * 1024 * 1024
)

// FEntry represents a file system table entry
type FEntry struct {
	Type       byte
	NameOffset uint32
	Offset     uint32
	Length     uint32
	Flags      uint16
	ContentID  uint16
}

// FSTData represents file system table data
type FSTData struct {
	FSTReader   *bytes.Reader
	EntryCount  uint32
	Entries     uint32
	NamesOffset uint32
	FSTEntries  []FEntry
}

func extractFileHash(src *os.File, partDataOffset uint64, fileOffset uint64, size uint64, path string, cipherHashTree cipher.Block) error {
	writeSize := HASH_BLOCK_SIZE
	blockNumber := (fileOffset / HASH_BLOCK_SIZE) & 0x0F

	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create '%s': %w", path, err)
	}
	defer dst.Close()
	bw := bufio.NewWriterSize(dst, BLOCK_SIZE_HASHED)
	defer bw.Flush()

	roffset := fileOffset / HASH_BLOCK_SIZE * BLOCK_SIZE_HASHED
	soffset := fileOffset - (fileOffset / HASH_BLOCK_SIZE * HASH_BLOCK_SIZE)

	if soffset+size > uint64(writeSize) {
		writeSize = writeSize - int(soffset)
	}

	_, err = src.Seek(int64(partDataOffset+roffset), io.SeekStart)
	if err != nil {
		return err
	}

	encryptedHashedContentBuffer := make([]byte, BLOCK_SIZE_HASHED)
	decryptedHashedContentBuffer := make([]byte, BLOCK_SIZE_HASHED)
	hashes := make([]byte, HASHES_SIZE)

	for size > 0 {
		if uint64(writeSize) > size {
			writeSize = int(size)
		}

		stat, err := src.Stat()
		if err != nil {
			return err
		}
		fileSize := stat.Size()
		currentPos := int64(partDataOffset + roffset)
		remainingFile := fileSize - currentPos

		readLen := BLOCK_SIZE_HASHED
		if int64(readLen) > remainingFile {
			readLen = int(remainingFile)
		}

		if readLen%aes.BlockSize != 0 {
			return fmt.Errorf("read length %d is not a multiple of AES block size", readLen)
		}

		if _, err := io.ReadFull(src, encryptedHashedContentBuffer[:readLen]); err != nil {
			return fmt.Errorf("failed to read encrypted content block at offset %d (read %d of %d bytes): %w",
				partDataOffset+roffset, 0, BLOCK_SIZE_HASHED, err)
		}

		var zeroIV [aes.BlockSize]byte
		cipher.NewCBCDecrypter(cipherHashTree, zeroIV[:]).CryptBlocks(hashes, encryptedHashedContentBuffer[:HASHES_SIZE])

		h0Hash := hashes[0x14*blockNumber : 0x14*blockNumber+sha1.Size]

		var ivBlock [aes.BlockSize]byte
		copy(ivBlock[:], hashes[0x14*blockNumber:0x14*blockNumber+aes.BlockSize])

		cipher.NewCBCDecrypter(cipherHashTree, ivBlock[:]).CryptBlocks(decryptedHashedContentBuffer, encryptedHashedContentBuffer[HASHES_SIZE:])

		hash := sha1.Sum(decryptedHashedContentBuffer[:HASH_BLOCK_SIZE])
		if !bytes.Equal(hash[:], h0Hash) {
			return errors.New("h0 hash mismatch")
		}

		dataToWrite := decryptedHashedContentBuffer

		n, err := bw.Write(dataToWrite[soffset : soffset+uint64(writeSize)])
		if err != nil {
			return err
		}

		size -= uint64(n)

		blockNumber++
		if blockNumber >= 16 {
			blockNumber = 0
		}

		if soffset != 0 {
			writeSize = HASH_BLOCK_SIZE
			soffset = 0
		}

		roffset += uint64(BLOCK_SIZE_HASHED)
	}

	return nil
}

func extractFile(src *os.File, partDataOffset uint64, fileOffset uint64, size uint64, path string, contentId uint16, cipherHashTree cipher.Block) error {
	writeSize := BLOCK_SIZE

	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create '%s': %w", path, err)
	}
	defer dst.Close()
	bw := bufio.NewWriterSize(dst, BLOCK_SIZE)
	defer bw.Flush()

	roffset := fileOffset / BLOCK_SIZE * BLOCK_SIZE
	soffset := fileOffset - (fileOffset / BLOCK_SIZE * BLOCK_SIZE)

	if soffset+size > uint64(writeSize) {
		writeSize = writeSize - int(soffset)
	}

	_, err = src.Seek(int64(partDataOffset+roffset), io.SeekStart)
	if err != nil {
		return err
	}

	var ivLocal [aes.BlockSize]byte
	ivLocal[1] = byte(contentId)
	aesCipher := cipher.NewCBCDecrypter(cipherHashTree, ivLocal[:])

	encryptedContentBuffer := make([]byte, BLOCK_SIZE)
	decryptedContentBuffer := make([]byte, BLOCK_SIZE)

	for size > 0 {
		if uint64(writeSize) > size {
			writeSize = int(size)
		}

		stat, err := src.Stat()
		if err != nil {
			return err
		}
		fileSize := stat.Size()
		currentPos := int64(partDataOffset + roffset)
		remainingFile := fileSize - currentPos

		readLen := BLOCK_SIZE
		if int64(readLen) > remainingFile {
			readLen = int(remainingFile)
		}

		if readLen%aes.BlockSize != 0 {
			return fmt.Errorf("read length %d is not a multiple of AES block size", readLen)
		}

		if _, err := io.ReadFull(src, encryptedContentBuffer[:readLen]); err != nil {
			return fmt.Errorf("failed to read encrypted content block at offset %d (expected %d bytes): %w",
				partDataOffset+roffset, readLen, err)
		}

		aesCipher.CryptBlocks(decryptedContentBuffer[:readLen], encryptedContentBuffer[:readLen])

		n, err := bw.Write(decryptedContentBuffer[soffset : soffset+uint64(writeSize)])
		if err != nil {
			return err
		}

		size -= uint64(n)

		if soffset != 0 {
			writeSize = BLOCK_SIZE
			soffset = 0
		}

		roffset += uint64(BLOCK_SIZE)
	}

	return nil
}

func (fst *FSTData) Parse() error {
	var err error

	if _, err := fst.FSTReader.Seek(0x8, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to entry count: %w", err)
	}
	if fst.EntryCount, err = readInt(fst.FSTReader, 4); err != nil {
		return fmt.Errorf("failed to read entry count: %w", err)
	}

	if _, err := fst.FSTReader.Seek(int64(0x20+fst.EntryCount*0x20+8), io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to entries: %w", err)
	}
	if fst.Entries, err = readInt(fst.FSTReader, 4); err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}
	fst.NamesOffset = 0x20 + fst.EntryCount*0x20 + fst.Entries*0x10

	if _, err := fst.FSTReader.Seek(4, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek to names offset: %w", err)
	}

	for i := uint32(0); i < fst.Entries; i++ {
		entry := FEntry{}
		if entry.Type, err = readByte(fst.FSTReader); err != nil {
			return fmt.Errorf("failed to read entry type for entry %d: %w", i, err)
		}

		nameOffset, err := read3BytesBE(fst.FSTReader)
		if err != nil {
			return fmt.Errorf("failed to read name offset for entry %d: %w", i, err)
		}
		entry.NameOffset = uint32(nameOffset)

		if entry.Offset, err = readInt(fst.FSTReader, 4); err != nil {
			return fmt.Errorf("failed to read entry offset for entry %d: %w", i, err)
		}
		if entry.Length, err = readInt(fst.FSTReader, 4); err != nil {
			return fmt.Errorf("failed to read entry length for entry %d: %w", i, err)
		}

		flags, err := readInt16(fst.FSTReader, 2)
		if err != nil {
			return fmt.Errorf("failed to read flags for entry %d: %w", i, err)
		}
		entry.Flags = flags

		contentID, err := readInt16(fst.FSTReader, 2)
		if err != nil {
			return fmt.Errorf("failed to read content ID for entry %d: %w", i, err)
		}
		entry.ContentID = contentID

		fst.FSTEntries = append(fst.FSTEntries, entry)
	}

	return nil
}

func readByte(f io.ReadSeeker) (byte, error) {
	buf := make([]byte, 1)
	n, err := f.Read(buf)
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, io.ErrUnexpectedEOF
	}
	return buf[0], nil
}

func readInt(f io.ReadSeeker, s int) (uint32, error) {
	buf := make([]byte, 4)

	n, err := f.Read(buf[:s])
	if err != nil {
		return 0, err
	}

	if n < s {
		if _, err := f.Seek(int64(-n), io.SeekCurrent); err != nil {
			return 0, err
		}
		return 0, io.ErrUnexpectedEOF
	}

	return binary.BigEndian.Uint32(buf), nil
}

func readInt16(f io.ReadSeeker, s int) (uint16, error) {
	buf := make([]byte, 2)

	n, err := f.Read(buf[:s])
	if err != nil {
		return 0, err
	}

	if n < s {
		if _, err := f.Seek(int64(-n), io.SeekCurrent); err != nil {
			return 0, err
		}
		return 0, io.ErrUnexpectedEOF
	}

	return binary.BigEndian.Uint16(buf), nil
}

func readString(f io.ReadSeeker) (string, error) {
	buffer := bytes.NewBuffer(nil)
	chunk := make([]byte, 64)

	for {
		n, err := f.Read(chunk)
		if err != nil && err != io.EOF {
			return "", err
		}

		if n == 0 {
			break
		}

		nullIndex := bytes.IndexByte(chunk[:n], 0)
		if nullIndex != -1 {
			buffer.Write(chunk[:nullIndex])
			_, err = f.Seek(int64(-(n - nullIndex - 1)), io.SeekCurrent)
			if err != nil {
				return "", err
			}
			return buffer.String(), nil
		}

		buffer.Write(chunk[:n])

		if err == io.EOF {
			break
		}
	}

	if buffer.Len() == 0 {
		return "", io.EOF
	}

	return buffer.String(), nil
}

func read3BytesBE(f io.ReadSeeker) (uint32, error) {
	b := make([]byte, 3)
	if _, err := f.Read(b); err != nil {
		return 0, err
	}
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16, nil
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func decryptContentToBuffer(encryptedFile *os.File, decryptedBuffer *bytes.Buffer, cipherHashTree cipher.Block, content Content) error {
	hasHashTree := content.Type&2 != 0
	encryptedStat, err := encryptedFile.Stat()
	if err != nil {
		return err
	}
	encryptedSize := encryptedStat.Size()
	path := filepath.Dir(encryptedFile.Name())

	var growSize int64
	if hasHashTree {
		growSize = encryptedSize
	} else {
		growSize = int64(content.Size)
	}

	if growSize > MAX_FST_SIZE {
		return fmt.Errorf("FST size %d exceeds maximum limit of %d", growSize, MAX_FST_SIZE)
	}

	decryptedBuffer.Grow(int(growSize))

	if hasHashTree {
		chunkCount := encryptedSize / 0x10000
		h3Data, err := os.ReadFile(filepath.Join(path, fmt.Sprintf("%s.h3", content.CIDStr)))
		if err != nil {
			return err
		}
		h3BytesSHASum := sha1.Sum(h3Data)
		if !bytes.Equal(h3BytesSHASum[:], content.Hash[:sha1.Size]) {
			return errors.New("H3 Hash mismatch")
		}

		h0HashNum := int64(0)
		h1HashNum := int64(0)
		h2HashNum := int64(0)
		h3HashNum := int64(0)

		hashes := make([]byte, HASHES_SIZE)
		hashesBuffer := make([]byte, HASHES_SIZE)
		decryptedDataBuffer := make([]byte, 0xFC00)

		for chunkNum := int64(0); chunkNum < chunkCount; chunkNum++ {
			if _, err := io.ReadFull(encryptedFile, hashesBuffer); err != nil {
				return err
			}
			var zeroIV [aes.BlockSize]byte
			cipher.NewCBCDecrypter(cipherHashTree, zeroIV[:]).CryptBlocks(hashes, hashesBuffer)

			h0Hashes := hashes[0:0x140]
			h1Hashes := hashes[0x140:0x280]
			h2Hashes := hashes[0x280:0x3c0]

			h0Hash := h0Hashes[(h0HashNum * 0x14):((h0HashNum + 1) * 0x14)]
			h1Hash := h1Hashes[(h1HashNum * 0x14):((h1HashNum + 1) * 0x14)]
			h2Hash := h2Hashes[(h2HashNum * 0x14):((h2HashNum + 1) * 0x14)]
			h3Hash := h3Data[(h3HashNum * 0x14):((h3HashNum + 1) * 0x14)]

			h0HashesHash := sha1.Sum(h0Hashes)
			h1HashesHash := sha1.Sum(h1Hashes)
			h2HashesHash := sha1.Sum(h2Hashes)

			if !bytes.Equal(h0HashesHash[:], h1Hash) {
				return errors.New("h0 Hashes Hash mismatch")
			}
			if !bytes.Equal(h1HashesHash[:], h2Hash) {
				return errors.New("h1 Hashes Hash mismatch")
			}
			if !bytes.Equal(h2HashesHash[:], h3Hash) {
				return errors.New("h2 Hashes Hash mismatch")
			}

			if _, err := io.ReadFull(encryptedFile, decryptedDataBuffer); err != nil {
				return err
			}

			cipher.NewCBCDecrypter(cipherHashTree, h0Hash[:16]).CryptBlocks(decryptedDataBuffer, decryptedDataBuffer)
			decryptedDataHash := sha1.Sum(decryptedDataBuffer)

			if !bytes.Equal(decryptedDataHash[:], h0Hash) {
				return errors.New("data block hash invalid")
			}

			if _, err = decryptedBuffer.Write(hashes); err != nil {
				return err
			}
			if _, err = decryptedBuffer.Write(decryptedDataBuffer); err != nil {
				return err
			}

			h0HashNum++
			if h0HashNum >= 16 {
				h0HashNum = 0
				h1HashNum++
			}
			if h1HashNum >= 16 {
				h1HashNum = 0
				h2HashNum++
			}
			if h2HashNum >= 16 {
				h2HashNum = 0
				h3HashNum++
			}
		}
	} else {
		f, err := os.Open(encryptedFile.Name())
		if err == nil {
			h := sha1.New()
			if _, err := io.Copy(h, f); err == nil {
				if bytes.Equal(content.Hash[:sha1.Size], h.Sum(nil)) {
					_, _ = f.Seek(0, 0)
					decryptedBuffer.Reset()
					_, _ = io.Copy(decryptedBuffer, f)
					_ = f.Close()
					return nil
				}
			}
			_ = f.Close()
		}

		_, _ = encryptedFile.Seek(0, io.SeekStart)

		var ivContent [aes.BlockSize]byte
		copy(ivContent[:], content.Index)
		cipherContent := cipher.NewCBCDecrypter(cipherHashTree, ivContent[:])
		contentHash := sha1.New()
		left := content.Size
		leftHash := content.Size

		decBuf := make([]byte, READ_SIZE)
		readSizedBuffer := make([]byte, READ_SIZE)

		for i := 0; i <= int(content.Size/READ_SIZE)+1; i++ {
			toRead := minUint64(READ_SIZE, left)
			toReadHash := minUint64(READ_SIZE, leftHash)

			toReadAligned := (toRead + 15) &^ 15

			if _, err = io.ReadFull(encryptedFile, readSizedBuffer[:toReadAligned]); err != nil {
				return err
			}

			cipherContent.CryptBlocks(decBuf[:toReadAligned], readSizedBuffer[:toReadAligned])
			contentHash.Write(decBuf[:toReadHash])
			if _, err = decryptedBuffer.Write(decBuf[:toRead]); err != nil {
				return err
			}

			left -= toRead
			leftHash -= toRead

			if left == 0 {
				break
			}
		}
		if !bytes.Equal(content.Hash[:sha1.Size], contentHash.Sum(nil)) {
			return errors.New("content hash mismatch")
		}
	}
	return nil
}

func extractU8(data []byte, outputPath string) error {
	reader := bytes.NewReader(data)
	var magic uint32
	if err := binary.Read(reader, binary.BigEndian, &magic); err != nil {
		return err
	}
	if magic != 0x55AA382D {
		return errors.New("invalid U8 magic")
	}

	var rootNodeOffset, headerSize, dataOffset uint32
	_ = binary.Read(reader, binary.BigEndian, &rootNodeOffset)
	_ = binary.Read(reader, binary.BigEndian, &headerSize)
	_ = binary.Read(reader, binary.BigEndian, &dataOffset)

	if _, err := reader.Seek(int64(rootNodeOffset), io.SeekStart); err != nil {
		return err
	}

	var rootType uint16
	var rootNameOffset uint16
	var rootDataOffset uint32
	var totalNodes uint32
	_ = binary.Read(reader, binary.BigEndian, &rootType)
	_ = binary.Read(reader, binary.BigEndian, &rootNameOffset)
	_ = binary.Read(reader, binary.BigEndian, &rootDataOffset)
	_ = binary.Read(reader, binary.BigEndian, &totalNodes)

	if totalNodes == 0 || totalNodes > 100000 {
		return fmt.Errorf("invalid U8: totalNodes=%d", totalNodes)
	}

	_ = os.MkdirAll(outputPath, 0755)

	nodes := make([]struct {
		Type       uint16
		NameOffset uint16
		DataOffset uint32
		Size       uint32
	}, totalNodes)

	nodes[0].Type = rootType
	nodes[0].NameOffset = rootNameOffset
	nodes[0].DataOffset = rootDataOffset
	nodes[0].Size = totalNodes

	for i := uint32(1); i < totalNodes; i++ {
		_ = binary.Read(reader, binary.BigEndian, &nodes[i].Type)
		_ = binary.Read(reader, binary.BigEndian, &nodes[i].NameOffset)
		_ = binary.Read(reader, binary.BigEndian, &nodes[i].DataOffset)
		_ = binary.Read(reader, binary.BigEndian, &nodes[i].Size)
	}

	stringTableOffset := rootNodeOffset + (totalNodes * 12)
	stringTableSize := dataOffset - stringTableOffset
	stringTable := make([]byte, stringTableSize)
	_, _ = reader.Seek(int64(stringTableOffset), io.SeekStart)
	_, _ = io.ReadFull(reader, stringTable)

	currentDir := outputPath
	dirStack := []string{outputPath}
	breakNodes := make([]uint32, 128)
	stackIdx := 0
	breakNodes[0] = totalNodes

	for i := uint32(1); i < totalNodes; i++ {
		name := ""
		for j := nodes[i].NameOffset; j < uint16(len(stringTable)) && stringTable[j] != 0; j++ {
			name += string(stringTable[j])
		}

		if nodes[i].Type == 0x0100 {
			currentDir = filepath.Join(currentDir, name)
			_ = os.MkdirAll(currentDir, 0755)
			stackIdx++
			dirStack = append(dirStack, currentDir)
			breakNodes[stackIdx] = nodes[i].Size
		} else {
			if nodes[i].Size > 100*1024*1024 {
				continue
			}
			filePath := filepath.Join(currentDir, name)
			fileData := make([]byte, nodes[i].Size)
			_, _ = reader.Seek(int64(nodes[i].DataOffset), io.SeekStart)
			_, _ = io.ReadFull(reader, fileData)
			_ = os.WriteFile(filePath, fileData, 0644)
		}

		for stackIdx > 0 && breakNodes[stackIdx] == i+1 {
			dirStack = dirStack[:len(dirStack)-1]
			currentDir = dirStack[len(dirStack)-1]
			stackIdx--
		}
	}

	return nil
}

// DecryptContents decrypts the contents of a downloaded Wii U title
func DecryptContents(path string, progressReporter ProgressReporter, deleteEncryptedContents bool) error {
	tmdPath := filepath.Join(path, "title.tmd")
	if _, err := os.Stat(tmdPath); os.IsNotExist(err) {
		return err
	}

	tmdData, err := os.ReadFile(tmdPath)
	if err != nil {
		return err
	}
	tmd, err := ParseTMD(tmdData)
	if err != nil {
		return err
	}

	for i := range tmd.Contents {
		tmd.Contents[i].CIDStr = fmt.Sprintf("%08X", tmd.Contents[i].ID)
		_, err := os.Stat(filepath.Join(path, tmd.Contents[i].CIDStr+".app"))
		if err != nil {
			tmd.Contents[i].CIDStr = fmt.Sprintf("%08x", tmd.Contents[i].ID)
			_, err = os.Stat(filepath.Join(path, tmd.Contents[i].CIDStr+".app"))
			if err != nil {
				return errors.New("content not found")
			}
		}
	}

	var encryptedTitleKey []byte

	ticketPath := filepath.Join(path, "title.tik")
	var ticketKeyIndex byte = 0xFF

	if _, err := os.Stat(ticketPath); err == nil {
		cetk, err := os.Open(ticketPath)
		if err == nil {
			_, _ = cetk.Seek(0x1BF, 0)
			encryptedTitleKey = make([]byte, 0x10)
			if _, err := io.ReadFull(cetk, encryptedTitleKey); err != nil {
				_ = cetk.Close()
				return err
			}
			_, _ = cetk.Seek(0x1F1, 0)
			_ = binary.Read(cetk, binary.BigEndian, &ticketKeyIndex)
			if err := cetk.Close(); err != nil {
				return err
			}
		}
	}

	var selectedCommonKey []byte
	if tmd.Version == 0 {
		if key, ok := wiiCommonKeys[ticketKeyIndex]; ok {
			selectedCommonKey = key
		} else {
			selectedCommonKey = wiiCommonKeys[0]
		}
	} else {
		selectedCommonKey = wiiUCommonKey
	}

	c, err := aes.NewCipher(selectedCommonKey)
	if err != nil {
		return err
	}

	var titleIDBytes [8]byte
	binary.BigEndian.PutUint64(titleIDBytes[:], tmd.TitleID)
	var ivTitle [16]byte
	copy(ivTitle[:], titleIDBytes[:])
	cbc := cipher.NewCBCDecrypter(c, ivTitle[:])

	decryptedTitleKey := make([]byte, len(encryptedTitleKey))
	cbc.CryptBlocks(decryptedTitleKey, encryptedTitleKey)

	cipherHashTree, err := aes.NewCipher(decryptedTitleKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if tmd.Version == TMD_VERSION_WIIU {
		fstEncFile, err := os.Open(filepath.Join(path, tmd.Contents[0].CIDStr+".app"))
		if err != nil {
			return err
		}

		decryptedBuffer := bytes.Buffer{}
		if err := decryptContentToBuffer(fstEncFile, &decryptedBuffer, cipherHashTree, tmd.Contents[0]); err != nil {
			if err := fstEncFile.Close(); err != nil {
				return err
			}
			return err
		}
		if err := fstEncFile.Close(); err != nil {
			return err
		}
		fst := FSTData{FSTReader: bytes.NewReader(decryptedBuffer.Bytes()), FSTEntries: make([]FEntry, 0), EntryCount: 0, Entries: 0, NamesOffset: 0}
		if err := fst.Parse(); err != nil {
			return err
		}

		outputPath := path
		entry := make([]uint32, 0x10)
		lEntry := make([]uint32, 0x10)
		level := uint32(0)

		for i := uint32(0); i < fst.Entries-1; i++ {
			progressReporter.UpdateDecryptionProgress(float64(i) / float64(fst.Entries-1))
			if level > 0 {
				for (level >= 1) && (lEntry[level-1] == i+1) {
					level--
				}
			}

			if fst.FSTEntries[i].Type&1 != 0 {
				entry[level] = i
				lEntry[level] = fst.FSTEntries[i].Length
				level++
				if level >= MAX_LEVELS {
					return errors.New("level >= MAX_LEVELS")
				}
			} else {
				pathOffset := uint32(0)
				outputPath = path
				for j := uint32(0); j < level; j++ {
					pathOffset = fst.FSTEntries[entry[j]].NameOffset & 0x00FFFFFF
					fst.FSTReader.Seek(int64(fst.NamesOffset+pathOffset), io.SeekStart)
					directory, err := readString(fst.FSTReader)
					if err != nil {
						return fmt.Errorf("failed to read directory name: %w", err)
					}
					outputPath = filepath.Join(outputPath, directory)
					if err := os.MkdirAll(outputPath, 0755); err != nil {
						return fmt.Errorf("failed to create directory: %w", err)
					}
				}
				pathOffset = fst.FSTEntries[i].NameOffset & 0x00FFFFFF
				fst.FSTReader.Seek(int64(fst.NamesOffset+pathOffset), io.SeekStart)
				fileName, err := readString(fst.FSTReader)
				if err != nil {
					return fmt.Errorf("failed to read file name: %w", err)
				}
				outputPath = filepath.Clean(filepath.Join(outputPath, fileName))
				contentOffset := uint64(fst.FSTEntries[i].Offset)
				if fst.FSTEntries[i].Flags&4 == 0 {
					contentOffset <<= 5
				}
				if fst.FSTEntries[i].Type&0x80 == 0 {
					matchingContent := tmd.Contents[fst.FSTEntries[i].ContentID]
					tmdFlags := matchingContent.Type
					srcFile, err := os.Open(filepath.Join(path, matchingContent.CIDStr+".app"))
					if err != nil {
						return err
					}
					if tmdFlags&0x02 != 0 {
						err = extractFileHash(srcFile, 0, contentOffset, uint64(fst.FSTEntries[i].Length), outputPath, cipherHashTree)
					} else {
						err = extractFile(srcFile, 0, contentOffset, uint64(fst.FSTEntries[i].Length), outputPath, fst.FSTEntries[i].ContentID, cipherHashTree)
					}
					srcFile.Close()
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		for i, content := range tmd.Contents {
			progressReporter.UpdateDecryptionProgress(float64(i) / float64(len(tmd.Contents)))
			srcFile, err := os.Open(filepath.Join(path, content.CIDStr+".app"))
			if err != nil {
				return err
			}

			decryptedBuffer := bytes.Buffer{}
			if err := decryptContentToBuffer(srcFile, &decryptedBuffer, cipherHashTree, content); err != nil {
				srcFile.Close()
				return err
			}
			srcFile.Close()

			decData := decryptedBuffer.Bytes()

			foundU8 := false
			extractCount := 0
			for pos := 0; pos < len(decData)-32; pos += 16 {
				if binary.BigEndian.Uint32(decData[pos:pos+4]) == 0x55AA382D {
					if pos+32 > len(decData) {
						continue
					}
					rootNodeOffset := binary.BigEndian.Uint32(decData[pos+4 : pos+8])
					dataOffset := binary.BigEndian.Uint32(decData[pos+12 : pos+16])

					if rootNodeOffset < 32 || rootNodeOffset > uint32(len(decData)) {
						continue
					}
					if dataOffset < rootNodeOffset || dataOffset > uint32(len(decData)) {
						continue
					}
					if dataOffset-rootNodeOffset > 1000000 {
						continue
					}

					foundU8 = true
					var outPath string
					if extractCount == 0 && i == 0 {
						outPath = path
					} else if extractCount == 0 {
						outPath = filepath.Join(path, content.CIDStr)
					} else {
						outPath = filepath.Join(path, content.CIDStr, fmt.Sprintf("u8_%X", pos))
					}

					if err := extractU8(decData[pos:], outPath); err == nil {
						extractCount++
					}
				}
			}

			if !foundU8 {
				_ = os.WriteFile(filepath.Join(path, content.CIDStr+".app"), decData, 0644)
			}
		}
	}

	progressReporter.UpdateDecryptionProgress(1.0)

	if deleteEncryptedContents {
		doDeleteEncryptedContents(path)
	}
	return nil
}

func doDeleteEncryptedContents(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) > 4 && (name[len(name)-4:] == ".app" || name[len(name)-3:] == ".h3") ||
			name == "title.tmd" ||
			name == "title.tik" ||
			name == "title.cert" {
			if err := os.Remove(filepath.Join(path, name)); err != nil {
				return err
			}
		}
	}
	return nil
}
