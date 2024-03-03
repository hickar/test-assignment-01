//go:build e2e_test

package e2e_test

import (
	"context"
	"testing"
	"time"

	"github.com/hickar/crtex_test_assignment/events"
	"github.com/hickar/crtex_test_assignment/order/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestCreateOrder(t *testing.T) {
	suite.Run(t, new(e2eTestSuite))
}

type e2eTestSuite struct {
	suite.Suite
	compose tc.ComposeStack
}

func (ts *e2eTestSuite) SetupSuite() {
	compose, err := tc.NewDockerCompose("testdata/docker-compose.yaml")
	require.NoError(ts.T(), err, "NewDockerComposeAPI()")

	ts.compose = compose

	ctx, cancel := context.WithCancel(context.Background())
	ts.T().Cleanup(cancel)

	require.NoError(
		ts.T(),
		compose.Up(ctx, tc.Wait(true)),
		"compose.Up()",
	)
}

func (ts *e2eTestSuite) TearDownSuite() {
	ts.T().Cleanup(func() {
		require.NoError(
			ts.T(),
			ts.compose.Down(
				context.Background(),
				tc.RemoveOrphans(true),
				tc.RemoveImagesLocal,
			),
			"compose.Down()",
		)
	})
}

func (ts *e2eTestSuite) TestValidOrderCreation() {
	ts.T().Log("TestOrderCreation Test case")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ts.T().Log("Initializing GRPC client")

	conn, err := grpc.DialContext(
		ctx,
		"localhost:8880",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(ts.T(), err)

	client := proto.NewOrderClient(conn)

	ts.T().Log("Sending CreateOrder request")
	orderReq := &proto.CreateOrderRequest{
		UserId: 1,
		Amount: 10000,
	}
	createResp, err := client.CreateOrder(ctx, orderReq)
	require.NoError(ts.T(), err)
	assert.NotZero(ts.T(), createResp.TransactionId)

	orderID := createResp.TransactionId

	var getStatusResp *proto.GetOrderResponse
	require.Eventually(
		ts.T(),
		func() bool {
			ts.T().Log("Trying to get order by ID")
			getStatusResp, err = client.GetOrder(ctx, &proto.GetOrderRequest{TransactionId: orderID})

			return err == nil &&
				getStatusResp.Status.String() != string(events.OrderStatusCreated)
		},
		time.Second*30,
		time.Second*5,
	)

	ts.T().Log("Got order by its ID")

	require.NotNil(ts.T(), getStatusResp)
	assert.Equal(ts.T(), getStatusResp.Status, proto.Status_PAID)
	assert.Equal(ts.T(), getStatusResp.Amount, orderReq.Amount)
	assert.Equal(ts.T(), getStatusResp.ClientId, orderReq.UserId)
}

func (ts *e2eTestSuite) TestInvalidOrderCreationWithNonExistentUser() {
	ts.T().Log("TestInvalidOrderCreationWithNonExistentUser Test case")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ts.T().Log("Initializing GRPC client")

	conn, err := grpc.DialContext(
		ctx,
		"localhost:8880",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(ts.T(), err)

	client := proto.NewOrderClient(conn)

	ts.T().Log("Sending CreateOrder request")
	orderReq := &proto.CreateOrderRequest{
		UserId: 100,
		Amount: 10000,
	}
	createResp, err := client.CreateOrder(ctx, orderReq)
	require.NoError(ts.T(), err)
	assert.NotZero(ts.T(), createResp.TransactionId)

	orderID := createResp.TransactionId

	var getStatusResp *proto.GetOrderResponse
	require.Eventually(
		ts.T(),
		func() bool {
			ts.T().Log("Trying to get order by ID")
			getStatusResp, err = client.GetOrder(ctx, &proto.GetOrderRequest{TransactionId: orderID})

			return err == nil &&
				getStatusResp.Status.String() != string(events.OrderStatusCreated)
		},
		time.Second*30,
		time.Second*5,
	)

	ts.T().Log("Got order by its ID")

	require.NotNil(ts.T(), getStatusResp)
	assert.Equal(ts.T(), getStatusResp.Status, proto.Status_CANCELED)
	assert.Equal(ts.T(), getStatusResp.Amount, orderReq.Amount)
	assert.Equal(ts.T(), getStatusResp.ClientId, orderReq.UserId)
}

func (ts *e2eTestSuite) TestInvalidOrderCreationWithOverdraft() {
	ts.T().Log("TestInvalidOrderCreationWithNonExistentUser Test case")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ts.T().Log("Initializing GRPC client")

	conn, err := grpc.DialContext(
		ctx,
		"localhost:8880",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(ts.T(), err)

	client := proto.NewOrderClient(conn)

	ts.T().Log("Sending CreateOrder request")
	orderReq := &proto.CreateOrderRequest{
		UserId: 2,
		Amount: 10000000,
	}
	createResp, err := client.CreateOrder(ctx, orderReq)
	require.NoError(ts.T(), err)
	assert.NotZero(ts.T(), createResp.TransactionId)

	orderID := createResp.TransactionId

	var getStatusResp *proto.GetOrderResponse
	require.Eventually(
		ts.T(),
		func() bool {
			ts.T().Log("Trying to get order by ID")
			getStatusResp, err = client.GetOrder(ctx, &proto.GetOrderRequest{TransactionId: orderID})

			return err == nil &&
				getStatusResp.Status.String() != string(events.OrderStatusCreated)
		},
		time.Second*30,
		time.Second*5,
	)

	ts.T().Log("Got order by its ID")

	require.NotNil(ts.T(), getStatusResp)
	assert.Equal(ts.T(), getStatusResp.Status, proto.Status_CANCELED)
	assert.Equal(ts.T(), getStatusResp.Amount, orderReq.Amount)
	assert.Equal(ts.T(), getStatusResp.ClientId, orderReq.UserId)
}
