package src

import (
	"net"
	"time"
)

type Peer struct {
	Name        string       `json:"uid"`
	Address     *net.UDPAddr `json:"address"`
	LastHeardOf time.Time
}

func NewPeer(name string, addr *net.UDPAddr) *Peer {
	return &Peer{
		Name:        name,
		Address:     addr,
		LastHeardOf: time.Now(),
	}
}

func (p *Peer) String() string {
	return p.Address.String()
}

func (p *Peer) UpdateLastSeen() {
	p.LastHeardOf = time.Now()
}
