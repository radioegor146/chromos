package chromos

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"io"
	"strconv"
	"strings"
)

const (
	googleTimeServiceURL = "http://clients2.google.com/time/1/current"
	googleKeyPubBytes = "\x30\x59\x30\x13\x06\x07\x2A\x86\x48\xCE\x3D\x02\x01\x06\x08\x2A\x86\x48\xCE\x3D\x03\x01\x07\x03\x42\x00\x04\x51\x8B\x06\x03\x4D\xEA\x13\xC3\x32\x9B\x15\x73\xD6\xBC\x47\x33\x3F\xB6\x95\x0E\x5D\x52\x73\x70\x5D\xE4\x92\xBD\xFD\xC5\xB9\xC6\x51\x81\x2D\x8B\x46\xC4\x4C\xB0\xA5\xC6\xDB\x5B\xE4\xDB\x80\x57\x6B\x4D\x08\x9C\x3D\x8B\xC2\xD9\x27\x9A\xDE\x3D\xE2\xCC\x0A\x20"
	googleKeyVersion = 9

	microsoftTimeServiceURL = "http://edge.microsoft.com/browsernetworktime/time/1/current"
	microsoftKeyPubBytes = "\x30\x59\x30\x13\x06\x07\x2A\x86\x48\xCE\x3D\x02\x01\x06\x08\x2A\x86\x48\xCE\x3D\x03\x01\x07\x03\x42\x00\x04\xBB\x37\xA5\xF6\x3A\xF8\x32\x58\x1C\x89\x29\xEC\x3F\x91\x69\x23\x9B\x32\xE3\x35\xDB\x54\xFC\xD8\x8D\xAB\x36\xCD\x68\x71\x95\x50\xDD\xB4\x82\xE6\xF8\x94\xE9\xEB\x3B\x01\x4A\x9E\x15\x71\xBE\x57\x10\x8D\x8C\x1C\x7F\x39\x14\x09\xF9\x63\xD1\xA3\x81\x99\x3D\x22"
	microsoftKeyVersion = 2

	nonceLength = 32
	emptySha256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type TimeServiceConfig struct {
	timeServiceURL string
	keyVersion     int
	keyPubBytes    []byte
}

type timeServiceResponse struct {
	CurrentTimeMillis int64 `json:"current_time_millis"`
}

func generateRequestParams(config TimeServiceConfig) ([]byte, string, error) {
	nonce := make([]byte, nonceLength)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, "", err
	}

	nonceBase64 := base64.RawURLEncoding.EncodeToString(nonce)

	cup2key := strconv.Itoa(config.keyVersion) + ":" + nonceBase64

	return nonce, cup2key, nil
}

func verifyResponse(response *http.Response, config TimeServiceConfig, cup2key string, nonce []byte) ([]byte, error) {
	cupServerProofHeader := response.Header.Get("x-cup-server-proof")
	if cupServerProofHeader == "" {
		return nil, errors.New("no x-cup-server-proof header in response")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}

	parts := strings.Split(cupServerProofHeader, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("x-cup-server-proof is invalid: %s", cupServerProofHeader)
	}

	signature := parts[0]
	requestHash := parts[1]

	if requestHash != emptySha256 {
		return nil, fmt.Errorf("response request hash is invalid: %s != %s", requestHash, emptySha256)
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to dehex signature from response: %v", err)
	}

	hasher := sha256.New()
	requestHashBytes, err := hex.DecodeString(emptySha256)
	if err != nil {
		return nil, err
	}
	hasher.Write(requestHashBytes)
	bodyHash := sha256.Sum256(body)
	hasher.Write(bodyHash[:])
	hasher.Write([]byte(cup2key))

	hashToVerify := hasher.Sum(nil)
	
	publicKey, err := x509.ParsePKIXPublicKey(config.keyPubBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	var ecdsaKey *ecdsa.PublicKey

	switch publicKey := publicKey.(type) {
	case *ecdsa.PublicKey:
		ecdsaKey = publicKey
	default:
		return nil, fmt.Errorf("public key is not ecdsa: %v", publicKey)
	}

	hashedHashToVerify := sha256.Sum256(hashToVerify)

	isValid := ecdsa.VerifyASN1(ecdsaKey, hashedHashToVerify[:], signatureBytes)
	if !isValid {
		return nil, errors.New("signature invalid")
	}

	return body, nil
}

func FetchTime(config TimeServiceConfig) (int64, error) {
	client := &http.Client{}
	
	nonce, cup2key, err := generateRequestParams(config)
	if err != nil {
		return 0, err
	}

	values := url.Values{}
	values.Set("cup2key", cup2key)
	values.Set("cup2hreq", emptySha256)

	response, err := client.Get(fmt.Sprintf("%s?%s", config.timeServiceURL, values.Encode()))
	if err != nil {
		return 0, err
	}

	responseBody, err := verifyResponse(response, config, cup2key, nonce)
	if err != nil {
		return 0, err
	}

	var parsedResponse timeServiceResponse
	err = json.Unmarshal(responseBody[5:], &parsedResponse)
	if err != nil {
		return 0, err
	}

	return parsedResponse.CurrentTimeMillis, nil
}

func GetGoogleConfig() TimeServiceConfig {
	return TimeServiceConfig{
		timeServiceURL: googleTimeServiceURL,
		keyPubBytes: []byte(googleKeyPubBytes),
		keyVersion: googleKeyVersion,
	}
}

func GetMicrosoftConfig() TimeServiceConfig {
	return TimeServiceConfig{
		timeServiceURL: microsoftTimeServiceURL,
		keyPubBytes: []byte(microsoftKeyPubBytes),
		keyVersion: microsoftKeyVersion,
	}
}