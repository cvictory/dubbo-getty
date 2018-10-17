/******************************************************
# DESC    : echo package
# AUTHOR  : Alex Stocks
# LICENCE : Apache License 2.0
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-08-22 17:44
# FILE    : echo.go
******************************************************/

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"
)

import (
	log "github.com/AlexStocks/log4go"
)

////////////////////////////////////////////
//  echo command
////////////////////////////////////////////

type echoCommand uint32

const (
	heartbeatCmd = iota
	echoCmd
)

var echoCommandStrings = [...]string{
	"heartbeat",
	"echo",
}

func (c echoCommand) String() string {
	return echoCommandStrings[c]
}

////////////////////////////////////////////
// EchoPkgHandler
////////////////////////////////////////////

const (
	echoPkgMagic     = 0x20160905
	maxEchoStringLen = 0xff

	echoHeartbeatRequestString  = "ping"
	echoHeartbeatResponseString = "pong"
)

var (
	ErrNotEnoughStream = errors.New("packet stream is not enough")
	ErrTooLargePackage = errors.New("package length is exceed the echo package's legal maximum length.")
	ErrIllegalMagic    = errors.New("package magic is not right.")
)

var (
	echoPkgHeaderLen int
)

func init() {
	echoPkgHeaderLen = (int)((uint)(unsafe.Sizeof(EchoPkgHeader{})))
}

type EchoPkgHeader struct {
	Magic uint32
	LogID uint32 // log id

	Sequence  uint32 // request/response sequence
	ServiceID uint32 // service id

	Command uint32 // operation command code
	Code    int32  // error code

	Len uint16 // body length
	_   uint16
	_   int32 // reserved, maybe used as package md5 checksum
}

type EchoPackage struct {
	H EchoPkgHeader
	B string
}

func (p EchoPackage) String() string {
	return fmt.Sprintf("log id:%d, sequence:%d, command:%s, echo string:%s",
		p.H.LogID, p.H.Sequence, (echoCommand(p.H.Command)).String(), p.B)
}

func (p EchoPackage) Marshal() (*bytes.Buffer, error) {
	var (
		err error
		buf *bytes.Buffer
	)

	buf = &bytes.Buffer{}
	err = binary.Write(buf, binary.LittleEndian, p.H)
	if err != nil {
		return nil, err
	}
	buf.WriteByte((byte)(len(p.B)))
	buf.WriteString(p.B)

	return buf, nil
}

func (p *EchoPackage) Unmarshal(buf *bytes.Buffer) (int, error) {
	var (
		err error
		len byte
	)

	if buf.Len() < echoPkgHeaderLen {
		return 0, ErrNotEnoughStream
	}

	// header
	err = binary.Read(buf, binary.LittleEndian, &(p.H))
	if err != nil {
		return 0, err
	}
	if p.H.Magic != echoPkgMagic {
		log.Error("@p.H.Magic{%x}, right magic{%x}", p.H.Magic, echoPkgMagic)
		return 0, ErrIllegalMagic
	}
	if buf.Len() < (int)(p.H.Len) {
		return 0, ErrNotEnoughStream
	}
	if maxEchoStringLen < p.H.Len-1 {
		return 0, ErrTooLargePackage
	}

	len, err = buf.ReadByte()
	if err != nil {
		return 0, nil
	}
	p.B = (string)(buf.Next((int)(len)))
	//if strings.HasPrefix(p.B, "Hello, getty!") {
	//	gxlog.CError("idx:%d, body:%s", idx, p.B)
	//	idx++
	//}

	return (int)(p.H.Len) + echoPkgHeaderLen, nil
}