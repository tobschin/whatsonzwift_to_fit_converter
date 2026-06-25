package fit

import (
	"bytes"
	"encoding/binary"
	"time"
)

// FIT protocol constants
const (
	fitHeaderSize   = 14
	fitProtocolVer  = 0x20 // 2.0
	fitProfileVer   = 0x0814
	fitDataType     = ".FIT"
)

// Message numbers
const (
	mesgNumFileID      = 0
	mesgNumWorkout     = 26
	mesgNumWorkoutStep = 27
)

// Field types
const (
	fitTypeEnum    = 0
	fitTypeSint8   = 1
	fitTypeUint8   = 2
	fitTypeSint16  = 3
	fitTypeUint16  = 4
	fitTypeSint32  = 5
	fitTypeUint32  = 6
	fitTypeString  = 7
	fitTypeUint32z = 10
	fitTypeUint16z = 11
)

// Intensity values
const (
	intensityActive  = 0
	intensityRest    = 1
	intensityWarmup  = 2
	intensityCooldown = 3
)

// Duration types
const (
	durationTime              = 0
	durationRepeatUntilSteps  = 6
)

// Target types
const (
	targetPower = 4
)

// File types
const (
	fileTypeWorkout = 5
)

// Sport types
const (
	sportCycling = 2
)

// Encoder writes FIT binary data
type Encoder struct {
	buf          bytes.Buffer
	crc          uint16
	localMesgMap map[uint16]byte
	nextLocalID  byte
}

func NewEncoder() *Encoder {
	return &Encoder{
		localMesgMap: make(map[uint16]byte),
	}
}

// Encode writes the complete FIT file
func (e *Encoder) Encode(name string, steps []Step) ([]byte, error) {
	// Reserve space for header (write at end)
	e.buf.Reset()

	// Write data messages
	e.writeFileID()
	e.writeWorkout(name, uint16(len(steps)))

	for i, step := range steps {
		e.writeWorkoutStep(uint16(i), step)
	}

	// Build final file
	dataBytes := e.buf.Bytes()
	dataSize := uint32(len(dataBytes))

	var result bytes.Buffer

	// Write FIT header
	header := make([]byte, fitHeaderSize)
	header[0] = fitHeaderSize
	header[1] = fitProtocolVer
	binary.LittleEndian.PutUint16(header[2:4], fitProfileVer)
	binary.LittleEndian.PutUint32(header[4:8], dataSize)
	copy(header[8:12], fitDataType)
	headerCRC := crc16(header[:12])
	binary.LittleEndian.PutUint16(header[12:14], headerCRC)
	result.Write(header)

	// Write data
	result.Write(dataBytes)

	// Write file CRC
	fileCRC := crc16(result.Bytes())
	var crcBytes [2]byte
	binary.LittleEndian.PutUint16(crcBytes[:], fileCRC)
	result.Write(crcBytes[:])

	return result.Bytes(), nil
}

func (e *Encoder) writeFileID() {
	// Definition message
	e.writeDefinition(mesgNumFileID, []fieldDef{
		{num: 0, size: 1, baseType: fitTypeEnum},    // type
		{num: 1, size: 2, baseType: fitTypeUint16},   // manufacturer
		{num: 2, size: 2, baseType: fitTypeUint16},   // product
		{num: 3, size: 4, baseType: fitTypeUint32z},  // serial_number
		{num: 4, size: 4, baseType: fitTypeUint32},   // time_created
	})

	// Data message
	localID := e.localMesgMap[mesgNumFileID]
	e.buf.WriteByte(localID) // record header

	e.buf.WriteByte(fileTypeWorkout)                        // type = workout
	binary.Write(&e.buf, binary.LittleEndian, uint16(1))   // manufacturer = garmin
	binary.Write(&e.buf, binary.LittleEndian, uint16(0))   // product
	binary.Write(&e.buf, binary.LittleEndian, uint32(12345)) // serial_number
	// time_created: seconds since 1989-12-31 00:00:00 UTC
	garminEpoch := time.Date(1989, 12, 31, 0, 0, 0, 0, time.UTC)
	ts := uint32(time.Now().UTC().Sub(garminEpoch).Seconds())
	binary.Write(&e.buf, binary.LittleEndian, ts)
}

func (e *Encoder) writeWorkout(name string, numSteps uint16) {
	nameBytes := fitString(name, 64)

	// Definition message
	e.writeDefinition(mesgNumWorkout, []fieldDef{
		{num: 4, size: 1, baseType: fitTypeEnum},             // sport
		{num: 6, size: 2, baseType: fitTypeUint16},           // num_valid_steps
		{num: 8, size: uint8(len(nameBytes)), baseType: fitTypeString}, // wkt_name
	})

	// Data message
	localID := e.localMesgMap[mesgNumWorkout]
	e.buf.WriteByte(localID)

	e.buf.WriteByte(sportCycling)                                 // sport
	binary.Write(&e.buf, binary.LittleEndian, numSteps)           // num_valid_steps
	e.buf.Write(nameBytes)                                        // wkt_name
}

func (e *Encoder) writeWorkoutStep(index uint16, step Step) {
	// Definition message (only write once, reuse local ID)
	if _, exists := e.localMesgMap[mesgNumWorkoutStep]; !exists {
		e.writeDefinition(mesgNumWorkoutStep, []fieldDef{
			{num: 254, size: 2, baseType: fitTypeUint16},  // message_index
			{num: 1, size: 1, baseType: fitTypeEnum},      // duration_type
			{num: 2, size: 4, baseType: fitTypeUint32},    // duration_value
			{num: 3, size: 1, baseType: fitTypeEnum},      // target_type
			{num: 4, size: 4, baseType: fitTypeUint32},    // target_value
			{num: 5, size: 4, baseType: fitTypeUint32},    // custom_target_value_low
			{num: 6, size: 4, baseType: fitTypeUint32},    // custom_target_value_high
			{num: 7, size: 1, baseType: fitTypeEnum},      // intensity
		})
	}

	localID := e.localMesgMap[mesgNumWorkoutStep]
	e.buf.WriteByte(localID)

	binary.Write(&e.buf, binary.LittleEndian, index) // message_index

	if step.IsRepeat {
		e.buf.WriteByte(durationRepeatUntilSteps)                          // duration_type
		binary.Write(&e.buf, binary.LittleEndian, step.DurationValue)     // duration_value (step to repeat to)
		e.buf.WriteByte(0)                                                 // target_type
		binary.Write(&e.buf, binary.LittleEndian, step.TargetValue)       // target_value (repeat count)
		binary.Write(&e.buf, binary.LittleEndian, uint32(0))              // custom low
		binary.Write(&e.buf, binary.LittleEndian, uint32(0))              // custom high
		e.buf.WriteByte(intensityActive)                                   // intensity
	} else {
		e.buf.WriteByte(durationTime)                                           // duration_type
		binary.Write(&e.buf, binary.LittleEndian, step.DurationValue)          // duration_value (ms)
		e.buf.WriteByte(targetPower)                                            // target_type
		binary.Write(&e.buf, binary.LittleEndian, uint32(0))                   // target_value (0 = custom)
		binary.Write(&e.buf, binary.LittleEndian, step.CustomTargetLow)        // custom_target_value_low
		binary.Write(&e.buf, binary.LittleEndian, step.CustomTargetHigh)       // custom_target_value_high
		e.buf.WriteByte(step.Intensity)                                         // intensity
	}
}

func (e *Encoder) writeDefinition(globalMesgNum uint16, fields []fieldDef) {
	localID := e.nextLocalID
	e.localMesgMap[globalMesgNum] = localID
	e.nextLocalID++

	// Record header: definition message (bit 6 set)
	e.buf.WriteByte(0x40 | localID)

	// Reserved byte
	e.buf.WriteByte(0)
	// Architecture (0 = little endian)
	e.buf.WriteByte(0)
	// Global message number
	binary.Write(&e.buf, binary.LittleEndian, globalMesgNum)
	// Number of fields
	e.buf.WriteByte(uint8(len(fields)))

	for _, f := range fields {
		e.buf.WriteByte(f.num)
		e.buf.WriteByte(f.size)
		e.buf.WriteByte(f.baseType)
	}
}

type fieldDef struct {
	num      uint8
	size     uint8
	baseType uint8
}

// Step represents a FIT workout step
type Step struct {
	IsRepeat        bool
	DurationValue   uint32 // milliseconds for time steps, step index for repeats
	TargetValue     uint32 // 0 for power custom target, repeat count for repeats
	CustomTargetLow uint32 // % FTP (e.g., 80 = 80% FTP)
	CustomTargetHigh uint32 // % FTP
	Intensity       uint8
}

func fitString(s string, maxLen int) []byte {
	if len(s) >= maxLen {
		s = s[:maxLen-1]
	}
	b := make([]byte, len(s)+1)
	copy(b, s)
	b[len(s)] = 0 // null terminator
	return b
}

// CRC-16 lookup table for FIT files
var crc16Table = [16]uint16{
	0x0000, 0xCC01, 0xD801, 0x1400, 0xF001, 0x3C00, 0x2800, 0xE401,
	0xA001, 0x6C00, 0x7800, 0xB401, 0x5000, 0x9C01, 0x8801, 0x4400,
}

func crc16(data []byte) uint16 {
	crc := uint16(0)
	for _, b := range data {
		tmp := crc16Table[crc&0xF]
		crc = (crc >> 4) & 0x0FFF
		crc = crc ^ tmp ^ crc16Table[b&0xF]

		tmp = crc16Table[crc&0xF]
		crc = (crc >> 4) & 0x0FFF
		crc = crc ^ tmp ^ crc16Table[(b>>4)&0xF]
	}
	return crc
}
