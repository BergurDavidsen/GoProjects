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
		currentItemIndex int
	}

	Auction struct {
		clients 		map[Service.AuctionService_AuctionServiceServer]int32
		clientBid 		[]int32
		mu 				sync.Mutex
		highestBid 		int32
		auctionItem 	string
		isActive 		bool
		startTime 		time.Time
		duration 		time.Duration
		// TODO: implement Lamport timestamps
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
	// TODO: Handle client-disconnections

	for{
		request, err := csi.Recv()
		if err != nil {
			log.Fatalf("Could not load request :: %s", err)
		}

		switch request := request.Request.(type) {
			case *Service.ClientRequest_Bid:
				// TODO: finish implementing bidding
				//csi.Send(bid(csi, as, request.Bid.Price))
			case *Service.ClientRequest_Query:
				csi.Send(as.getStatus(int(request.Query.ItemId)))

			default:
				log.Println("Unknown request type from client")
		}
	}
}

func bid(csi Service.AuctionService_AuctionServiceServer, as *AuctionServer, bid int32) *Service.ServerResponse{
	if as.currentItemIndex == 0 {
		as.newAuction()
	}

	var auction = as.auctions[as.currentItemIndex]
	auction.clients[csi] = bid
	if len(auction.clientBid) < 0 {
		auction.clientBid[0] = bid
	} else {
		auction.clientBid[len(auction.clientBid)-1] = bid
	}

	if remainingTime(auction.startTime, auction.duration) < 0 {
		return &Service.ServerResponse{
			Response: &Service.ServerResponse_Bid{
				Bid: &Service.BidResponse{
					Success: false,
					HighestBid: int32(auction.highestBid),
					CurrentItem: "Auction has ended for item: " + auction.auctionItem,
				},
			},
		}
	}

	if auction.highestBid < bid {
		auction.highestBid = bid
	}


	return nil

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

	auction := as.auctions[len(as.auctions)-1]
    remainingTime := remainingTime(auction.startTime, auction.duration)


    if remainingTime < 0 {
        remainingTime = 0
    }


	response := &Service.ServerResponse{
		Response: &Service.ServerResponse_Query{
			Query: &Service.QueryResponse{
				RemainingTime: int32(remainingTime.Seconds()),
				HighestBid: int32(auction.highestBid),
				CurrentItem: as.auctions[id].auctionItem,
			},
		},
	}

	as.mu.Unlock()
	return response
}

func remainingTime(start time.Time, duration time.Duration) time.Duration {
	return duration - time.Since(start)
}

func (as *AuctionServer) newAuction(){
	if as.currentItemIndex > len(items) {
		log.Fatal("Error :: Auction House has no more items")
	}

	as.mu.Lock()
	
	as.auctions[as.currentItemIndex] = &Auction{
		clients: make(map[Service.AuctionService_AuctionServiceServer]int32),
		auctionItem: items[as.currentItemIndex],
		isActive: true,
		startTime: time.Now(),
		duration: time.Duration(30) * time.Second,
	}

	as.currentItemIndex++

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
		currentItemIndex: 0,
	}

	Service.RegisterAuctionServiceServer(grpcServer, &auctionServer)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}

	
}