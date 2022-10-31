package main

// package keygen

import (
	"context"
	"sync"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/libp2p/go-libp2p-core/host"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
)

//Discovered values
// type peer_details struct {
// 	id   peer.ID
// 	addr peer.AddrInfo
// }

// var peer_details_list []peer_details = make([]peer_details, 0, 10)
var peer_details_list []string

// var round = make(map[string]int)
var peer_map = make(map[string]string)

var sorted_peer_id []string
var my_index int = 0

//Store the current phase value received
var sent_peer_phase = make(map[string]int)

//Store the phase value of peer acknowledgement
var receive_peer_phase = make(map[string]int)

//Lock map to avoid concurrent map writes
var l = sync.Mutex{}

type P2P struct {

	// Represents the libp2p host
	Host      host.Host
	Host_ip   string
	Host_peer string
	Ctx       context.Context
	Peers     []string
	Round     int
}

var p2p P2P

//Rework flags and channels to conform to this struct
type Status struct {
	Phase     int
	Chan      string
	Num_peers int
}

type Round1_Data struct {
	EPK_j map[string]string
	EPK_i curves.Point
	ESK_i curves.Scalar
	curve *curves.Curve
}

type Round2_Data struct {
	BPK_i  kyber.Point
	BPK_j  map[string]string
	shares []*share.PriShare
	suite  *bn256.Suite
}

type Round3_Data struct {
	fOfi map[string]string
}

var round1_data Round1_Data
var round2_data Round2_Data
var round3_data Round3_Data

var status_struct Status
var all_ok = true

var peer_index = make(map[string]int)
