// Package goqmk provides a Go wrapper to QMK's asynchronous API that Web and GUI tools can use to compile arbitrary keymaps for any keyboard supported by QMK.
package goqmk

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const qmkAPI = "https://api.qmk.fm/v1/keyboards"

var httpClient = &http.Client{
	Timeout: time.Second * 2,
}

// A Keyboard contains metadata as to available layouts,
// keymaps, identifier etc...
type Keyboard struct {
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Keymaps       []string `json:"keymaps"`
	GitHash       string   `json:"git_hash"`
	BootLoader    string   `json:"bootloader"`
	Platform      string   `json:"platform"`
	Processor     string   `json:"processor,omitempty"`
	ProcessorType string   `json:"processor_type"`
	DeviceVersion string   `json:"device_ver"`
	Identifier    string   `json:"identifier"`
	Maintainer    string   `json:"maintainer,omitempty"`
	Description   string   `json:"description,omitempty"`
	Layouts       []string `json:"layout"`
}

// A rawData is the super data structure for Keyboard
type rawData struct {
	LastUpdated string              `json:"last_updated"`
	Keyboards   map[string]Keyboard `json:"keyboards"`
}

// Keyboards queries QMK API and returns an array of keyboard
// names.
func AllKeyboardsList() []string {
	var keyboardsList []string
	keyboardsData := queryQMK("all")

	for k, _ := range keyboardsData {
		keyboardsList = append(keyboardsList, k)
	}

	return keyboardsList
}

// Keyboard queries QMK API and returns a Keyboard containing
// its particulars.
func KeyboardData(keyboard string) Keyboard {
	rawData := queryQMK(keyboard)
	return rawData[keyboard]
}

// Keymaps queries QMK API for a list of keymaps associated with a particular
// keyboard.
func Keymaps(kb string) []string {
	keyboardData := queryQMK(kb)
	return keyboardData[kb].Keymaps
}

func Layouts(kb string) []string {
	var layoutList []string
	keyboardData := queryQMK(kb)

	for _, v := range keyboardData[kb].Layouts {
		layoutList = append(layoutList, v)
	}
	return layoutList
}

func GetBootLoaderType(keyboard string) string {
	rawData := queryQMK(keyboard)
	return rawData[keyboard].BootLoader
}

// queryQMK is the principal internal function that queries QMK's api
func queryQMK(kb string) map[string]Keyboard {
	var rawJSON json.RawMessage
	var rawData rawData
	var escapedString string

	escapedString = (&url.URL{Path: kb}).String()

	keyboardURL := fmt.Sprintf("%s/%s", qmkAPI, escapedString)

	req, err := http.NewRequest(http.MethodGet, keyboardURL, nil)
	if err != nil {
		log.Print(err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("The HTTP request failed with error: %s", err)
	}

	rawJSON, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("The http request could not be read: %s", err)
	}

	err = json.Unmarshal(rawJSON, &rawData)
	if err != nil {
		log.Fatalf("The JSON could not be parsed: %s", err)
	}

	return rawData.Keyboards
}

func DownloadHex(keyboard string, keymap string) error {
	var depathifyString string
	depathifyString = strings.ReplaceAll(keyboard, "/", "_")
	keymap = "default"
	fileName := fmt.Sprintf("%s_%s.hex", depathifyString, keymap)
	keyboardURL := fmt.Sprintf("%s/%s", qmkAPI, fileName)

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(keyboardURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
