package uuidutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

/*
	UUID v7 byte representation.
	UUID v7 stores timestamp in the first 48 bits (6 bytes).
	The timestamp is in milliseconds since Unix epoch.

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                           unix_ts_ms                          |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          unix_ts_ms           |  ver  |  rand_a (12 bit seq)  |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|var|                        rand_b                             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                            rand_b                             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

var (
	errUUIDInvalidVersion = errors.New("invalid uuid version")
)

// GetUUID7Timestamp extracts the timestamp from a UUID v7
func GetUUID7Timestamp(u uuid.UUID) (time.Time, error) {
	if u.Version() != 7 {
		return time.Time{}, fmt.Errorf("checking uuid version: %w", errUUIDInvalidVersion)
	}

	// Get UUID timestamp part
	timestampBytes := u[0:6]

	// Convert bytes to uint64 timestamp (milliseconds)
	timestamp := uint64(timestampBytes[0])<<40 |
		uint64(timestampBytes[1])<<32 |
		uint64(timestampBytes[2])<<24 |
		uint64(timestampBytes[3])<<16 |
		uint64(timestampBytes[4])<<8 |
		uint64(timestampBytes[5])

	// Convert milliseconds to time.Time
	return time.UnixMilli(int64(timestamp)), nil
}

// NewUUID7FromTimestamp creates a UUID v7 from a provided timestamp
func NewUUID7FromTimestamp(timestamp time.Time) (uuid.UUID, error) {
	// Convert timestamp to milliseconds since Unix epoch
	millis := timestamp.UnixMilli()

	// Create a new UUID with random data
	uuidOut, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	// Extract nanoseconds for sequence number (12 bits)
	nano := timestamp.UnixNano()
	seq := (nano - millis*1000000) >> 8 // 12 bits of fractional milliseconds

	// Set the timestamp in the first 6 bytes (48 bits)
	uuidOut[0] = byte(millis >> 40)
	uuidOut[1] = byte(millis >> 32)
	uuidOut[2] = byte(millis >> 24)
	uuidOut[3] = byte(millis >> 16)
	uuidOut[4] = byte(millis >> 8)
	uuidOut[5] = byte(millis)

	// Set version 7 (0x70) and sequence number in bytes 6-7
	uuidOut[6] = 0x70 | (0x0F & byte(seq>>8))
	uuidOut[7] = byte(seq)

	// Bytes 8-15 remain random (already set by NewRandom)

	return uuidOut, nil
}

// NewUUID7FromString converts a string UUID to UUID v7 format
// This function extracts the timestamp from the input UUID and creates a new v7 UUID
func NewUUID7FromString(u string) (uuid.UUID, error) {
	uuidIn, err := uuid.Parse(u)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing input string uuid: %w", err)
	}

	uuidTimestamp, err := GetUUID7Timestamp(uuidIn)
	if err != nil {
		return uuid.Nil, fmt.Errorf("getting uuid v7 timestamp: %w", err)
	}

	uuidV7FromTimestamp, err := NewUUID7FromTimestamp(uuidTimestamp)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating uuid v7 from timestamp: %w", err)
	}

	// Create a new UUID v7 from the extracted timestamp
	return uuidV7FromTimestamp, nil
}
