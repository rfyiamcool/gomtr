package icmp

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/rfyiamcool/gomtr/common"
)

const (
	ProtocolICMP     = 1  // Internet Control Message
	ProtocolIPv6ICMP = 58 // ICMP for IPv6
)

func Icmp(destAddr string, ttl, pid int, timeout time.Duration, seq int) (hop common.IcmpReturn, err error) {
	ip := net.ParseIP(destAddr)
	if ip == nil {
		return hop, fmt.Errorf("ip: %v is invalid", destAddr)
	}

	ipaddr := net.IPAddr{IP: ip}
	if p4 := ip.To4(); len(p4) == net.IPv4len {
		return icmpIpv4("0.0.0.0", &ipaddr, ttl, pid, timeout, seq) // need sudo permission
	} else {
		return icmpIpv6("::", &ipaddr, ttl, pid, timeout, seq)
	}
}

func icmpIpv4(localAddr string, dst net.Addr, ttl, pid int, timeout time.Duration, seq int) (hop common.IcmpReturn, err error) {
	hop.Success = false
	start := time.Now()
	c, err := icmp.ListenPacket("ip4:icmp", localAddr)
	if err != nil {
		return hop, err
	}
	defer c.Close()

	if err = c.IPv4PacketConn().SetTTL(ttl); err != nil {
		return hop, err
	}

	if err = c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return hop, err
	}

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(seq))
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID: pid, Seq: seq,
			Data: append(bs, 'x'),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		return hop, err
	}

	if _, err := c.WriteTo(wb, dst); err != nil {
		return hop, err
	}

	peer, _, err := listenForSpecific4(c, append(bs, 'x'), pid, seq)
	if err != nil {
		return hop, err
	}

	elapsed := time.Since(start)
	hop.Elapsed = elapsed
	hop.Addr = peer
	hop.Success = true
	return hop, err
}

func icmpIpv6(localAddr string, dst net.Addr, ttl, pid int, timeout time.Duration, seq int) (hop common.IcmpReturn, err error) {
	hop.Success = false
	start := time.Now()
	c, err := icmp.ListenPacket("ip6:ipv6-icmp", localAddr)
	if err != nil {
		return hop, err
	}
	defer c.Close()

	if err = c.IPv6PacketConn().SetHopLimit(ttl); err != nil {
		return hop, err
	}

	if err = c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return hop, err
	}

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(seq))
	wm := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID: pid, Seq: seq,
			Data: append(bs, 'x'),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return hop, err
	}

	if _, err := c.WriteTo(wb, dst); err != nil {
		return hop, err
	}

	peer, _, err := listenForSpecific6(c, append(bs, 'x'), pid, seq)
	if err != nil {
		return hop, err
	}

	elapsed := time.Since(start)
	hop.Elapsed = elapsed
	hop.Addr = peer
	hop.Success = true
	return hop, err
}

func listenForSpecific4(conn *icmp.PacketConn, neededBody []byte, needId, needSeq int) (string, []byte, error) {
	for {
		b := make([]byte, 1500)
		n, peer, err := conn.ReadFrom(b)
		if err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				return "", []byte{}, neterr
			}
		}
		if n == 0 {
			continue
		}

		x, err := icmp.ParseMessage(ProtocolICMP, b[:n])
		if err != nil {
			continue
		}

		if typ, ok := x.Type.(ipv4.ICMPType); ok && typ.String() == "time exceeded" {
			body := x.Body.(*icmp.TimeExceeded).Data

			x, _ := icmp.ParseMessage(ProtocolICMP, body[20:])
			switch x.Body.(type) {
			case *icmp.Echo:
				msg := x.Body.(*icmp.Echo)
				if msg.ID == needId && msg.Seq == needSeq {
					return peer.String(), []byte{}, nil
				}
			default:
				// ignore
			}
		}

		if typ, ok := x.Type.(ipv4.ICMPType); ok && typ.String() == "echo reply" {
			b, _ := x.Body.Marshal(ProtocolICMP)
			if string(b[4:]) != string(neededBody) || x.Body.(*icmp.Echo).ID != needId {
				continue
			}

			return peer.String(), b[4:], nil
		}
	}
}

func listenForSpecific6(conn *icmp.PacketConn, neededBody []byte, needId, needSeq int) (string, []byte, error) {
	for {
		b := make([]byte, 1500)
		n, peer, err := conn.ReadFrom(b)
		if err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				return "", []byte{}, neterr
			}
		}
		if n == 0 {
			continue
		}

		x, err := icmp.ParseMessage(ProtocolIPv6ICMP, b[:n])
		if err != nil {
			continue
		}

		if x.Type.(ipv6.ICMPType) == ipv6.ICMPTypeTimeExceeded {
			body := x.Body.(*icmp.TimeExceeded).Data
			x, _ := icmp.ParseMessage(ProtocolIPv6ICMP, body[40:])
			switch x.Body.(type) {
			case *icmp.Echo:
				msg := x.Body.(*icmp.Echo)
				if msg.ID == needId && msg.Seq == needSeq {
					return peer.String(), []byte{}, nil
				}
			default:
				// ignore
			}
		}

		if typ, ok := x.Type.(ipv6.ICMPType); ok && typ == ipv6.ICMPTypeEchoReply {
			b, _ := x.Body.Marshal(1)
			if string(b[4:]) != string(neededBody) || x.Body.(*icmp.Echo).ID != needId {
				continue
			}

			return peer.String(), b[4:], nil
		}
	}
}
