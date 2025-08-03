// Package oasis_sdk implements XMPP file upload functionality according to XEP-0363 HTTP File Upload
// specification. It provides methods for requesting upload slots and performing file uploads.
package oasis_sdk

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"mellium.im/xmpp/stanza"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadRequestDetails represents the XML structure for requesting an upload slot
// from an XMPP server. It follows the XEP-0363 specification format.
type UploadRequestDetails struct {
	XMLName     xml.Name `xml:"urn:xmpp:http:upload:0 request"`
	Filename    string   `xml:"filename,attr"`     // Name of file to be uploaded
	Size        int64    `xml:"size,attr"`         // Size of file in bytes
	ContentType *string  `xml:"content-type,attr"` // Optional MIME type of the file
}

type Header struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type PutURL struct {
	URL     string   `xml:"url,attr"`
	Headers []Header `xml:"header"`
}

type GetURL struct {
	URL string `xml:"url,attr"`
}

type UploadSlotResponsePayload struct {
	XMLName xml.Name `xml:"urn:xmpp:http:upload:0 slot"`
	Put     PutURL   `xml:"put"`
	Get     GetURL   `xml:"get"`
}

type UploadSlotResponse struct {
	stanza.IQ
	Slot UploadSlotResponsePayload `xml:"slot"`
}

// getUploadSlot requests an upload slot from the XMPP server's HTTP upload component.
// It returns the PUT URL with headers for uploading and the GET URL for retrieving the file.
// Returns an error if the upload component isn't available or if the file size exceeds limits.
func (client *XmppClient) getUploadSlot(request UploadRequestDetails) (*PutURL, string, error) {
	if client.HttpUploadComponent == nil || client.HttpUploadComponent.Jid.String() == "" {
		return nil, "", errors.New("no upload component found yet, try discovering services")
	}

	//we assume server is telling the truth
	if request.Size > client.HttpUploadComponent.MaxFileSize {
		return nil, "", fmt.Errorf(
			"upload size too large, want %d, have %d",
			request.Size, client.HttpUploadComponent.MaxFileSize,
		)
	}

	//client.Session.encode
	header := stanza.IQ{
		ID:   uuid.New().String(),
		To:   client.HttpUploadComponent.Jid,
		Type: stanza.GetIQ,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // Important to prevent context leak

	// send request for upload slot
	t, err := client.Session.EncodeIQElement(ctx, request, header)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send iq requesting upload slot, %w", err)
	}

	// decode upload slot details
	d := xml.NewTokenDecoder(t)
	response := &UploadSlotResponse{}
	err = d.Decode(response)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode upload slot response, %w", err)
	}

	return &response.Slot.Put, response.Slot.Get.URL, nil
}

// UploadFileFromBytes handles the complete process of uploading a file to the XMPP server.
// It first requests an upload slot, then performs the HTTP PUT request to upload the file.
// Returns the GET URL where the file can be downloaded from, or an error if the upload fails.
func (client *XmppClient) UploadFileFromBytes(filename string, content []byte) (string, error) {
	if filename == "" || len(content) == 0 {
		return "", errors.New("filename and content cannot be empty")
	}

	// put together data
	request := UploadRequestDetails{
		Filename: filepath.Base(filename),
		Size:     int64(len(content)),
	}

	// request upload slot
	putData, getURL, err := client.getUploadSlot(request)
	if err != nil {
		return "", fmt.Errorf("failed to get upload slot: %w", err)
	}

	//create new request object
	req, err := http.NewRequest(http.MethodPut, putData.URL, bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	//add auth headers
	for _, header := range putData.Headers {
		req.Header.Set(header.Name, header.Value)
	}

	//make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	//check if request succeeded
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed with status code: %d", resp.StatusCode)
	}

	return getURL, nil
}

// UploadFile handles the complete process of uploading a file to the XMPP server.
// It first requests an upload slot, then performs the HTTP PUT request to upload the file.
// Returns the GET URL where the file can be downloaded from, or an error if the upload fails.
func (client *XmppClient) UploadFile(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	//open file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	//get metadata
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// put together data
	request := UploadRequestDetails{
		Filename: filepath.Base(path),
		Size:     fileInfo.Size(),
	}

	// request upload slot
	putData, getURL, err := client.getUploadSlot(request)
	if err != nil {
		return "", fmt.Errorf("failed to get upload slot: %w", err)
	}

	//create new request object
	req, err := http.NewRequest(http.MethodPut, putData.URL, file)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	// explicitly set the Content-Length header
	req.ContentLength = fileInfo.Size()

	//add auth headers
	for _, header := range putData.Headers {
		req.Header.Set(header.Name, header.Value)
	}

	//make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	//check if request succeeded
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed with status code: %d", resp.StatusCode)
	}

	return getURL, nil
}
