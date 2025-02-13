package protocols

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
)

func readRFB(conn net.Conn, logger Logger) error {
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		logger.Error(fmt.Sprintf("[rfb     ] error: %v", err))
		return err
	}
	logger.Info(fmt.Sprintf("[rfb     ] message %q", msg))
	return nil
}

// PixelFormat represents a RFB communication unit
type PixelFormat struct {
	Width, Heigth                   uint16
	BPP, Depth                      uint8
	BigEndian, TrueColour           uint8 // flags; 0 or non-zero
	RedMax, GreenMax, BlueMax       uint16
	RedShift, GreenShift, BlueShift uint8
	Padding                         [3]uint8
	ServerNameLength                int32
}

// HandleRFB takes a net.Conn and does basic RFB/VNC communication
func HandleRFB(ctx context.Context, conn net.Conn, logger Logger, h Honeypot) error {
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(fmt.Sprintf("[rfb     ] error: %v", err))
		}
	}()

	if _, err := conn.Write([]byte("RFB 003.008\n")); err != nil {
		return err
	}
	if err := readRFB(conn, logger); err != nil {
		return err
	}
	var authNone uint32 = 1
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, authNone)
	if _, err := conn.Write(bs); err != nil {
		return err
	}

	serverName := "rfb-go"
	lenName := int32(len(serverName))

	buf := new(bytes.Buffer)
	f := PixelFormat{
		Width:            1,
		Heigth:           1,
		BPP:              16,
		Depth:            16,
		BigEndian:        0,
		TrueColour:       1,
		RedMax:           0x1f,
		GreenMax:         0x1f,
		BlueMax:          0x1f,
		RedShift:         0xa,
		GreenShift:       0x5,
		BlueShift:        0,
		ServerNameLength: lenName,
	}
	if err := binary.Write(buf, binary.LittleEndian, f); err != nil {
		return err
	}
	if _, err := conn.Write(buf.Bytes()); err != nil {
		return err
	}
	return readRFB(conn, logger)
}
