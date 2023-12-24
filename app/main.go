package main

import (
	"encoding/binary"
	"fmt"

	// "io"
	"strings"

	"net"
)

type Message struct {
	DnsHeader      *Header
	Question       *DNSQuestion
	ResourceRecord *ResourceRecord
}

func (header *Message) serialize() []byte {
	headerBytes := header.DnsHeader.serialize()
	QuestionBytes := header.Question.serialize()
	AnswerBytes := header.ResourceRecord.serialize()
	result := append(headerBytes, QuestionBytes...)
	return append(result, AnswerBytes...)
}

type Header struct {
	ID      uint16
	Flags   DnsMsgFlags
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

func (msg *Header) serialize() []byte {
	result := make([]byte, 12)

	binary.BigEndian.PutUint16(result[:2], msg.ID)
	var flags uint16

	if msg.Flags.QR { // the Id is 16 bits
		flags |= 1 << 15 // this means 1 shifted 15 0s to the right ( or 1  multiplies by 2 15 times )
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

func (m *Message) DecodeMsg(data []byte) error {
	m.DnsHeader = &Header{}
	err := m.DnsHeader.DecodeHeader(data[:12])
	if err != nil {
		fmt.Println("Error deconing Header ", err)
		return err
	}

	return nil
}

func (m *Header) DecodeHeader(data []byte) error {
	m.ID = binary.BigEndian.Uint16(data[:2])
	flags := binary.BigEndian.Uint16(data[2:4])
	m.Flags.QR = flags>>15 != 0
	m.Flags.OPCode = uint8(flags >> 11)
	m.Flags.AA = flags>>10 != 0
	m.Flags.TC = flags>>9 != 0
	m.Flags.RD = flags>>8 != 0
	m.Flags.RA = flags>>7 != 0
	m.Flags.Z = uint8(flags >> 4)
	// m.Flags.Rcode = m.Flags.Rcode
	// fmt.Println(m.Flags.Rcode)
	m.QDCount = binary.BigEndian.Uint16(data[4:6])
	m.ANCount = binary.BigEndian.Uint16(data[6:8])
	m.NSCount = binary.BigEndian.Uint16(data[8:10])
	m.ARCount = binary.BigEndian.Uint16(data[10:12])

	return nil
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
	Type  int // 2 bytes
	Class int //2 bytes
}

// Serialize the question section into array of bytes
func (question *DNSQuestion) serialize() []byte {
	questionAdd := make([]byte, 4) // question in 4 bytes long + the domain name
	binary.BigEndian.PutUint16(questionAdd[:2], uint16(question.Type))
	binary.BigEndian.PutUint16(questionAdd[2:4], uint16(question.Class))
	result := append(LabelSequence(question.Name), questionAdd...)
	return result

}

type ResourceRecord struct {
	Name   string
	Type   uint16 //2 bytes
	Class  uint16 //2 bytes
	TTL    uint32 // 4 bytes
	Length uint16 // 2 bytes
	Data   uint32 // 4 bytes
}

// Serialize the Answer section into array of bytes
// it return [14 + domain name ] bytes array
func (Answer *ResourceRecord) serialize() []byte {
	questionAdd := make([]byte, 4)
	binary.BigEndian.PutUint16(questionAdd[:2], uint16(Answer.Type))
	binary.BigEndian.PutUint16(questionAdd[2:4], uint16(Answer.Class))
	result := append(LabelSequence(Answer.Name), questionAdd...)
	var answerData = make([]byte, 10)

	binary.BigEndian.PutUint32(answerData[:4], Answer.TTL)
	binary.BigEndian.PutUint16(answerData[4:6], Answer.Length)
	binary.BigEndian.PutUint32(answerData[6:10], Answer.Data)
	return append(result, answerData...)

}

// Convers the domain name into an array of bytes teminated with null \x00
func LabelSequence(q string) []byte {
	labels := strings.Split(q, ".") //lable : google.com
	var sequence []byte
	for _, lable := range labels {
		sequence = append(sequence, byte(len(lable))) // len(google) /x06
		sequence = append(sequence, lable...)         // google
	}
	sequence = append(sequence, '\x00') // terminate the lable with \x00
	return sequence
}

func getRcode(m *Message) uint8 {
	if m.DnsHeader.Flags.Rcode == 0 {
		return 0
	}
	return 4
}

// Create a response
func CreateResponse(req *Message) *Message {
	fmt.Println(req.DnsHeader.Flags.Rcode)
	return &Message{
		DnsHeader: &Header{
			ID: req.DnsHeader.ID,
			Flags: DnsMsgFlags{
				QR:     true,
				OPCode: req.DnsHeader.Flags.OPCode,
				AA:     false,
				TC:     false,
				RD:     req.DnsHeader.Flags.RD,
				RA:     false,
				Z:      0x0,
				Rcode:  getRcode(req),
			},
			QDCount: 0x1,
			ANCount: 0x1,
			NSCount: 0x0,
			ARCount: 0x0,
		},
		Question: &DNSQuestion{
			Name:  "codecrafters.io",
			Type:  1,
			Class: 1,
		},
		ResourceRecord: &ResourceRecord{
			Name:   "codecrafters.io",
			Type:   1,
			Class:  1,
			TTL:    60,
			Length: 4,
			Data:   0,
		},
	}
}

// 0 :Id  0000: Opcode  0: AA  0: TC  0: RD  0:RA  000:Z   0000 : Rcode

func main() {

	fmt.Println("Logs from your program will appear here!")

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

		// receivedData := string(buf[:size])
		// fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)
		m := &Message{}
		err = m.DecodeMsg(buf[:size])
		if err != nil {
			fmt.Println("error parsing message", err)
		}

		response := CreateResponse(m).serialize()
		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}

	}
}
