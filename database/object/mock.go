package object

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/patrickmn/go-cache"
	"github.com/gabriel-vasile/mimetype"
	"github.com/stevealexrs/Go-Libra/random"
)

// remember to listen and serve
type Mock struct {
	store *cache.Cache
	Address string
}

func NewMock() *Mock {
	return &Mock{
		store: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (m *Mock) Handler() http.Handler {
	r := chi.NewRouter()

	r.Get("/{fid}", func(w http.ResponseWriter, r *http.Request) {
		fid := chi.URLParam(r, "fid")
		v, err := m.Get(context.Background(), fid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", mimetype.Detect(v).String())
		w.Write(v)
	})
	return r
}

func (m *Mock) Get(ctx context.Context, fid string) ([]byte, error) {
	if v, found := m.store.Get(fid); found {
		return v.([]byte), nil
	}
	return nil, fmt.Errorf("key %v does not exist", fid)
}

func (m *Mock) Set(ctx context.Context, data io.Reader) (string, error) {
	fid, err := random.Token16Byte()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, data)
	if err != nil {
		return "", err
	} 

	m.store.SetDefault(fid, buf.Bytes())
	return fid, nil
}

func (m *Mock) Delete(ctx context.Context, fids ...string) error {
	for _, v := range fids {
		m.store.Delete(v)
	}
	return nil
}

func (m *Mock) FormatURL(fids ...string) ([]string, error) {
	links := make([]string, 0)
	for _, v := range fids {
		links = append(links, m.Address + "/" + v)
	}
	return links, nil
}

func (m *Mock) FormatFID(objectURL ...string) ([]string, error) {
	fids := make([]string, 0)
	for _, v := range objectURL {
		myUrl, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		fids = append(fids, strings.TrimPrefix(myUrl.Path, "/"))
	}
	return fids, nil
}


