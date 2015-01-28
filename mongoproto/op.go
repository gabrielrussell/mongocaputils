package mongoproto

import (
	"fmt"
	"io"
)

// ErrNotMsg is returned if a provided buffer is too small to contain a Mongo message
var ErrNotMsg = fmt.Errorf("buffer is too small to be a Mongo message")

// Op is a Mongo operation
type Op interface {
	OpCode() OpCode
	FromReader(io.Reader) error
}

// ErrUnknownOpcode is an error that represents an unrecognized opcode.
type ErrUnknownOpcode int

func (e ErrUnknownOpcode) Error() string {
	return fmt.Sprintf("Unknown opcode %d", e)
}

// OpFromReader reads an Op from an io.Reader
func OpFromReader(r io.Reader) (Op, error) {
	b := make([]byte, MsgHeaderLen)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	var m MsgHeader
	m.fromWire(b)

	var result Op
	switch m.OpCode {
	case OpCodeQuery:
		result = &OpQuery{Header: m}
	case OpCodeReply:
		result = &OpReply{Header: m}
	case OpCodeGetMore:
		result = &OpGetMore{Header: m}
	case OpCodeInsert:
		result = &OpInsert{Header: m}
	default:
		result = &OpUnknown{Header: m}
	}
	err = result.FromReader(r)
	return result, err
}
