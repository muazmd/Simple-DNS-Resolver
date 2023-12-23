package main

import (
	"encoding/binary"
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
)

type Message struct {
	DnsHeader Header
	Question  DNSQuestion
}

type Header struct {
	ID      uint16
	Flags   DnsMsgFlags
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type DnsMsgFlags struct {
	QR     bool
	OPCode uint8
	AA     bool
	TC     bool
	RD     bool
	RA     bool
	Z      uint8
	Rcode  uint8
}
type DNSQuestion struct {
	Name  string
	Type  int
	Class int
}

func (question DNSQuestion) serialize() []byte {
	result := LabelSequence(string(question.Name))
	questionAdd := make([]byte, 4)
	binary.BigEndian.PutUint16(questionAdd[:2], uint16(question.Type))
	binary.BigEndian.PutUint16(questionAdd[2:4], uint16(question.Class))
	result = append(result, questionAdd...)
	return result

}

func LabelSequence(label string) []byte {
	labels := strings.Split(label, ".") //lable : google.com

	var sequence []byte
	for _, lable := range labels {
		sequence = append(sequence, byte(len(label))) // len(google) /x06
		sequence = append(sequence, lable...)         // google
	}
	sequence = append(sequence, '\x00') // terminate the lable with \x00
	return sequence
}

func CreateResponse() Message {
	return Message{
		DnsHeader: Header{
			ID: 1234,
			Flags: DnsMsgFlags{
				QR:     true,
				OPCode: 0x0,
				AA:     false,
				TC:     false,
				RD:     false,
				RA:     false,
				Z:      0x0,
				Rcode:  0x0,
			},
			QDCount: 0x1,
			ANCount: 0x0,
			NSCount: 0x0,
			ARCount: 0x0,
		},
		Question: DNSQuestion{
			Name:  "codecrafters.io",
			Type:  1,
			Class: 1,
		},
	}
}

// 0 :Id  0000: Opcode  0: AA  0: TC  0: RD  0:RA  000:Z   0000 : Rcode

func (header Message) serialize() []byte {
	headerBytes := header.DnsHeader.serialize()
	QuestionBytes := header.Question.serialize()
	return append(headerBytes, QuestionBytes...)
}

func (msg Header) serialize() []byte {
	result := make([]byte, 12)

	binary.BigEndian.PutUint16(result[:2], msg.ID)
	var flags uint16

	if msg.Flags.QR {
		flags |= 1 << 15
	}
	flags |= uint16(msg.Flags.OPCode) << 11

	if msg.Flags.AA {
		flags |= 1 << 10
	}
	if msg.Flags.TC {
		flags |= 1 << 9
	}
	if msg.Flags.RD {
		flags |= 1 << 8
	}
	if msg.Flags.RA {
		flags |= 1 << 7
	}

	flags |= uint16(msg.Flags.Z) << 4
	flags |= uint16(msg.Flags.Rcode)
	binary.BigEndian.PutUint16(result[2:4], flags)
	binary.BigEndian.PutUint16(result[4:6], msg.QDCount)
	binary.BigEndian.PutUint16(result[6:8], msg.ANCount)
	binary.BigEndian.PutUint16(result[8:10], msg.NSCount)
	binary.BigEndian.PutUint16(result[10:12], msg.ARCount)
	return result

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	//
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		// Create an empty response
		response := CreateResponse().serialize()

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
