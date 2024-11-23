package main

import (
	"AuctionSystem/Service"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	//"google.golang.org/grpc/metadata"
)

var ports = []string{"5001", "5002", "5003", "5004", "5005", "5006", "5007", "5008", "5009", "5010"}
var items = []string{
	"Luxury Safari Getaway Package", 
	"Collection of Vintage Comic Books", 
	"1967 Ford Mustang Fastback", 
	"Ming Dynasty Vase",
	"Custom Gibson Les Paul Guitar",
	"Rare First Edition of The Great Gatsby",
	"19th Century Oil Painting",
	"Antique Persian Rug",
	"Signed Michael Jordan Basketball",
	"Vintage Rolex Submariner Watch",
}

type(
	AuctionServer struct {
		Service.UnimplementedAuctionServiceServer
		clients 		map[Service.AuctionService_AuctionServiceServer]bool
		mu 				sync.Mutex
		auctions 		[]*Auction
	}

	Auction struct {
		clients 		map[Service.AuctionService_AuctionServiceServer]uint32
		mu 				sync.Mutex
		highestBid 		uint32
		auctionItem 	string
		isActive 		bool
		startTime 		time.Time
		duration 		time.Duration
	}
)

func (as *AuctionServer) AuctionService(csi Service.AuctionService_AuctionServiceServer) error {
	errch := make(chan error)
	as.mu.Lock()

	as.clients[csi] = true
	log.Println("User has joined the Auction House")

	go receiveFromStream(csi, as)

	as.mu.Unlock()
	return <-errch
}

func receiveFromStream(csi Service.AuctionService_AuctionServiceServer, as *AuctionServer) {
	for{
		request, err := csi.Recv()
		if err != nil {
			log.Fatalf("Could not load request :: %s", err)
		}

		switch request := request.Request.(type) {
			case *Service.ClientRequest_Bid:
				log.Println("a")				
			case *Service.ClientRequest_Query:
				csi.Send(as.getStatus(int(request.Query.ItemId)))

			default:
				log.Println("Unknown request type from client")
		}
	}
}

func bid(csi Service.AuctionService_AuctionServiceServer, as *AuctionServer){
	

}

func (as *AuctionServer) getStatus(id int) *Service.ServerResponse {
	as.mu.Lock()
	if len(as.auctions) == 0 {
		response := &Service.ServerResponse{
			Response: &Service.ServerResponse_Query{
				Query: &Service.QueryResponse{
					RemainingTime: int32(30),
					HighestBid: int32(0),
					CurrentItem: items[0],
				},
			},
		}
		
		as.mu.Unlock()
		return response
	} else if id == -1 {
		id = len(as.auctions) -1
	}


	response := &Service.ServerResponse{
		Response: &Service.ServerResponse_Query{
			Query: &Service.QueryResponse{
				RemainingTime: int32(time.Since(as.auctions[id].startTime)),
				HighestBid: int32(as.auctions[id].highestBid),
				CurrentItem: as.auctions[id].auctionItem,
			},
		},
	}

	as.mu.Unlock()
	return response
}

func (as *AuctionServer) newAuction(id int){
	if !(len(items) > 0) {
		log.Fatal("Error :: Auction House has no more items")
	}

	as.mu.Lock()
	
	as.auctions[0] = &Auction{
		clients: make(map[Service.AuctionService_AuctionServiceServer]uint32),
		auctionItem: items[0],
		isActive: true,
		startTime: time.Now(),
		duration: time.Duration(30) * time.Second,
	}



	as.mu.Unlock()
}

func listener() (net.Listener, string, error) {
	for _, port := range ports {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			log.Printf("Could not listen @ %v :: %v", port, err)
			continue
		}
		return listener, port, nil
	}
	return nil, "err",  fmt.Errorf("no available ports in the list")
}

func main() {
	listener, port, err := listener()
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", port, err)
	}
	log.Println("Server listening @ :" + port)

	grpcServer := grpc.NewServer()
	auctionServer := AuctionServer{
		clients: make(map[Service.AuctionService_AuctionServiceServer]bool),
		auctions: []*Auction{},
	}

	Service.RegisterAuctionServiceServer(grpcServer, &auctionServer)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}

	
}