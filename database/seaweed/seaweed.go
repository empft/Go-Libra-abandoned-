package seaweed

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	seaweedURL string
	httpClient *http.Client
}

func New(seaweedURL string, httpClient *http.Client) *Client {
	myURL, err := url.Parse(seaweedURL)
	if err != nil {
		log.Panicf("invalid seaweed url: %s/n", err)
	}

	return &Client{
		seaweedURL: myURL.String(),
		httpClient: httpClient,
	}
}

type FileDetail struct {
	Fid 	 string `json:"fid"`
	FileName string `json:"fileName"`
	FileUrl  string `json:"fileUrl"`
	Size	 int64	`json:"size"`
}

func (c *Client) makePathURL(fidOrRawURL string) (*url.URL, error) {
	if !strings.HasPrefix(fidOrRawURL, c.seaweedURL) {
		myURL, err := url.Parse(c.seaweedURL)
		if err != nil {
			return nil, err
		}
		myURL.Path = fidOrRawURL
		return myURL, nil
	}
	return url.Parse(fidOrRawURL)
}

func (c *Client) Set(ctx context.Context, data io.Reader) (string, error) {
	myURL, err := url.Parse(c.seaweedURL)
	if err != nil {
		log.Panicf("invalid seaweed url: %s/n", err)
	}
	myURL.Path = "submit"

	req, err := http.NewRequestWithContext(ctx, "POST", myURL.String(), data)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", err
    }

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
        return "", err
    }

	var result FileDetail
	err = json.Unmarshal(content, &result)
	if err != nil {
        return "", err
    }

	return result.Fid, nil	
}

func (c *Client) Get(ctx context.Context, fid string) ([]byte, error) {
	myURL, err := c.makePathURL(fid)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", myURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
	defer resp.Body.Close()

	if code := resp.StatusCode; !(code >= 200 && code < 300) {
		return nil, ErrFailedResponseStatus
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
        return nil, err
    }
	return buf.Bytes(), nil
}

func (c *Client) Delete(ctx context.Context, fids ...string) error {
	for _, fid := range fids {
		myURL, err := c.makePathURL(fid)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, "DELETE", myURL.String(), nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; !(code >= 200 && code < 300) {
			return ErrFailedResponseStatus
		}
	}
	return nil
}

// Get url from fid
func (c *Client) FormatURL(fid ...string) ([]string, error) {
	urlList := make([]string, 0)
	for _, v := range fid {
		myURL, err := c.makePathURL(v)
		if err != nil {
			return nil, err
		}
		urlList = append(urlList, myURL.String())
	}
	return urlList, nil
}

// Get fid from url
func (c *Client) FormatFID(objectURL ...string) ([]string, error) {
	fidList := make([]string, 0)
	for _, v := range objectURL {
		myURL, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		fidList = append(fidList, myURL.Path)
	}
	return fidList, nil
}