// Copyright 2019 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"io"
	"net/http"
	"net/url"

	csapi "github.com/sorintlab/agola/internal/services/configstore/api"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

type ReposHandler struct {
	log               *zap.SugaredLogger
	configstoreClient *csapi.Client
}

func NewReposHandler(logger *zap.Logger, configstoreClient *csapi.Client) *ReposHandler {
	return &ReposHandler{log: logger.Sugar(), configstoreClient: configstoreClient}
}

func (h *ReposHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	path := vars["rest"]

	h.log.Infof("path: %s", path)

	u, err := url.Parse("http://172.17.0.1:4003")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.Path = path
	u.RawQuery = r.URL.RawQuery

	h.log.Infof("u: %s", u.String())
	// TODO(sgotti) Check authorized call from client

	defer r.Body.Close()
	// proxy all the request body to the destination server
	req, err := http.NewRequest(r.Method, u.String(), r.Body)
	req = req.WithContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// copy request headers
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// copy response headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	// copy status
	w.WriteHeader(resp.StatusCode)

	defer resp.Body.Close()
	// copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
