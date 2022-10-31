package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
)

func keygen_Stream_listener(h host.Host) {
	//fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	//Return Channel details
	h.SetStreamHandler("/keygen/0.0.1", func(s network.Stream) {
		//log.Println("sender received new stream")
		if err := process_input(s, h); err != nil {
			log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}

	})
	// 'stream' will stay open until you close it (or the other side closes it).

}

func WriteLocalStorage(message string) {
	filename := "peer_" + strconv.Itoa(my_index) + ".txt"

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err = f.WriteString(message + "\n"); err != nil {
		log.Println(err)
	}
}

func process_input(s network.Stream, h host.Host) error {

	//log.Println(s)
	buf := bufio.NewReader(s)
	//log.Println(s)
	str, err := buf.ReadBytes('\n')
	if err != nil {
		log.Println(err)
		return err
	}
	bytes := []byte(str)
	var message_receive message
	json.Unmarshal(bytes, &message_receive)
	//log.Println(s.Conn().RemotePeer())

	//Check and rediect :
	//sender_id := s.ID()[1 : len(s.ID())-2]

	// if message_receive.Phase == 0 {
	// 	log.Println("Got ACK: ", message_receive.Value, " from ", s.Conn().RemotePeer())
	// 	fmt.Println("Phase 0")
	// 	// WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
	// 	acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	// } else
	if message_receive.Phase == 1 {
		log.Println("Got epk_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
		fmt.Println("Phase 1")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	} else if message_receive.Phase == 2 {
		log.Println("Got bpk_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
		fmt.Println("Phase 2")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	} else if message_receive.Phase == 3 {
		log.Println("Got kgc_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
		fmt.Println("Phase 3")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	} else if message_receive.Phase == 4 {
		log.Println("Got spub_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
		fmt.Println("Phase 4")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	} else if message_receive.Phase == 5 {
		log.Println("Got kgd_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
		fmt.Println("Phase 5")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	} else if message_receive.Phase == 6 {
		log.Println("Got shares: ", message_receive.Value, " from ", s.Conn().RemotePeer())
		fmt.Println("Phase 6")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	} else if message_receive.Phase == 7 {
		log.Println("FIN")
		WriteLocalStorage(s.Conn().RemotePeer().String() + ">" + message_receive.Value)
		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)
	}

	_, err = s.Write([]byte(""))
	return err
}
