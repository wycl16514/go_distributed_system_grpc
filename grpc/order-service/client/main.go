package main 

import (
	"context"
	"log"
	pb "order_management"
	"time"
	"google.golang.org/grpc"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"io"
)

const (
	address = "localhost:50051"  //向服务器端口发出请求
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure()) //创建与服务端的连接
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewOrderManagementClient(conn) //生成客户端对象
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()
	//调用getOrder接口向服务端提供订单ID，发起服务请求
	retriveOrder, err := client.GetOrder(ctx, &wrapper.StringValue{Value : "106"})
	log.Print("GetOrder Response -> :", retriveOrder)
    
	order1 := pb.Order{Id:"101", Items:[]string{"IPhone XS", "Mac Book Pro"},
    Destination: "San Jose, CA", Price: 2300.00}
	res, _ := client.AddOrder(ctx, &order1)
	if res != nil {
		log.Print("AddOrder response -> ", res.Value)
	}

	//stream from server
	
	searchStream, _ := client.SearchOrders(ctx, &wrapper.StringValue{Value: "Google"})
	//如果server 使用stream传输结果，客户端需要使用Recv()接收多个返回
	for {
		searchOrder, err := searchStream.Recv()
		if err == io.EOF {
			log.Print("EOF")
			break
		}
		if err == nil {
			log.Print("Search result: ", searchOrder)
		}
	}

	updOrder1 := pb.Order{Id: "102", Items:[]string{"Google Pixel 3A", "Google Pixel Book"}, Destination:"Mountain View, CA", Price:1100.00}
	updOrder2 := pb.Order{Id: "103", Items:[]string{"Apple Watch S4", "Mac Book Pro", "iPad Pro"}, Destination:"San Jose, CA", Price:2800.00}
	updOrder3 := pb.Order{Id: "104", Items:[]string{"Google Home Mini", "Google Nest Hub", "iPad Mini"}, Destination:"Mountain View, CA", Price:2200.00}

	updateStream, err := client.UpdateOrders(ctx)
	if err != nil {
		log.Fatalf("%v.UpdateOrders(_) = , %v", client, err)
	}

	if err := updateStream.Send(&updOrder1); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder1, err)
	}

	if err := updateStream.Send(&updOrder2); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder2, err)
	}

	if err := updateStream.Send(&updOrder3); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder3, err)
	}

	updateRes, err := updateStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", updateStream, err, nil)
	}
	log.Printf("Update orders res: %s", updateRes)

	streamProcOrder, err := client.ProcessOrders(ctx)
    if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = , %v", client , err)
	}
	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "102", err)
	}
	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	}
	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "104", err)
	}

	channel := make(chan struct{})
	go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Milliscond * 1000)

	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "101", err)
	}

	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	channel <- struct{}{} 
	
}

func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan struct{}) {
	for {
		combinedShipment, errorProcOrder := streamProcOrder.Recv()
		if errProcOrder == io.EOF {
			break 
		}
		log.Printf("Combined shipment: ", combinedShipment.OrdersList)
	}
	<-c
}