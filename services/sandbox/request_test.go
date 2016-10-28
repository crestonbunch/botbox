package sandbox

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"
)

func multipartAddFile(
	writer *multipart.Writer, fieldname, filename string, contents []byte,
) error {
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return err
	}
	part.Write(contents)

	return nil
}

func multipartAddField(
	writer *multipart.Writer, fieldname, contents string,
) error {
	part, err := writer.CreateFormField(fieldname)
	if err != nil {
		return err
	}
	part.Write([]byte(contents))
	return nil
}

func TestGoodRequest(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	serverZip := new(bytes.Buffer)
	serverWriter := zip.NewWriter(serverZip)
	w, err := serverWriter.Create("main.go")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(tron_server))
	if err != nil {
		t.Error(err)
	}
	serverWriter.Close()

	agentZip := new(bytes.Buffer)
	agentWriter := zip.NewWriter(agentZip)
	w, err = agentWriter.Create("__init__.py")
	if err != nil {
		t.Error(err)
	}
	_, err = w.Write([]byte(random_agent))
	if err != nil {
		t.Error(err)
	}
	agentWriter.Close()
	agentBytes := agentZip.Bytes()

	err = multipartAddFile(writer, "server", "tron-server.zip", serverZip.Bytes())
	if err != nil {
		t.Error(err)
	}
	err = multipartAddFile(writer, "clients", "random_agent.zip", agentBytes)
	if err != nil {
		t.Error(err)
	}
	err = multipartAddFile(writer, "clients", "random_agent.zip", agentBytes)
	if err != nil {
		t.Error(err)
	}
	err = multipartAddField(writer, "ids", "id1")
	if err != nil {
		t.Error(err)
	}
	err = multipartAddField(writer, "ids", "id2")
	if err != nil {
		t.Error(err)
	}
	if err := writer.Close(); err != nil {
		t.Error(err)
	}

	mockReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost/",
		body,
	)
	if err != nil {
		t.Error(err)
	}
	mockReq.Header.Set("Content-Type", writer.FormDataContentType())

	req, err := FromHttp(mockReq)
	if err != nil {
		t.Error(err)
	}

	if req.Server == nil {
		t.Error("Request archive is nil!")
	}

	if len(req.Ids) != 2 {
		t.Error("Request does not have 2 ids.")
	}

	if len(req.Clients) != 2 {
		t.Error("Request does not have 2 clients.")
	}
}
