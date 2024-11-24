# Mandatory Hand-in 5
> Authors:  
> André Birk \<arbi@itu.dk\>  
> Bergur Davidsen \<berd@itu.dk\>  
> Bror Hansen \<broh@itu.dk\>  
> Konrad Adolph \<koad@itu.dk\>  

## Guide
### Setup server
1. Ensure you are in the Replication root folder.
2. Type `$ go run server/main.go <port>`.
3. Repeat this for 3 times.

### Client server
1. Ensure you are in the Replication root folder.
2. For each port you've opened enter them in the command below.  
    Type `$ go run client/main.go <name> <port1> <port2> <port3> ...`  
    if you have setup more than three server nodes up, you can feel free to add them to the end in the same format.

    