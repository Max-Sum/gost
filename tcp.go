package gost

import (
	"net"
	
	"github.com/libp2p/go-reuseport"
)

// tcpTransporter is a raw TCP transporter.
type tcpTransporter struct{}

// TCPTransporter creates a raw TCP client.
func TCPTransporter() Transporter {
	return &tcpTransporter{}
}

func (tr *tcpTransporter) Dial(addr string, options ...DialOption) (net.Conn, error) {
	opts := &DialOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.Chain != nil {
		return opts.Chain.Dial(addr, SrcAddrChainOption(opts.SrcAddr))
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DialTimeout
	}
	
	d := net.Dialer{
		Timeout: timeout,
		Control: reuseport.Control,
	}
	if len(opts.SrcAddr) == 0 {
		return d.Dial("tcp", addr)
	}
	var conn net.Conn
	var err error
	for _, srcAddr := range opts.SrcAddr {
		d.LocalAddr, err = net.ResolveTCPAddr("tcp", srcAddr)
		if err != nil {
			continue
		}
		conn, err = d.Dial("tcp", addr)
		if err == nil {
			break
		}
	}
	return conn, err
}

func (tr *tcpTransporter) Handshake(conn net.Conn, options ...HandshakeOption) (net.Conn, error) {
	return conn, nil
}

func (tr *tcpTransporter) Multiplex() bool {
	return false
}

type tcpListener struct {
	net.Listener
}

// TCPListener creates a Listener for TCP proxy server.
func TCPListener(addr string) (Listener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &tcpListener{Listener: tcpKeepAliveListener{ln}}, nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(KeepAliveTime)
	return tc, nil
}
