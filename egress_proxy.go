package ice

import (
	"net"
)

type PacketConnProxy interface {
	SendPacket(laddr net.Addr, raddr net.Addr, data []byte) (int, error)
}

// Help to implement a custom egress proxy
type ProxiedEgressPacketConn struct {
	net.PacketConn
	proxy     PacketConnProxy
	muxedConn *udpMuxedConn
}

func NewProxiedEgressPacketConn(conn net.PacketConn, proxy PacketConnProxy) net.PacketConn {
	proxiedConn := &ProxiedEgressPacketConn{}
	proxiedConn.PacketConn = conn
	proxiedConn.proxy = proxy
	if c, ok := conn.(*udpMuxedConn); ok {
		proxiedConn.muxedConn = c
	}
	return proxiedConn
}

func (proxiedConn *ProxiedEgressPacketConn) WriteTo(data []byte, raddr net.Addr) (n int, err error) {
	laddr := proxiedConn.PacketConn.LocalAddr()
	if proxiedConn.muxedConn != nil {
		// each time we write to a new address, we'll register it with the mux
		addr := raddr.String()
		if !proxiedConn.muxedConn.containsAddress(addr) {
			proxiedConn.muxedConn.addAddress(addr)
		}
	}
	return proxiedConn.proxy.SendPacket(laddr, raddr, data)
}
