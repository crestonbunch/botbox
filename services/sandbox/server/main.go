package main

import (
	"botbox/services/sandbox"
	"github.com/docker/engine-api/client"
	"log"
	"net/http"
)

// TODO: load the docker client settings from ENV variables
const DockerUserAgent = "botbox-game-1.0"
const DockerSocketPath = "unix:///var/run/docker.sock"
const DockerAPIVersion = "v1.23"
const ServerDockerFile = "server-image/"
const ClientDockerFile = "client-image/"

// Request to start a match. To create a listener, provide a cli interface to
// a Docker engine, and the HTTP response writer and reader. To start a match
// send a multipart/form request to the endpoint which contains a "server"
// entry which is a .zip file for the server and a "clients" entry which is
// a list of .zip files for each client.
// TODO: make this a transaction-like approach where if one part of the
// sandbox fails to start, we clean up what we made so there aren't a bunch of
// unused docker networks and containers floating around the host
func matchStarter(cli *client.Client, w http.ResponseWriter, r *http.Request) {
	log.Println("Received request!")
	// parse the request
	request, err := sandbox.FromHttp(r)
	if err != nil {
		log.Println("Error parsing request")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	defer request.Close()

	// setup the Sandbox
	netId, serverId, clientIds, err := sandbox.SetupSandbox(cli, request)
	if err != nil {
		log.Println("Error setting up sandbox.")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	// start the server
	err = sandbox.StartSandbox(cli, serverId, clientIds)
	if err != nil {
		log.Println("Error starting sandbox.")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	// Wait for the server to close, then destroy the sandbox
	err = sandbox.Wait(cli, serverId)
	if err != nil {
		log.Println("Error waiting for sandbox to close.")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	// Log container logs
	err = sandbox.LogSandbox(cli, serverId, clientIds)
	if err != nil {
		log.Println("Error logging sandbox.")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	err = sandbox.DestroySandbox(cli, netId, append(clientIds, serverId))
	if err != nil {
		log.Println("Error destroying sandbox.")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
}

func main() {
	defaultHeaders := map[string]string{"User-Agent": DockerUserAgent}
	cli, err := client.NewClient(DockerSocketPath, DockerAPIVersion, nil, defaultHeaders)
	if err != nil {
		panic(err)
	}
	// build the docker server images when the server starts up
	log.Println("Building botbox-server image")
	r, err := sandbox.BuildImage(cli, ServerDockerFile, sandbox.ServerImageName)
	if err != nil {
		panic(err)
	}
	log.Printf(string(r))
	// build the docker client images when the server starts up
	log.Println("Building botbox-client image")
	r, err = sandbox.BuildImage(cli, ClientDockerFile, sandbox.ClientImageName)
	if err != nil {
		panic(err)
	}
	log.Printf(string(r))

	log.Println("Waiting for connections")

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		matchStarter(cli, w, r)
	})
	http.ListenAndServe(":8080", nil)
}
