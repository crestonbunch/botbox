package sandbox

import (
	"archive/tar"
	"bufio"
	"bytes"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

const ServerDropDir = "/botbox-server"
const ClientDropDir = "/botbox-client"
const ServerUser = "sandbox"
const ClientUser = "sandbox"
const ServerImageName = "botbox-sandbox-server"
const ClientImageName = "botbox-sandbox-client"
const ClientServerEnvVar = "BOTBOX_SERVER"

// Convert a directory into a tar file to pass to the Docker image build
// Path should end with a trailing slash
func tarFile(path string) (io.Reader, error) {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(buf)
	for _, f := range dir {
		contents, err := ioutil.ReadFile(path + f.Name())
		if err != nil {
			return nil, err
		}
		tr.WriteHeader(&tar.Header{
			Name: f.Name(),
			Size: int64(f.Size()),
		})
		tr.Write(contents)
	}
	tr.Close()
	return bytes.NewReader(buf.Bytes()), nil
}

// Build a docker image from a Dockerfile with the given name
func BuildImage(cli *client.Client, path, name string) ([]byte, error) {
	file, err := tarFile(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	opts := types.ImageBuildOptions{Tags: []string{name}}
	response, err := cli.ImageBuild(context.Background(), reader, opts)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	response.Body.Close()
	return buf.Bytes(), nil
}

// Setup a sandboxed environment with docker containers for every script
// and a network bridge between them. Returns the networkId, serverId,
// a list of clientIds, or an error.
func SetupSandbox(cli *client.Client, req *MatchRequest) (string, string, []string, error) {
	// Setup network
	log.Println("Setting up network.")
	netId, err := SetupNetwork(cli)
	if err != nil {
		return "", "", nil, err
	}

	// Setup server container
	log.Println("Setting up server.")
	servId, err := SetupServer(cli, req.Server)
	if err != nil {
		return "", "", nil, err
	}

	// Connect server to the network.
	log.Println("Connecting server.")
	servEpSet := &network.EndpointSettings{}
	clientEpSet := &network.EndpointSettings{}

	err = cli.NetworkConnect(context.Background(), netId, servId, servEpSet)
	if err != nil {
		return "", "", nil, err
	}

	// Get the server IP address on the network
	netInfo, err := cli.NetworkInspect(context.Background(), netId)
	if err != nil {
		return "", "", nil, err
	}
	servIp := netInfo.Containers[servId].IPv4Address

	clients := []string{}
	for _, c := range req.Clients {
		log.Println("Setting up client.")
		clientId, err := SetupClient(cli, c, servIp)
		if err != nil {
			return "", "", nil, err
		}
		log.Println("Connecting client.")
		cli.NetworkConnect(context.Background(), netId, clientId, clientEpSet)
		clients = append(clients, clientId)
	}

	return netId, servId, clients, nil
}

// Start the server and clients in a sandbox.
func StartSandbox(cli *client.Client, server string, clients []string) error {

	opts := types.ContainerStartOptions{}
	err := cli.ContainerStart(context.Background(), server, opts)
	if err != nil {
		return err
	}
	for _, c := range clients {
		err := cli.ContainerStart(context.Background(), c, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

// Blocks until the server container stops.
func Wait(cli *client.Client, serverId string) error {
	log.Println("Waiting for container to stop.")
	_, err := cli.ContainerWait(context.Background(), serverId)
	if err != nil {
		return err
	}
	log.Println("Container stopped.")
	return nil
}

// Destroy a sandbox by passing it a list of container ids and the network id.
// It will disconnect clients from the network, remove the containers, and
// then remove the network.
func DestroySandbox(cli *client.Client, network string, containers []string) error {
	log.Println("Destroying sandbox.")
	removeOpts := types.ContainerRemoveOptions{Force: true}
	for _, c := range containers {
		err := cli.NetworkDisconnect(context.Background(), network, c, true)
		if err != nil {
			return err
		}
		err = cli.ContainerRemove(context.Background(), c, removeOpts)
		if err != nil {
			return err
		}
	}

	cli.NetworkRemove(context.Background(), network)
	return nil
}

// Setup a Docker bridge network to connect the server with the clients.
func SetupNetwork(cli *client.Client) (string, error) {
	t := time.Now().Unix()
	name := "sandbox_" + strconv.FormatInt(t, 10)
	createConfig := types.NetworkCreate{Driver: "bridge"}
	netResponse, err := cli.NetworkCreate(
		context.Background(),
		name,
		createConfig,
	)

	if err != nil {
		return "", err
	}

	return netResponse.ID, nil
}

// Setup a server sandbox in an isolated container. Returns the ID of the
// container if it was created successfully.
func SetupServer(cli *client.Client, archive Archive) (string, error) {

	// create container, but don't start it
	containerConfig := &container.Config{
		Cmd:        []string{"/bin/bash", "run.sh"},
		WorkingDir: ServerDropDir,
		User:       ServerUser,
		Image:      ServerImageName,
	}
	// TODO: send score results to scoreboard service
	hostConfig := &container.HostConfig{}
	netConfig := &network.NetworkingConfig{}
	log.Println("Creating server container.")
	response, err := cli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		netConfig,
		"",
	)

	if err != nil {
		return "", err
	}

	log.Println("Coping server files.")
	tar, err := ArchiveToTar(archive)
	if err != nil {
		return "", err
	}
	err = cli.CopyToContainer(
		context.Background(),
		response.ID,
		ServerDropDir,
		tar,
		types.CopyToContainerOptions{},
	)

	return response.ID, nil
}

// Setup a client sandbox in an isolated container. Returns the ID of the
// container if it was created successfully.
func SetupClient(cli *client.Client, archive Archive, serverIP string) (string, error) {

	// create container, but don't start it
	containerConfig := &container.Config{
		Cmd:        []string{"/bin/bash", "run.sh"},
		WorkingDir: ClientDropDir,
		User:       ClientUser,
		Image:      ClientImageName,
		Env:        []string{ClientServerEnvVar + "=" + serverIP},
	}
	hostConfig := &container.HostConfig{}
	netConfig := &network.NetworkingConfig{}
	log.Println("Creating client container.")
	response, err := cli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		netConfig,
		"",
	)

	if err != nil {
		return "", err
	}

	log.Println("Copying client files to container.")
	tar, err := ArchiveToTar(archive)
	if err != nil {
		return "", nil
	}
	err = cli.CopyToContainer(
		context.Background(),
		response.ID,
		ClientDropDir,
		tar,
		types.CopyToContainerOptions{},
	)

	return response.ID, nil
}
