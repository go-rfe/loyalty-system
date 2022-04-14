package server_test

import (
	"context"
	"testing"

	accrualMocks "github.com/go-rfe/loyalty-system/internal/accrual/mocks"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
	ordersMocks "github.com/go-rfe/loyalty-system/internal/repository/orders/mocks"
	"github.com/go-rfe/loyalty-system/internal/server"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

type testOrders struct {
	order        orders.Order
	accrualOrder *orders.Order
}

type testPoller struct {
	name       string
	buildStubs func(client *accrualMocks.MockClient, store *ordersMocks.MockStore)
}

func TestUpdateOrders(t *testing.T) {
	ordersForTests := []testOrders{
		{
			order: orders.Order{
				Number: "9278923470",
				Status: "NEW",
			},
			accrualOrder: &orders.Order{
				Number:  "9278923470",
				Status:  "PROCESSED",
				Accrual: decimal.NewFromInt(500),
			},
		},
		{
			order: orders.Order{
				Number: "346436439",
				Status: "NEW",
			},
			accrualOrder: &orders.Order{
				Number: "346436439",
				Status: "REGISTERED",
			},
		},
		{
			order: orders.Order{
				Number: "12345678903",
				Status: "NEW",
			},
			accrualOrder: &orders.Order{
				Number: "12345678903",
				Status: "INVALID",
			},
		},
	}
	tests := []testPoller{
		{
			name: "Accrued order",
			buildStubs: func(client *accrualMocks.MockClient, store *ordersMocks.MockStore) {
				store.EXPECT().GetUnprocessedOrders(gomock.Any()).Return([]orders.Order{ordersForTests[0].order}, nil)
				client.EXPECT().GetOrder(gomock.Any(),
					ordersForTests[0].order.Number).Return(ordersForTests[0].accrualOrder, nil).Times(1)
				store.EXPECT().UpdateOrder(gomock.Any(), ordersForTests[0].accrualOrder).Return(nil).Times(1)
			},
		},
		{
			name: "Skipped order",
			buildStubs: func(client *accrualMocks.MockClient, store *ordersMocks.MockStore) {
				store.EXPECT().GetUnprocessedOrders(gomock.Any()).Return([]orders.Order{ordersForTests[1].order}, nil)
				client.EXPECT().GetOrder(gomock.Any(),
					ordersForTests[1].order.Number).Return(ordersForTests[1].accrualOrder, nil).Times(1)
				store.EXPECT().UpdateOrder(gomock.Any(), ordersForTests[1].accrualOrder).Return(nil).Times(0)
			},
		},
		{
			name: "Invalid order",
			buildStubs: func(client *accrualMocks.MockClient, store *ordersMocks.MockStore) {
				store.EXPECT().GetUnprocessedOrders(gomock.Any()).Return([]orders.Order{ordersForTests[2].order}, nil)
				client.EXPECT().GetOrder(gomock.Any(),
					ordersForTests[2].order.Number).Return(ordersForTests[2].accrualOrder, nil).Times(1)
				store.EXPECT().UpdateOrder(gomock.Any(), ordersForTests[2].accrualOrder).Return(nil).Times(1)
			},
		},
	}

	client, store := getMocks(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.buildStubs(client, store)
			server.UpdateOrders(context.Background(), client, store)
		})
	}
}

func getMocks(t *testing.T) (*accrualMocks.MockClient, *ordersMocks.MockStore) {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	a := accrualMocks.NewMockClient(ctrl)
	s := ordersMocks.NewMockStore(ctrl)

	return a, s
}
