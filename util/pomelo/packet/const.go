package packet

import "errors"

const (
	_ Type = iota
	// Handshake represents a handshake: request(client) <====> handshake response(server)
	Handshake = 0x01

	// HandshakeAck represents a handshake ack from client to server
	HandshakeAck = 0x02

	// Heartbeat represents a heartbeat
	Heartbeat = 0x03

	// Data represents a common data packet
	Data = 0x04

	// Kick represents a kick off packet
	Kick = 0x05 // disconnect message from server
)

// ErrWrongPacketType represents a wrong packet type.
var ErrWrongPacketType = errors.New("wrong packet type")

// ErrInvalidHeader represents an invalid header
var ErrInvalidHeader = errors.New("invalid header")
