package srv

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/srv/proto"
)

type Client struct {
	timeout int
	port    int
}

func NewClient(port, timeout int) *Client {
	return &Client{port: port, timeout: timeout}
}

func (c *Client) Connect(ip string) (*grpc.ClientConn, proto.MizarServiceClient, context.Context, context.CancelFunc, error) {
	grpcHostURL := fmt.Sprintf("%s:%d", ip, c.port)
	conn, err := grpc.Dial(grpcHostURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return conn, nil, nil, nil, err
	}
	client := proto.NewMizarServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	return conn, client, ctx, cancel, nil
}
