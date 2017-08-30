package server

import (
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSH struct {
	*Test
	PrivateKey    ssh.Signer
	ServerAddress string
	conn          *ssh.ServerConn
}

type Test struct {
	ReceivedKeys []ssh.PublicKey
}

func NewSSH() *SSH {
	return &SSH{}
}

func (s *SSH) Start() {
	config := &ssh.ServerConfig{
		PublicKeyCallBack: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permission, error) {
			s.Test.ReceivedKeys = append(s.Test.ReceivedKeys, pubKey)
		},
	}

	config.AddHostKey(s.PrivateKey)

	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	s.ServerAddress = l.Addr().String()

	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("failed to accept connections: %s", err)
	}

	conn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Fatalf("failed creating ssh server connections: %s", err)
	}

	s.conn = conn

	go ssh.DiscardRequests(reqs)

	for c := range chans {
	}
}

func (s *SSH) Stop() {
	err := s.conn.Close()
	log.Fatalf("failed shutting down ssh connection: %s", err)
}
