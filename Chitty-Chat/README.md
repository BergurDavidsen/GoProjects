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
1.3 This will by default start the server on port 5001. If you want to change it, run:
```bash
$ export PORT=<YOUR DESIRED PORT>
```

#### 2. Run a client  
#### **For each Client you must open a new terminal**  
2.1 Open a new terminal and navigate to the root-folder.  
2.2 Execute:
``` bash
$ go run client/client.go
```
2.3 Enter your name in when prompted. This is the name that will be used in the chat. E.g:
```bash
$ name: <YOUR NAME>
```

2.4 You must enter the same port as the server is connected to e.g

```bash
$ localhost:5001
```
2.5 Now you can chat to all other connected clients

#### 3. Leave the chat
3.1 To leave the chat, simply terminate the program with:
```bash
ctrl+c
```


