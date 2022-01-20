package main 

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "order_management"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"io"
	"log"
	"net"
	"strings"
)

const (
	port = ":50051"  //服务端将在50051端口监听客户端请求
	orderBatchSize = 3
)

var orderMap = make(map[string]pb.Order)

type server struct {  //server就是要实现接口getOrder的实例
	orderMap map[string]*pb.Order 
}

func (s *server)AddOrder(ctx context.Context, 
	orderReq *pb.Order) (*wrapper.StringValue, error) {
		log.Printf("Order added. ID: %v", orderReq.Id)
		orderMap[orderReq.Id] = *orderReq 
		return &wrapper.StringValue{Value: "Order Added: " + orderReq.Id}, nil 
}

func (s *server) GetOrder(ctx context.Context, 
	orderId *wrapper.StringValue) (*pb.Order, error) {
		//根据客户端发来ID查找订单数据
		ord, exists := orderMap[orderId.Value]
		if exists {
			return &ord, status.New(codes.OK, "").Err()
		}

		return nil, status.Errorf(codes.NotFound, "Order does not exist. : ", orderId)
}

func (s *server) SearchOrders(searchQuery *wrappers.StringValue, 
	stream pb.OrderManagement_SearchOrdersServer) error {
		for key, order := range orderMap {
			log.Print(key, order)
			for _, itemStr := range order.Items {
				log.Print(itemStr)
				if strings.Contains(itemStr, searchQuery.Value) {
					err := stream.Send(&order)
					if err != nil {
						return fmt.Errorf("error sending message to stream: %v", err)
					}
					log.Print("Matching Order Found: " + key)
				    break
				}
				
			}
		}
		return nil //返回nil，gRPC会关闭服务器发往客户端的数据管道
	}

func (s *server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	ordersStr := "Updated Order IDs: "
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			//通知客户端不用继续发送
			return stream.SendAndClose(&wrapper.StringValue{Value: "Orders processed" + ordersStr})
		}

		orderMap[order.Id] = *order 
	    log.Printf("Order ID ", order.Id, ": Updated")
		ordersStr += order.Id + ", "
	}
}

func (s *server) ProcessOrder(stream pb.OrderManagement_ProcessOrdersServer) error {
	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order: %s", orderId)
		if err == io.EOF {
			log.Printf("EOF: %s", orderId)
			for _, shipment := range combinedShipmentMap {
				if err := stream.Send(&shipment); err != nil {
					return err 
				}
			}
			return nil //返回nil，gRPC框架会关闭调server发送给客户端的管道
		}
		if err != nil {
			log.Println(err)
			return err 
		}
		destination := orderMap[orderId.GetValue()].Destination 
		shipment, found := combinedShipmentMap[destination]
		if found {
			ord := orderMap[orderId.GetValue()]
			shipment.OrdersList = append(shipment.OrderList, &ord)
			combinedShipmentMap[destination] = shipment 
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (orderMap[orderId.GetValue()].Destination), Status: "Processed!",}
			ord := orderMap[orderId.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, &ord)
			combinedShipmentMap[destination] = comShip 
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping: %v -> %v", comb.Id, len(comb.OrdersList))
				if err := stream.Send(&comb); err != nil {
					return err 
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}

func initSampleData() {
	//虚拟一些数据以便客户端进行请求
	orderMap["102"] = pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	orderMap["103"] = pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	orderMap["104"] = pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	orderMap["105"] = pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	orderMap["106"] = pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}

func main() {
	initSampleData()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &server{})
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}



