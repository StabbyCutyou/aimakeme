package aimakeme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Request is an api request to be made against openapis image generate endpoint
type Request struct {
	N              int    `json:"n"`
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Quality        string `json:"quality"`
	ResponseFormat string `json:"response_format"`
	Size           string `json:"size"`
	Style          string `json:"style"`
	User           string `json:"user"`
}

// Response is an api response back from openapis image generate endpoint
type Response struct {
	Created int `json:"created"`
	Data    []struct {
		RevisedPrompt string `json:"revised_prompt"`
		URL           string `json:"url"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param"`
		Type    string `json:"type"`
	} `json:"error"`
}

func Run(opts Options) error {
	// We'll make the directory upfront before multiple threads attempt to
	if _, err := makeDirIfNotExist(*opts.Folder); err != nil {
		return err
	}
	for i := 0; i < *opts.N; i++ {
		// background the openai call, defer signal the WG when we're done
		go func(n int) {
			defer opts.Wait.Done() // last one will unblock the main thread and allow exit
			if err := Generate(opts, n); err != nil {
				fmt.Println(err) // printing IS handling it lol
			}
		}(i)
	}
	return nil
}

// Generate will take options and the number of run this is, and call the api, get the reponse,
// and store it locally
func Generate(opts Options, n int) error {
	resp, err := Post(&Request{
		N:              1,                         // some models only support some N vals, locking to 1 for now
		Model:          "dall-e-3",                // can also be dall-e-2 but that one isn't as good
		Prompt:         *opts.Prompt,              // taken from the user
		Quality:        "standard",                // 'standard', 'hd'
		ResponseFormat: "url",                     // can also be base64
		Size:           "1024x1024",               // '256x256', '512x512', '1024x1024', '1024x1792', '1792x1024'
		Style:          *opts.Style,               // 'vivid', 'natural'
		User:           strconv.Itoa(os.Getuid()), // This doesn't really help, it should likely be username+hostname
	}, opts.APIKey)
	if err != nil {
		return err
	}
	if err := Resolve(resp, opts, n); err != nil {
		return err
	}
	return nil
}

// Post actually makes the http call to open api and returns a serialized response object
func Post(r *Request, key string) (*Response, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	reply := &Response{}
	if err := json.NewDecoder(resp.Body).Decode(reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// Resolve takes a Response object and actually then downloads the image and stores the pompts
func Resolve(r *Response, opts Options, n int) error {
	// Make / Get the full path, at this point folder should already exist
	dir, err := makeDirIfNotExist(*opts.Folder)
	if err != nil {
		return err
	}
	// Call Get on the image url so we can download it
	response, err := http.Get(r.Data[0].URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	rootName := makeName(n)
	//open a file for writing
	file, err := os.Create(strings.Join([]string{dir, rootName + ".jpg"}, "/"))
	if err != nil {
		return err
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	if _, err = io.Copy(file, response.Body); err != nil {
		return err
	}

	// Dump the data into a json object that has original and revised prompts
	p, err := json.Marshal(struct {
		Prompt        string `json:"prompt"`
		RevisedPrompt string `json:"revised_prompt"`
	}{
		Prompt:        *opts.Prompt,
		RevisedPrompt: r.Data[0].RevisedPrompt,
	})
	if err != nil {
		return err
	}

	// now write the adjusted prompt
	if err := os.WriteFile(strings.Join([]string{dir, rootName + ".prompt"}, "/"), p, 0644); err != nil {
		return err
	}
	return nil
}

// makeName makes the root file name, which is just metadata about time.Now() + the number of iteration this is
func makeName(n int) string {
	t := time.Now()
	return fmt.Sprintf("%d%d%d%d%d%d_%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), n)
}

// makeDirIfNotExist makes a directory to store the images in if need be, and returns the
// full path to it either way
func makeDirIfNotExist(f string) (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fullpath := strings.Join([]string{dirname, "aimakeme", f}, "/")
	if _, err := os.Stat(fullpath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullpath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return fullpath, nil
}
