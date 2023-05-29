package keeper

import (
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

// Client - клиент для взаимодействия с сервером.
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
		return nil, fmt.Errorf("grpc keeper NewClient: dial: %w", err)
	}

	c.KeeperClient = pb.NewKeeperClient(c.conn)

	return c, nil
}

// Close - закрываем соединение с сервером.
func (s *Client) Close() error {
	err := s.conn.Close()
	if err != nil {
		return fmt.Errorf("grpc keeper Client close: %w", err)
	}

	return nil
}
