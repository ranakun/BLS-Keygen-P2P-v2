package main

type test_struct struct {
	Peer_list string `json:"peer_list"`
}

type test_struct_2 struct {
	Pvt_key   string `json:"pvt_key"`
	Peer_list string `json:"peer_list"`
	T         string `json:"t"`
	N         string `json:"n"`
}
type gen_share struct {
	Pvt  string `json:"pvt"`
	List string `json:"peer_list"`
	T    int    `json:"t"`
}

var start_p2p_flag = 0

var debug = true
