package main

//Message Struct
type message struct {
	Phase int
	Name  string
	Value string
	To    string
}

type ack_message struct {
	Phase int
}

var ack_msg ack_message
