package types

import (
	"fmt"
	"time"
)

// TicketType determines ticket access type
type TicketType string

const (
	// TicketTypeRead is for read
	TicketTypeRead TicketType = "read"
	// TicketTypeWrite is for write
	TicketTypeWrite TicketType = "write"
)

// IRODSTicket contains irods ticket information
type IRODSTicket struct {
	ID int64
	// Name is ticket string
	Name string
	// Type is access type
	Type TicketType
	// Owner has the owner's name
	Owner string
	// ObjectType is type of object
	ObjectType ObjectType
	// Path is path to the object
	Path string
	// ExpireTime is time that the ticket expires
	ExpireTime time.Time
	// UsesLimit is an access limit
	UsesLimit int64
	// UsesCount is an access count
	UsesCount int64
	// WriteFileLimit is a write file limit
	WriteFileLimit int64
	// WriteFileCount is a write file count
	WriteFileCount int64
	// WriteByteLimit is a write byte limit
	WriteByteLimit int64
	// WriteByteCount is a write byte count
	WriteByteCount int64
}

// ToString stringifies the object
func (ticket *IRODSTicket) ToString() string {
	return fmt.Sprintf("<IRODSTicket %d %s %s %s>", ticket.ID, ticket.Name, ticket.Owner, ticket.Path)
}

// IRODSTicketForAnonymousAccess contains minimal irods ticket information for anonymous access
type IRODSTicketForAnonymousAccess struct {
	ID int64
	// Name is ticket string
	Name string
	// Type is access type
	Type TicketType
	// Path is path to the object
	Path string
	// ExpireTime is time that the ticket expires
	ExpireTime time.Time
}

// ToString stringifies the object
func (ticket *IRODSTicketForAnonymousAccess) ToString() string {
	return fmt.Sprintf("<IRODSTicketForAnonymousAccess %d %s %s>", ticket.ID, ticket.Name, ticket.Path)
}