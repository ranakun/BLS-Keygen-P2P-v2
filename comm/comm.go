package comm

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const N = 4

var PeerList [N - 1]Peer

type Peer struct {
	ID   int
	Name string
}

func (peer *Peer) Init(ID int) {
	peer.ID = ID
	peer.Name = fmt.Sprint("peer_", ID)
	count := 0
	for i := 1; i <= N; i++ {
		if i == peer.ID {
			continue
		}
		PeerList[count] = Peer{i, fmt.Sprint("peer_", i)}
		count++
	}
}

func GetMessages(str string) map[string][]string {
	var d = make(map[string][]string)
	s := strings.Split(str, "\n")
	for i := 0; i < len(s); i++ {
		if i == 0 {
			continue
		}
		res := strings.Split(s[i], ">")
		d[res[0]] = append(d[res[0]], res[1])
	}

	//fmt.Println(d)
	return d
}

func (peer Peer) ReadPeerInfoFromFile() map[string][]string {
	f, err := os.Open("./tmp/" + peer.Name + ".txt")
	if err != nil {
		log.Fatal(err)
	}

	var d = make(map[string][]string)

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)

	i := 0
	for scanner.Scan() {
		if i == 0 {
			i += 1
			continue
		}
		res := strings.Split(scanner.Text(), ">")
		d[res[0]] = append(d[res[0]], res[1])
		i += 1
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	return d
}

func (sender Peer) Send(receiver Peer, message string) {
	filename := "./tmp/" + receiver.Name + ".txt"

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(sender.Name + ">" + message + "\n"); err != nil {
		panic(err)
	}
}

func (sender Peer) Broadcast(message string) {
	for _, peer := range PeerList {
		sender.Send(peer, message)
	}
}

func (peer Peer) ReadLocalStorage() (string, error) {
	content, err := os.ReadFile("./tmp/" + peer.Name + ".txt")
	return string(content), err
}

func (peer Peer) WriteLocalStorage(message string) {
	filename := "./tmp/" + peer.Name + ".txt"

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(message + "\n"); err != nil {
		panic(err)
	}
}

func (peer Peer) ReadInbox() (string, error) {
	content, err := os.ReadFile("./tmp/" + peer.Name + ".txt")
	return string(content), err
}
