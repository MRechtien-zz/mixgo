package qu

import (
	"fmt"
	"net"
)

type Mixer struct {
	Output chan []byte
}

func NewMixer() Mixer {
	mixer := Mixer{
		Output: make(chan []byte),
	}
	go sendToMixer(mixer.Output)
	return mixer
}

func sendToMixer(output chan []byte) {
	conn, err := net.Dial("tcp", "192.168.0.150:51325")
	if err != nil {
		fmt.Println("could not connect to TCP server: ", err)
	}
	for message := range output {
		fmt.Printf("Sending message to mixer: %v\n", message)
		conn.Write(message)
	}
	defer conn.Close()
}
