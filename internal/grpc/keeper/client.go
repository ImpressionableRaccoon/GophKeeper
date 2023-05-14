package keeper

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

type Client struct {
	conn *grpc.ClientConn
	pb.KeeperClient
}

// NewClient - создаем новый grpc клиент.
func NewClient(serverAddress string) (*Client, error) {
	c := &Client{}

	config := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, //nolint:gosec
	}

	var err error
	c.conn, err = grpc.Dial(serverAddress, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		return nil, err
	}

	c.KeeperClient = pb.NewKeeperClient(c.conn)

	return c, nil
}

// Close - закрываем соединение с сервером.
func (s *Client) Close() error {
	return s.conn.Close()
}
