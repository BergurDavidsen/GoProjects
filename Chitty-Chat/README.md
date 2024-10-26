# BSDISYS1KU, Distributed Systems, BSc (Autumn 2024)
## Mandatory Activity 3 - Chitty-Chat
### Authored by Team Burger:  
> André Birk <arbi@itu.dk>  
> Bergur Davidsen <Berd@itu.dk>  
> Bror Hansen <Broh@itu.dk>  
> Konrad Meno Adolph <Koad@itu.dk>  



## How-to-use guide:   
#### 1. Run a server  
1.1 Open a terminal and navigate to the root-folder.  
1.2 Execute: 
``` bash 
$ go run server/server.go
```  
1.3 You must enter a port you want to connect to in the terminal (preferable a port not in use like 5000).  
```bash
$ localhost:5000
``` 
1.4 The server is now open.

#### 2. Run a client  
#### **For each Client you must open a new terminal**  
2.1 Open a new terminal and navigate to the root-folder.  
2.2 Execute:
``` bash
$ go run client/client.go
```
1.3 You must enter the same port as the server is connected to e.g

```bash
$ localhost:5000
```