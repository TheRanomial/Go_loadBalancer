package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface{
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter,r *http.Request)
}

type simpleServer struct {
	addr string
	proxy *httputil.ReverseProxy
}

type LoadBalancer struct {
	port string
	roundRobinCount int
	servers []Server
}

func NewSimpleServer(addr string) *simpleServer{
	serverUrl,err:=url.Parse(addr)

	handleError(err)

	return &simpleServer{
		addr: addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalancer(port string,servers []Server) *LoadBalancer {

	return &LoadBalancer{
		port:port,
		roundRobinCount: 0,
		servers: servers,
	}
}

func handleError(err error){
	if err!=nil{
		fmt.Printf("Error %v\n",err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string{
	return s.addr
}

func (s *simpleServer) IsAlive() bool{
	return true
}

func (s *simpleServer) Serve(w http.ResponseWriter,r *http.Request){
	s.proxy.ServeHTTP(w,r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server{

	server:=lb.servers[lb.roundRobinCount%len(lb.servers)]

	for !server.IsAlive(){
		lb.roundRobinCount++;
		server=lb.servers[lb.roundRobinCount%len(lb.servers)]
	}

	lb.roundRobinCount++;
	return server
}

func (lb *LoadBalancer) serverProxy(w http.ResponseWriter,r *http.Request){
	targetServer:=lb.getNextAvailableServer()
	fmt.Printf("forwarding request to %q\n",targetServer.Address())
	targetServer.Serve(w,r)
}

func main(){

	servers:=[]Server{
		NewSimpleServer("https://www.facebook.com"),
		NewSimpleServer("http://www.bing.com"),
		NewSimpleServer("http://www.duckduckgo.com"),
	}

	lb:=NewLoadBalancer("8000",servers)

	handleRedirect:=func (w http.ResponseWriter,r *http.Request){
		lb.serverProxy(w,r)
	}

	http.HandleFunc("/",handleRedirect)
	fmt.Printf("Serving request at port")
	http.ListenAndServe(":"+lb.port,nil)
}