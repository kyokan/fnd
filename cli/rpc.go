package cli

import (
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

func DialRPC(cmd *cobra.Command) (*grpc.ClientConn, error) {
	rpcHost, _ := cmd.Flags().GetString(FlagRPCHost)
	rpcPort, _ := cmd.Flags().GetInt(FlagRPCPort)
	return grpc.Dial(net.JoinHostPort(rpcHost, strconv.Itoa(rpcPort)), grpc.WithInsecure())
}
