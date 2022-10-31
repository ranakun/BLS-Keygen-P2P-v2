package main

// import (
// 	"bufio"
// 	"encoding/json"
// 	"log"

// 	"github.com/libp2p/go-libp2p-core/host"
// 	"github.com/libp2p/go-libp2p-core/network"
// )

// func keygen_Stream_listener(h host.Host) {
// 	//fmt.Println("Got a new stream!")

// 	// Create a buffer stream for non blocking read and write.
// 	//Return Channel details
// 	h.SetStreamHandler("/keygen/0.0.1", func(s network.Stream) {
// 		//log.Println("sender received new stream")
// 		if err := process_input(s, h); err != nil {
// 			log.Println(err)
// 			s.Reset()
// 		} else {
// 			s.Close()
// 		}

// 	})
// 	// 'stream' will stay open until you close it (or the other side closes it).

// }

// func process_input(s network.Stream, h host.Host) error {

// 	//log.Println(s)
// 	buf := bufio.NewReader(s)
// 	//log.Println(s)
// 	str, err := buf.ReadBytes('\n')
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	bytes := []byte(str)
// 	var message_receive message
// 	json.Unmarshal(bytes, &message_receive)
// 	//log.Println(s.Conn().RemotePeer())

// 	//Check and rediect :
// 	//sender_id := s.ID()[1 : len(s.ID())-2]

// 	if message_receive.Phase == 1 {

// 		log.Println("Got ppk_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
// 		//Index peer_index[s.Conn().RemotePeer().String()] use this instead of sort.Search()
// 		log.Println(s.Conn().RemotePeer().String())
// 		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)

// 	} else if message_receive.Phase == 2 {
// 		log.Println("Got Kgc_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
// 		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)

// 	} else if message_receive.Phase == 3 {
// 		log.Println("Got Kgd_j: ", message_receive.Value, " from ", s.Conn().RemotePeer())
// 		acknowledge(s.Conn().RemotePeer().String(), message_receive.Phase, h)

// 	}

// 	_, err = s.Write([]byte(""))
// 	return err
// }
