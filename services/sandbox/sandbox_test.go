package sandbox

import (
	"archive/zip"
	"bytes"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"log"
	"testing"
)

const random_agent = `import botbox_tron
import random
def move(p, actions, state):
	safe = botbox_tron.safe_moves(p, state)
	actions = [a for a in actions if a in safe]
	if actions:
		return random.choice(actions)
	else:
		return
botbox_tron.start(move)`

const write_agent = `import botbox_tron
import random
def move(p, actions, state):
	safe = botbox_tron.safe_moves(p, state)
	actions = [a for a in actions if a in safe]
	with open('test.b0tn3t', 'w') as fh:
			fh.write("You've been hacked!")
	except Exception as e:
			print(e)
	if actions:
		return random.choice(actions)
	else:
		return
botbox_tron.start(move)`

const root_agent = `import botbox_tron
import random
import subprocess
def move(p, actions, state):
	safe = botbox_tron.safe_moves(p, state)
	actions = [a for a in actions if a in safe]
	try:
		subprocess.Popen(['/usr/bin/pkexec','touch','/malware.r00t'])
	except Exception as e:
			print(e)
	if actions:
		return random.choice(actions)
	else:
		return
botbox_tron.start(move)`

const tron_server = `package main
import (
	"github.com/crestonbunch/botbox/common/game"
	"github.com/crestonbunch/botbox/games/tron"
	"golang.org/x/net/websocket"
)
func main() {
	exitChan := make(chan bool)
	go func() {
		game.RunAuthenticatedServer(
			func(idList, secretList []string) (websocket.Handler, error) {
				writer, err := game.NewSimpleGameRecorder("./")
				if err != nil {
					return nil, err
				}
				stateMan := game.NewSynchronizedStateManager(
					tron.NewTwoPlayerTron(32, 32), game.MoveTimeout,
				)
				return game.GameHandler(
					exitChan,
					game.NewSimpleConnectionManager(),
					game.NewAuthenticatedClientManager(
						stateMan.NewClient, idList, secretList, game.ConnTimeout,
					),
					stateMan,
					writer,
				), nil
			},
		)
	}()
	<-exitChan
}`

const DockerUserAgent = "botbox-game-1.0"
const DockerSocketPath = "unix:///var/run/docker.sock"
const DockerAPIVersion = "v1.24"
const ServerDockerFile = "server-image/"
const ClientDockerFile = "client-image/"

var defaultHeaders = map[string]string{"User-Agent": DockerUserAgent}
var cli, _ = client.NewClient(DockerSocketPath, DockerAPIVersion, nil, defaultHeaders)

func TestSetupNetwork(t *testing.T) {
	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}

	res, err := cli.NetworkInspect(context.Background(), netId)

	if res.Driver != "bridge" {
		t.Error("Network driver is not bridge!")
	}
	if res.Internal != true {
		t.Error("Network is not an internal network!")
	}

	cli.NetworkRemove(context.Background(), netId)
}

func TestGenerateSecrets(t *testing.T) {
	// Generate client secrets so they can't connect more than once.
	secrets, err := GenerateSecrets(2)
	if err != nil {
		t.Error(err)
	}
	if len(secrets) != 2 {
		t.Error("Did not generate 2 secrets!")
	}

	if secrets[0] == secrets[1] {
		t.Error("Did not generate unique secrets!")
	}
}

func TestSetupZipServer(t *testing.T) {

	// build server zip
	serverBuf := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverBuf)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()
	serverReader := bytes.NewReader(serverBuf.Bytes())
	serverZip, err := zip.NewReader(serverReader, int64(serverBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	serverArchive := &ZipArchive{serverZip}

	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkRemove(context.Background(), netId)

	removeOpts := types.ContainerRemoveOptions{Force: true}
	// create the server
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	servId, err := SetupServer(cli, ids, secrets, serverArchive)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkDisconnect(context.Background(), netId, servId, true)
	defer cli.ContainerRemove(context.Background(), servId, removeOpts)

	// start the server
	_, err = StartServer(cli, netId, servId)
	if err != nil {
		t.Error(err)
	}

	res, err := cli.NetworkInspect(context.Background(), netId)
	if err != nil {
		t.Error(err)
	}
	if _, ok := res.Containers[servId]; !ok {
		t.Error("Server was not connected to network!")
	}

	diff, err := cli.ContainerDiff(context.Background(), servId)
	if err != nil {
		t.Error(err)
	}
	valid := false
	for _, d := range diff {
		if d.Path == "/botbox-server/main.go" {
			valid = true
		}
	}
	if !valid {
		t.Error("Server did not get main.go!")
	}

}

func TestSetupZipClients(t *testing.T) {

	// build server zip
	serverBuf := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverBuf)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()
	serverReader := bytes.NewReader(serverBuf.Bytes())
	serverZip, err := zip.NewReader(serverReader, int64(serverBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	serverArchive := &ZipArchive{serverZip}

	// build client zips
	clientBuf := new(bytes.Buffer)
	clientWriter := zip.NewWriter(clientBuf)
	w, err = clientWriter.Create("__init__.py")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(random_agent))
	if err != nil {
		t.Error(err)
	}
	clientWriter.Close()
	clientReader := bytes.NewReader(clientBuf.Bytes())
	clientZip, err := zip.NewReader(clientReader, int64(clientBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	clientArchive := &ZipArchive{clientZip}

	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkRemove(context.Background(), netId)

	removeOpts := types.ContainerRemoveOptions{Force: true}
	// create the server
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	servId, err := SetupServer(cli, ids, secrets, serverArchive)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkDisconnect(context.Background(), netId, servId, true)
	defer cli.ContainerRemove(context.Background(), servId, removeOpts)
	// start the server
	servIp, err := StartServer(cli, netId, servId)
	if err != nil {
		t.Error(err)
	}

	// create the clients
	clientIds, err := SetupClients(
		cli, netId, servIp, secrets, []Archive{clientArchive, clientArchive},
	)
	if err != nil {
		t.Error(err)
	}
	for _, id := range clientIds {
		defer cli.NetworkDisconnect(context.Background(), netId, id, true)
		defer cli.ContainerRemove(context.Background(), id, removeOpts)
	}
	// start the clients
	err = StartClients(cli, netId, clientIds)
	if err != nil {
		t.Error(err)
	}
	res, err := cli.NetworkInspect(context.Background(), netId)
	if err != nil {
		t.Error(err)
	}
	for _, id := range clientIds {
		if _, ok := res.Containers[id]; !ok {
			t.Error("Client did not connect to network!")
		}

		diff, err := cli.ContainerDiff(context.Background(), id)
		if err != nil {
			t.Error(err)
		}
		valid := false
		for _, d := range diff {
			if d.Path == "/botbox-client/__init__.py" {
				valid = true
			}
		}
		if !valid {
			t.Error("Client did not get __init__.py!")
		}
	}
}

func TestRunClients(t *testing.T) {

	// build server zip
	serverBuf := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverBuf)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()
	serverReader := bytes.NewReader(serverBuf.Bytes())
	serverZip, err := zip.NewReader(serverReader, int64(serverBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	serverArchive := &ZipArchive{serverZip}

	// build client zips
	clientBuf := new(bytes.Buffer)
	clientWriter := zip.NewWriter(clientBuf)
	w, err = clientWriter.Create("__init__.py")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(random_agent))
	if err != nil {
		t.Error(err)
	}
	clientWriter.Close()
	clientReader := bytes.NewReader(clientBuf.Bytes())
	clientZip, err := zip.NewReader(clientReader, int64(clientBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	clientArchive := &ZipArchive{clientZip}

	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkRemove(context.Background(), netId)

	removeOpts := types.ContainerRemoveOptions{Force: true}
	// create the server
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	servId, err := SetupServer(cli, ids, secrets, serverArchive)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkDisconnect(context.Background(), netId, servId, true)
	defer cli.ContainerRemove(context.Background(), servId, removeOpts)
	// start the server
	servIp, err := StartServer(cli, netId, servId)
	if err != nil {
		t.Error(err)
	}

	// create the clients
	clientIds, err := SetupClients(
		cli, netId, servIp, secrets, []Archive{clientArchive, clientArchive},
	)
	if err != nil {
		t.Error(err)
	}
	for _, id := range clientIds {
		defer cli.NetworkDisconnect(context.Background(), netId, id, true)
		defer cli.ContainerRemove(context.Background(), id, removeOpts)
	}
	// start the clients
	err = StartClients(cli, netId, clientIds)
	if err != nil {
		t.Error(err)
	}

	// Wait for the server to close
	err = Wait(cli, servId)
	if err != nil {
		t.Error(err)
	}

	conns, err := ClientsConnected(cli, servId)
	if err != nil {
		t.Error(err)
	}
	if !((conns[0] == ids[0] || conns[1] == ids[0]) &&
		(conns[0] == ids[1] || conns[1] == ids[1]) &&
		(conns[0] != conns[1])) {
		t.Error("Connected client ids are wrong!")
	}

	bad, err := BadClients(cli, servId)
	if err != nil {
		t.Error(err)
	}
	if len(bad) > 0 {
		t.Error("No clients should have committed a sin!")
	}

	results, err := GameResult(cli, servId)
	if err != nil {
		t.Error(err)
	}
	if !(((results[0] == 1 || results[1] == 1) && results[0] != results[1]) ||
		(results[0] == 0 && results[0] == 0)) {
		t.Error("Game results are not valid.")
	}

	states, err := GameHistory(cli, servId)
	if err != nil {
		t.Error(err)
	}
	if len(states) == 0 {
		t.Error("No game states recorded!")
	}

	logs, err := ContainerLogs(cli, servId)
	if err != nil {
		t.Error(err)
	}
	if len(logs) == 0 {
		t.Error("No server Stdout!")
	}

}

func TestWriteRoot(t *testing.T) {

	// build server zip
	serverBuf := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverBuf)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()
	serverReader := bytes.NewReader(serverBuf.Bytes())
	serverZip, err := zip.NewReader(serverReader, int64(serverBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	serverArchive := &ZipArchive{serverZip}

	// build client zips
	clientArchives := []*ZipArchive{}
	for _, z := range []string{write_agent, root_agent} {
		clientBuf := new(bytes.Buffer)
		clientWriter := zip.NewWriter(clientBuf)
		w, err = clientWriter.Create("__init__.py")
		if err != nil {
			t.Error(err)
		}
		_, err = w.Write([]byte(z))
		if err != nil {
			t.Error(err)
		}
		clientWriter.Close()
		clientReader := bytes.NewReader(clientBuf.Bytes())
		clientZip, err := zip.NewReader(clientReader, int64(clientBuf.Len()))
		if err != nil {
			t.Error(err)
		}
		clientArchive := &ZipArchive{clientZip}
		clientArchives = append(clientArchives, clientArchive)
	}

	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkRemove(context.Background(), netId)

	removeOpts := types.ContainerRemoveOptions{Force: true}
	// create the server
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	servId, err := SetupServer(cli, ids, secrets, serverArchive)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkDisconnect(context.Background(), netId, servId, true)
	defer cli.ContainerRemove(context.Background(), servId, removeOpts)
	// start the server
	servIp, err := StartServer(cli, netId, servId)
	if err != nil {
		t.Error(err)
	}

	// create the clients
	clientIds, err := SetupClients(
		cli, netId, servIp, secrets, []Archive{clientArchives[0], clientArchives[1]},
	)
	if err != nil {
		t.Error(err)
	}
	for _, id := range clientIds {
		defer cli.NetworkDisconnect(context.Background(), netId, id, true)
		defer cli.ContainerRemove(context.Background(), id, removeOpts)
	}
	// start the clients
	err = StartClients(cli, netId, clientIds)
	if err != nil {
		t.Error(err)
	}

	// Wait for the server to close
	err = Wait(cli, servId)
	if err != nil {
		t.Error(err)
	}

	diff, err := cli.ContainerDiff(context.Background(), clientIds[0])
	if err != nil {
		t.Error(err)
	}
	valid := false
	for _, d := range diff {
		if d.Path == "/malware.r00t" || d.Path == "/botbox-client/test.b0tn3t" {
			valid = true
		}
	}
	if valid {
		t.Error("Client was infected!")
	}

	diff, err = cli.ContainerDiff(context.Background(), clientIds[1])
	if err != nil {
		t.Error(err)
	}
	valid = false
	for _, d := range diff {
		if d.Path == "/malware.r00t" || d.Path == "/botbox-client/test.b0tn3t" {
			valid = true
		}
	}
	if valid {
		t.Error("Client was infected!")
	}

}

func TestPingServer(t *testing.T) {

	// build server zip
	serverBuf := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverBuf)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()
	serverReader := bytes.NewReader(serverBuf.Bytes())
	serverZip, err := zip.NewReader(serverReader, int64(serverBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	serverArchive := &ZipArchive{serverZip}

	// build client zips
	clientBuf := new(bytes.Buffer)
	clientWriter := zip.NewWriter(clientBuf)
	w, err = clientWriter.Create("__init__.py")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(random_agent))
	if err != nil {
		t.Error(err)
	}
	clientWriter.Close()
	clientReader := bytes.NewReader(clientBuf.Bytes())
	clientZip, err := zip.NewReader(clientReader, int64(clientBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	clientArchive := &ZipArchive{clientZip}

	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkRemove(context.Background(), netId)

	removeOpts := types.ContainerRemoveOptions{Force: true}
	// create the server
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	servId, err := SetupServer(cli, ids, secrets, serverArchive)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkDisconnect(context.Background(), netId, servId, true)
	defer cli.ContainerRemove(context.Background(), servId, removeOpts)
	// start the server
	servIp, err := StartServer(cli, netId, servId)
	if err != nil {
		t.Error(err)
	}

	// create the clients
	clientIds, err := SetupClients(
		cli, netId, servIp, secrets, []Archive{clientArchive, clientArchive},
	)
	if err != nil {
		t.Error(err)
	}
	for _, id := range clientIds {
		defer cli.NetworkDisconnect(context.Background(), netId, id, true)
		defer cli.ContainerRemove(context.Background(), id, removeOpts)
	}
	// start the clients
	err = StartClients(cli, netId, clientIds)
	if err != nil {
		t.Error(err)
	}

	resp, err := cli.ContainerExecCreate(
		context.Background(),
		clientIds[0],
		types.ExecConfig{
			AttachStderr: true,
			AttachStdout: true,
			Tty:          true,
			Cmd:          []string{"ping", "google.com"},
		},
	)

	err = cli.ContainerExecStart(
		context.Background(),
		resp.ID,
		types.ExecStartCheck{},
	)
	if err != nil {
		t.Error(err)
	}

	insp, err := cli.ContainerExecInspect(
		context.Background(),
		resp.ID,
	)
	if err != nil {
		t.Error(err)
	}

	if insp.ExitCode == 0 {
		t.Error("There was no error pinging Google!")
	}
}

func TestDestroySandbox(t *testing.T) {

	// build server zip
	serverBuf := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverBuf)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()
	serverReader := bytes.NewReader(serverBuf.Bytes())
	serverZip, err := zip.NewReader(serverReader, int64(serverBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	serverArchive := &ZipArchive{serverZip}

	// build client zips
	clientBuf := new(bytes.Buffer)
	clientWriter := zip.NewWriter(clientBuf)
	w, err = clientWriter.Create("__init__.py")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(random_agent))
	if err != nil {
		t.Error(err)
	}
	clientWriter.Close()
	clientReader := bytes.NewReader(clientBuf.Bytes())
	clientZip, err := zip.NewReader(clientReader, int64(clientBuf.Len()))
	if err != nil {
		t.Error(err)
	}
	clientArchive := &ZipArchive{clientZip}

	// create the network
	netId, err := SetupNetwork(cli)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkRemove(context.Background(), netId)

	removeOpts := types.ContainerRemoveOptions{Force: true}
	// create the server
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	servId, err := SetupServer(cli, ids, secrets, serverArchive)
	if err != nil {
		t.Error(err)
	}
	defer cli.NetworkDisconnect(context.Background(), netId, servId, true)
	defer cli.ContainerRemove(context.Background(), servId, removeOpts)
	// start the server
	servIp, err := StartServer(cli, netId, servId)
	if err != nil {
		t.Error(err)
	}

	// create the clients
	clientIds, err := SetupClients(
		cli, netId, servIp, secrets, []Archive{clientArchive, clientArchive},
	)
	if err != nil {
		t.Error(err)
	}
	for _, id := range clientIds {
		defer cli.NetworkDisconnect(context.Background(), netId, id, true)
		defer cli.ContainerRemove(context.Background(), id, removeOpts)
	}
	// start the clients
	err = StartClients(cli, netId, clientIds)
	if err != nil {
		t.Error(err)
	}

	// Destroy the sandbox
	err = DestroySandbox(cli, netId, append(clientIds, servId))
	if err != nil {
		t.Error(err)
	}

	_, err = cli.ContainerInspect(context.Background(), servId)
	if err == nil {
		t.Error("There was no error inspecting the server!")
	}
	_, err = cli.ContainerInspect(context.Background(), clientIds[0])
	if err == nil {
		t.Error("There was no error inspecting client 1!")
	}
	_, err = cli.ContainerInspect(context.Background(), clientIds[1])
	if err == nil {
		t.Error("There was no error inspecting client 2!")
	}
	_, err = cli.NetworkInspect(context.Background(), netId)
	if err == nil {
		t.Error("There was no error inspecting the network!")
	}
}

func TestBuildServer(t *testing.T) {
	_, err := BuildImage(cli, "server/server-image/", "test-image-please-ignore")
	log.Println(err)
	if err != nil {
		t.Error("There was an error building the image")
	}
	cli.ImageRemove(
		context.Background(),
		"test-image-please-ignore",
		types.ImageRemoveOptions{},
	)
}

func TestBuildClient(t *testing.T) {
	_, err := BuildImage(cli, "server/client-image/", "tst-image-please-ignore")
	log.Println(err)
	if err != nil {
		t.Error("There was an error building the image")
	}
	cli.ImageRemove(
		context.Background(),
		"tst-image-please-ignore",
		types.ImageRemoveOptions{},
	)
}
