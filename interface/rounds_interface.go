package rounds_interface

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
var Peer_details_list []string

// var round = make(map[string]int)
var Peer_map = make(map[string]string)

var Sorted_peer_id []string
var My_index int = 0

//Store the current phase value received
var Sent_peer_phase = make(map[string]int)

//Store the phase value of peer acknowledgement
var Receive_peer_phase = make(map[string]int)

//Lock map to avoid concurrent map writes
var L = sync.Mutex{}

type P2P struct {

	// Represents the libp2p host
	Host      host.Host
	Host_ip   string
	Host_peer string
	Ctx       context.Context
	Peers     []string
	Round     int
}

var P2p P2P

//Rework flags and channels to conform to this struct
type Status struct {
	Phase     int
	Chan      string
	Num_peers int
}

var T_array []int

type Round1_Data struct {
	EPK_j map[string]string
	EPK_i curves.Point
	ESK_i curves.Scalar
	Curve *curves.Curve
}

type Round2_Data struct {
	BPK_i  kyber.Point
	BPK_j  map[string]string
	Shares []*share.PriShare
	Suite  *bn256.Suite
}

type Round3_Data struct {
	FOfi map[string]string
	C1   curves.Point
	C2   string
	C3   []byte
}

var Round1_data Round1_Data
var Round2_data Round2_Data
var Round3_data Round3_Data

var Status_struct Status
var All_ok = true

var Peer_index = make(map[string]int)
