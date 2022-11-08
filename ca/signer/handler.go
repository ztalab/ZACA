/*
Copyright 2022-present The Ztalab Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package signer

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/ztalab/cfssl/api"
	"github.com/ztalab/cfssl/auth"
	"github.com/ztalab/cfssl/bundler"
	"github.com/ztalab/cfssl/errors"
	"github.com/ztalab/cfssl/helpers"
	"github.com/ztalab/cfssl/hook"
	"github.com/ztalab/cfssl/log"
	"github.com/ztalab/cfssl/signer"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/database/mysql/cfssl-model/dao"
	"github.com/ztalab/ZACA/logic/events"
	"github.com/ztalab/ZACA/pkg/spiffe"
)

// NoBundlerMessage is used to alert the user that the server does not have a bundler initialized.
const NoBundlerMessage = `This request requires a bundler, but one is not initialized for the API server.`

// A Handler accepts requests with a hostname and certficate
// parameter (which should be PEM-encoded) and returns a new signed
// certificate. It includes upstream servers indexed by their
// profile name.
type Handler struct {
	signer  signer.Signer
	bundler *bundler.Bundler
}

// NewHandlerFromSigner generates a new Handler directly from
// an existing signer.
func NewHandlerFromSigner(signer signer.Signer) (h *api.HTTPHandler, err error) {
	policy := signer.Policy()
	if policy == nil {
		err = errors.New(errors.PolicyError, errors.InvalidPolicy)
		return
	}

	// Sign will only respond for profiles that have no auth provider.
	// So if all of the profiles require authentication, we return an error.
	haveUnauth := policy.Default.Provider == nil
	for _, profile := range policy.Profiles {
		haveUnauth = haveUnauth || (profile.Provider == nil)
	}

	if !haveUnauth {
		err = errors.New(errors.PolicyError, errors.InvalidPolicy)
		return
	}

	return &api.HTTPHandler{
		Handler: &Handler{
			signer: signer,
		},
		Methods: []string{"POST"},
	}, nil
}

// SetBundler allows injecting an optional Bundler into the Handler.
func (h *Handler) SetBundler(caBundleFile, intBundleFile string) (err error) {
	h.bundler, err = bundler.NewBundler(caBundleFile, intBundleFile)
	return err
}

// This type is meant to be unmarshalled from JSON so that there can be a
// hostname field in the API
// TODO: Change the API such that the normal struct can be used.
type jsonSignRequest struct {
	Hostname string                 `json:"hostname"`
	Hosts    []string               `json:"hosts"`
	Request  string                 `json:"certificate_request"`
	Subject  *signer.Subject        `json:"subject,omitempty"`
	Profile  string                 `json:"profile"`
	Label    string                 `json:"label"`
	Serial   *big.Int               `json:"serial,omitempty"`
	Bundle   bool                   `json:"bundle"`
	Metadata map[string]interface{} `json:"metadata"`
}

func jsonReqToTrue(js jsonSignRequest) signer.SignRequest {
	sub := new(signer.Subject)
	if js.Subject == nil {
		sub = nil
	} else {
		// make a copy
		*sub = *js.Subject
	}

	if js.Hostname != "" {
		return signer.SignRequest{
			Hosts:    signer.SplitHosts(js.Hostname),
			Subject:  sub,
			Request:  js.Request,
			Profile:  js.Profile,
			Label:    js.Label,
			Serial:   js.Serial,
			Metadata: js.Metadata,
		}
	}

	return signer.SignRequest{
		Hosts:    js.Hosts,
		Subject:  sub,
		Request:  js.Request,
		Profile:  js.Profile,
		Label:    js.Label,
		Serial:   js.Serial,
		Metadata: js.Metadata,
	}
}

// Handle responds to requests for the CA to sign the certificate request
// present in the "certificate_request" parameter for the host named
// in the "hostname" parameter. The certificate should be PEM-encoded. If
// provided, subject information from the "subject" parameter will be used
// in place of the subject information from the CSR.
// The Handler will not be called, we usually use the following AuthHandler
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	log.Info("signature request received")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	var req jsonSignRequest

	err = json.Unmarshal(body, &req)
	if err != nil {
		return errors.NewBadRequestString("Unable to parse sign request")
	}

	signReq := jsonReqToTrue(req)

	if req.Request == "" {
		return errors.NewBadRequestString("missing parameter 'certificate_request'")
	}

	var cert []byte
	profile, err := signer.Profile(h.signer, req.Profile)
	if err != nil {
		return err
	}

	if profile.Provider != nil {
		log.Error("profile requires authentication")
		return errors.NewBadRequestString("authentication required")
	}

	cert, err = h.signer.Sign(signReq)
	if err != nil {
		log.Warningf("failed to sign request: %v", err)
		return err
	}

	result := map[string]interface{}{"certificate": string(cert)}
	if req.Bundle {
		if h.bundler == nil {
			return api.SendResponseWithMessage(w, result, NoBundlerMessage,
				errors.New(errors.PolicyError, errors.InvalidRequest).ErrorCode)
		}

		bundle, err := h.bundler.BundleFromPEMorDER(cert, nil, bundler.Optimal, "")
		if err != nil {
			return err
		}

		result["bundle"] = bundle
	}
	log.Info("wrote response")
	return api.SendResponse(w, result)
}

// An AuthHandler verifies and signs incoming signature requests.
type AuthHandler struct {
	signer  signer.Signer
	bundler *bundler.Bundler
}

// NewAuthHandlerFromSigner creates a new AuthHandler from the signer
// that is passed in.
// issued the certificate API Handler
func NewAuthHandlerFromSigner(signer signer.Signer) (http.Handler, error) {
	policy := signer.Policy()
	if policy == nil {
		return nil, errors.New(errors.PolicyError, errors.InvalidPolicy)
	}

	if policy.Default == nil && policy.Profiles == nil {
		return nil, errors.New(errors.PolicyError, errors.InvalidPolicy)
	}

	// AuthSign will not respond for profiles that have no auth provider.
	// So if there are no profiles with auth providers in this policy,
	// we return an error.
	haveAuth := policy.Default.Provider != nil
	for _, profile := range policy.Profiles {
		if haveAuth {
			break
		}
		haveAuth = profile.Provider != nil
	}

	if !haveAuth {
		return nil, errors.New(errors.PolicyError, errors.InvalidPolicy)
	}

	return &api.HTTPHandler{
		Handler: &AuthHandler{
			signer: signer,
		},
		Methods: []string{"POST"},
	}, nil
}

// SetBundler allows injecting an optional Bundler into the Handler.
func (h *AuthHandler) SetBundler(caBundleFile, intBundleFile string) (err error) {
	h.bundler, err = bundler.NewBundler(caBundleFile, intBundleFile)
	return err
}

// Handle receives the incoming request, validates it, and processes it.
// Process signed certificate requests for authentication
func (h *AuthHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	log.Info("signature request received")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read response body: %v", err)
		return err
	}
	defer r.Body.Close()

	var aReq auth.AuthenticatedRequest
	err = json.Unmarshal(body, &aReq)
	if err != nil {
		log.Errorf("failed to unmarshal authenticated request: %v", err)
		return errors.NewBadRequest(err)
	}

	var req jsonSignRequest
	err = json.Unmarshal(aReq.Request, &req)
	if err != nil {
		log.Errorf("failed to unmarshal request from authenticated request: %v", err)
		return errors.NewBadRequestString("Unable to parse authenticated sign request")
	}

	// Sanity checks to ensure that we have a valid policy. This
	// should have been checked in NewAuthHandler.
	policy := h.signer.Policy()
	if policy == nil {
		log.Critical("signer was initialised without a signing policy")
		return errors.NewBadRequestString("invalid policy")
	}

	profile, err := signer.Profile(h.signer, req.Profile)
	if err != nil {
		return err
	}

	if profile.Provider == nil {
		log.Error("profile has no authentication provider")
		return errors.NewBadRequestString("no authentication provider")
	}

	validAuth := false
	if profile.Provider.Verify(&aReq) {
		validAuth = true
	} else if profile.PrevProvider != nil && profile.PrevProvider.Verify(&aReq) {
		validAuth = true
	}
	if !validAuth {
		log.Warning("received authenticated request with invalid token")
		return errors.NewBadRequestString("invalid token")
	}

	signReq := jsonReqToTrue(req)

	if signReq.Request == "" {
		return errors.NewBadRequestString("missing parameter 'certificate_request'")
	}

	// Can audit apply for certificate
	// Query whether the DB is marked with uniqueID to prohibit application
	for _, signHost := range signReq.Hosts {
		id, err := spiffe.ParseIDGIdentity(signHost)
		if err != nil {
			continue
		}

		if id.UniqueID != "" {
			query := core.Is.Db.Where("unique_id = ?", id.UniqueID).
				Where("deleted_at IS NULL")
			record, err := dao.GetForbid(query)
			if err == nil && record != nil {
				events.NewWorkloadLifeCycle("forbid-sign", events.OperatorSDK, events.CertOp{
					UniqueId: id.UniqueID,
				}).Log()
				return errors.NewBadRequestString("unique_id forbidden for signing certs")
			}
		}
	}

	// CFSSL In the issuing logic, if the certificate storage mode is vault, the database flag bit is added, and the certificate PEM is not actually stored
	cert, err := h.signer.Sign(signReq)
	if err != nil {
		log.Errorf("signature failed: %v", err)
		return err
	}

	x509Cert, _ := helpers.ParseCertificatePEM(cert)

	// After the certificate is issued, it is added and stored in the vault
	if hook.EnableVaultStorage {
		if err := core.Is.VaultSecret.StoreCertPEM(x509Cert.SerialNumber.String(), string(cert)); err != nil {
			core.Is.Logger.Errorf("vault store err: %s", err)
			return err
		}
	}

	// Metrics Timing record
	AddMetricsPoint(x509Cert)

	if x509Cert != nil {
		events.NewWorkloadLifeCycle("sign", events.OperatorSDK, events.CertOp{
			UniqueId: x509Cert.Subject.CommonName,
			SN:       x509Cert.SerialNumber.String(),
			AKI:      hex.EncodeToString(x509Cert.AuthorityKeyId),
		}).Log()
	}

	result := map[string]interface{}{"certificate": string(cert)}
	if req.Bundle {
		if h.bundler == nil {
			return api.SendResponseWithMessage(w, result, NoBundlerMessage,
				errors.New(errors.PolicyError, errors.InvalidRequest).ErrorCode)
		}

		bundle, err := h.bundler.BundleFromPEMorDER(cert, nil, bundler.Optimal, "")
		if err != nil {
			return err
		}

		result["bundle"] = bundle
	}
	log.Info("wrote response")
	return api.SendResponse(w, result)
}
