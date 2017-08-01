// Copyright 2017 The Kubernetes Dashboard Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jwe

import (
	"crypto/rand"
	"crypto/rsa"
	"log"

	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	jose "gopkg.in/square/go-jose.v2"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Implements TokenManager interface
type jweTokenManager struct {
	encrypter jose.Encrypter
	// TODO(floreks): Add key synchronization (between dashboard replicas), expiration and rotation options
	// 256-byte random RSA key pair. It is generated during the first backend start.
	tokenEncryptionKey *rsa.PrivateKey
}

// Generate and encrypt JWE token based on provided AuthInfo structure. AuthInfo will be embedded in a token payload and
// encrypted with autogenerated signing key.
// Used encryption alghoritms:
//    - Content encryption: AES-GCM (256)
//    - Key management: RSA-OAEP-SHA256
func (self *jweTokenManager) Generate(authInfo api.AuthInfo) (string, error) {
	marshalledAuthInfo, err := json.Marshal(authInfo)
	if err != nil {
		return "", err
	}

	// TODO(floreks): add token expiration header and handle it
	jweObject, err := self.encrypter.Encrypt(marshalledAuthInfo)
	if err != nil {
		return "", err
	}

	return jweObject.FullSerialize(), nil
}

// Decrypt provided token and return AuthInfo structure saved in a token payload.
func (self *jweTokenManager) Decrypt(jweToken string) (*api.AuthInfo, error) {
	jweTokenObject, err := self.validate(jweToken)
	if err != nil {
		return nil, err
	}

	decrypted, err := jweTokenObject.Decrypt(self.tokenEncryptionKey)
	// TODO(floreks): Check for decryption error and handle it
	if err != nil {
		return nil, err
	}

	authInfo := new(api.AuthInfo)
	err = json.Unmarshal(decrypted, authInfo)
	return authInfo, err
}

// Parses and validates provided token to check if it hasn't been manipulated with.
func (self *jweTokenManager) validate(jweToken string) (*jose.JSONWebEncryption, error) {
	// TODO(floreks): validate token expiration
	return jose.ParseEncrypted(jweToken)
}

// Initializes token manager instance.
func (self *jweTokenManager) init() {
	self.initEncryptionKey()
	self.initEncrypter()
}

// Generates encryption key used to encrypt token payload.
func (self *jweTokenManager) initEncryptionKey() {
	log.Print("Generating JWE encryption key")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	self.tokenEncryptionKey = privateKey
}

// Creates encrypter instance based on generated encryption key.
func (self *jweTokenManager) initEncrypter() {
	log.Print("Initializing encrypter")
	publicKey := &self.tokenEncryptionKey.PublicKey
	encrypter, err := jose.NewEncrypter(jose.A256GCM, jose.Recipient{Algorithm: jose.RSA_OAEP_256, Key: publicKey}, nil)
	if err != nil {
		panic(err)
	}

	self.encrypter = encrypter
}

// Creates and returns JWE token manager instance.
func NewJWETokenManager() authApi.TokenManager {
	manager := &jweTokenManager{}
	manager.init()
	return manager
}
