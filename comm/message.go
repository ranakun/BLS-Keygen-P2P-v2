package comm

//Message Struct
type Message struct {
	Phase int
	Name  string
	Value string
	To    string
}

type Ack_message struct {
	Phase int
}

var ack_msg Ack_message
