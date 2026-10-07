package main

// keyInput matches the Windows INPUT struct (40 bytes on 64-bit) for a keyboard event.
type keyInput struct {
	Type  uint32
	_     uint32
	Vk    uint16
	Scan  uint16
	Flags uint32
	Time  uint32
	_     uint32
	Extra uintptr
	_     uint64
}
