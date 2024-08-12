package swim

import (
	"context"
	"fmt"
	"log"
	"net"
)

type netTCP struct {
	protocolVersion uint8
	tcpLn           *net.TCPListener
	stream          func(conn net.Conn) error
}

func newNetTCP(port uint16, stream func(conn net.Conn) error) (*netTCP, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, fmt.Errorf("resolve tcp addr: %w", err)
	}
	tcpLn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}

	return &netTCP{
		tcpLn:  tcpLn,
		stream: stream,
	}, nil
}

func (nt *netTCP) listen(ctx context.Context) error {
	defer func() {
		_ = nt.tcpLn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("listen tcp: %w", ctx.Err())
		default:
			conn, err := nt.tcpLn.AcceptTCP()
			if err != nil {
				return fmt.Errorf("accept tcp: %w", err)
			}
			go nt.handleConn(ctx, conn)
		}
	}
}

func (nt *netTCP) handleConn(ctx context.Context, conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		err = nt.stream(conn)
	}
	if err != nil {
		log.Printf("handle conn: %s", err)
	}
}

func sendToTCP(ctx context.Context, addr net.Addr, msg []byte) error {
	var nd net.Dialer
	conn, err := nd.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return fmt.Errorf("dial to addr: %w", err)
	}
	defer func() { _ = conn.Close() }()

	select {
	case <-ctx.Done():
		return fmt.Errorf("send msg: %w", ctx.Err())
	default:
		if _, err = conn.Write(msg); err != nil {
			return fmt.Errorf("write to conn: %w", err)
		}
	}
	return nil
}
