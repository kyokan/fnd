package p2p

import (
	"io"
	"net"
	"time"
)

type DummyTCPConn struct {
	Reader io.Reader
	Writer io.Writer
	Closed bool
}

func (d *DummyTCPConn) Read(b []byte) (n int, err error) {
	return d.Reader.Read(b)
}

func (d *DummyTCPConn) Write(b []byte) (n int, err error) {
	return d.Writer.Write(b)
}

func (d *DummyTCPConn) Close() error {
	rc, ok := d.Reader.(io.ReadCloser)
	if ok {
		if err := rc.Close(); err != nil {
			return err
		}
	}
	wc, ok := d.Writer.(io.WriteCloser)
	if ok {
		if err := wc.Close(); err != nil {
			return err
		}
	}
	d.Closed = true
	return nil
}

func (d *DummyTCPConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 50000,
		Zone: "",
	}
}

func (d *DummyTCPConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 9097,
		Zone: "",
	}
}

func (d *DummyTCPConn) SetDeadline(t time.Time) error {
	return nil
}

func (d *DummyTCPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (d *DummyTCPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
