package hermes

import (
	"encoding/binary"
	"github.com/jtejido/hermes/t1ha"
	"math/rand"
	"reflect"
	"unsafe"
)

func getIndexKey(key string) uint64 {
	k := hasher([]byte(key), 0)
	return k
}

func getAltIndex(fp byte, i uint, numBuckets uint) uint {
	hash := uint(hasher([]byte{fp}, 0))
	return (i ^ hash) % numBuckets
}

func getFingerprint(data []byte) byte {
	hash := hasher(data, 1335)
	return byte(hash%255 + 1)
}

func getFilterComponents(data []byte, numBuckets uint) (uint, uint, byte) {
	hash := hasher(data, 0)
	f := getFingerprint(data)
	i1 := uint(hash) % numBuckets
	i2 := getAltIndex(f, i1, numBuckets)
	return i1, i2, f
}

func hasher(value []byte, seed uint64) uint64 {
	hasher := t1ha.New64WithSeed(seed)
	hasher.Write(value)
	return hasher.Sum64()
}

func mBToBytes(value int) int {
	return value * 1024 * 1024
}

func bytesToMB(value int) int {
	return value / 1024 / 1024
}

func randi(i1, i2 uint) uint {
	if rand.Intn(2) == 0 {
		return i1
	}
	return i2
}

func nextPowerOfTwo(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}

func wrapEntry(timestamp uint64, hash uint64, key string, entry []byte) []byte {
	keyLength := len(key)
	blobLength := len(entry) + headersSizeInBytes + keyLength

	buffer := make([]byte, blobLength)

	binary.LittleEndian.PutUint64(buffer, timestamp)
	binary.LittleEndian.PutUint64(buffer[timestampSizeInBytes:], hash)
	binary.LittleEndian.PutUint16(buffer[timestampSizeInBytes+hashSizeInBytes:], uint16(keyLength))
	copy(buffer[headersSizeInBytes:], key)
	copy(buffer[headersSizeInBytes+keyLength:], entry)

	return buffer[:blobLength]
}

// will use it later on
func getTimestampFromEntry(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

// will use it later on
func getHashFromEntry(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data[timestampSizeInBytes:])
}

func getKeyFromEntry(data []byte) string {
	length := binary.LittleEndian.Uint16(data[timestampSizeInBytes+hashSizeInBytes:])
	return bytesToString(data[headersSizeInBytes : headersSizeInBytes+length])
}

func getValueFromEntry(data []byte) []byte {
	length := binary.LittleEndian.Uint16(data[timestampSizeInBytes+hashSizeInBytes:])

	return data[headersSizeInBytes+length:]
}

func bytesToString(b []byte) string {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strHeader := reflect.StringHeader{Data: bytesHeader.Data, Len: bytesHeader.Len}
	return *(*string)(unsafe.Pointer(&strHeader))
}
