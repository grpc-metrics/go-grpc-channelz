package prometheus

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/grpc_channelz_v1"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"strconv"
)

const (
	namespace = "go_grpc"
	subsystem = "channelz"
)

type ChannelzMetrics struct {
	ServerCallsStarted   *prometheus.Desc
	ServerCallsSucceeded *prometheus.Desc
	ServerCallsFailed    *prometheus.Desc

	channelzClient grpc_channelz_v1.ChannelzClient
	logWriter      io.Writer
}

func NewChannelzMetrics(grpcServerAddress string, logWriter io.Writer) *ChannelzMetrics {
	cc, err := grpc.Dial(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return &ChannelzMetrics{
		ServerCallsStarted:   newDescriptor("server_calls_started", "Total number of calls started", nil, "server_id"),
		ServerCallsSucceeded: newDescriptor("server_calls_succeeded", "Total number of calls succeeded", nil, "server_id"),
		ServerCallsFailed:    newDescriptor("server_calls_failed", "Total number of calls failed", nil, "server_id"),

		channelzClient: grpc_channelz_v1.NewChannelzClient(cc),
		logWriter:      logWriter,
	}
}

func (c ChannelzMetrics) Describe(descriptors chan<- *prometheus.Desc) {
	descriptors <- c.ServerCallsStarted
	descriptors <- c.ServerCallsSucceeded
	descriptors <- c.ServerCallsFailed
}

func (c ChannelzMetrics) Collect(metrics chan<- prometheus.Metric) {
	resp, err := c.channelzClient.GetServers(context.Background(), &grpc_channelz_v1.GetServersRequest{})
	if err != nil {
		_, _ = fmt.Fprintf(c.logWriter, "WARNING: GetServers call to channelz client failed: %s", err)
		return
	}

	for _, server := range resp.Server {
		serverID := strconv.FormatInt(server.Ref.ServerId, 10)

		metrics <- prometheus.MustNewConstMetric(
			c.ServerCallsStarted, prometheus.CounterValue, float64(server.Data.CallsStarted), serverID)
		metrics <- prometheus.MustNewConstMetric(
			c.ServerCallsSucceeded, prometheus.CounterValue, float64(server.Data.CallsSucceeded), serverID)
		metrics <- prometheus.MustNewConstMetric(
			c.ServerCallsFailed, prometheus.CounterValue, float64(server.Data.CallsFailed), serverID)
	}
}

func newDescriptor(name, help string, constLabels prometheus.Labels, variableLabels ...string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help,
		variableLabels,
		constLabels,
	)
}
