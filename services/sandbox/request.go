package sandbox

import (
	"errors"
	"io"
	"net/http"
)

const SetupMemUsage = 100000 // bytes

// A request to start a match with readers to the directories for starting the
// match. A list of client unique Ids must be passed in the same order as a
// list of zip files that contain the agent scripts for each id. Each request
// must be a multipart-form encoded request.
type MatchRequest struct {
	// Passed in the 'server' property
	Server Archive
	// Passed in the 'ids' property
	Ids []string
	// Passed in the 'clients' property
	Clients []Archive
}

// Close open readers in the match request.
func (m *MatchRequest) Close() {
	m.Server.Close()
	for _, c := range m.Clients {
		c.Close()
	}
}

// Build the request from an HTTP multipart/form POST request. The request must
// contain a single server .zip file and a list of client .zip files.
// Remember to Close() the MatchRequest when you're done with it!
func FromHttp(r *http.Request) (*MatchRequest, error) {
	if r.Method != http.MethodPost {
		return nil, errors.New("Method not allowed")
	}
	err := r.ParseMultipartForm(SetupMemUsage)
	if err != nil && err != io.EOF {
		return nil, err
	}
	m := r.MultipartForm

	serverFiles := m.File["server"]
	clientFiles := m.File["clients"]

	if len(serverFiles) == 0 {
		return nil, errors.New("Missing server file.")
	}
	if len(serverFiles) > 1 {
		return nil, errors.New("Too many server files.")
	}
	if len(clientFiles) == 0 {
		return nil, errors.New("Missing client files.")
	}

	// open the server reader
	server, err := serverFiles[0].Open()
	if err != nil && err != io.EOF {
		return nil, err
	}

	serverArchive, err := OpenArchive(server)
	if err != nil {
		return nil, err
	}

	clientIds := m.Value["ids"]

	clientArchives := []Archive{}
	// open the client readers
	for _, f := range clientFiles {
		c, err := f.Open()
		if err != nil && err != io.EOF {
			return nil, err
		}
		archive, err := OpenArchive(c)
		if err != nil {
			return nil, err
		}
		clientArchives = append(clientArchives, archive)
	}

	return &MatchRequest{serverArchive, clientIds, clientArchives}, nil
}
