package webhook

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	ps "battleblocks-graffle/pubsub"
	"cloud.google.com/go/pubsub"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type WebhookServer struct {
	pubsubClient  *ps.PubsubClient
	graffleConfig GraffleConfig
}

type GraffleConfig struct {
	CompanyID string
	Secret    []byte
}

type GraffleAuthorization struct {
	CompanyID              string
	Base64RequestSignature string
	Nonce                  string
	RequestTimestamp       time.Time
}

var (
	nonceMu sync.Mutex
	nonces  = make(map[string]time.Time)
)

func NewWebhookServer(pubsubClient *ps.PubsubClient) *WebhookServer {
	viper.SetDefault("GRAFFLE_COMPANY_ID", "ead2dbd7-47e5-458a-bb65-f2fe4f0dfee2")
	viper.SetDefault("GRAFFLE_SECRET", "3ASDchangemeASD33333")

	secret, err := base64.StdEncoding.DecodeString(viper.GetString("GRAFFLE_SECRET"))

	if err != nil {
		log.Fatal().Msg("Can't decode Graffle secret, is it encoded in Base64?")
	}

	graffleConfig := GraffleConfig{
		CompanyID: viper.GetString("GRAFFLE_COMPANY_ID"),
		Secret:    secret,
	}

	return &WebhookServer{
		pubsubClient:  pubsubClient,
		graffleConfig: graffleConfig,
	}
}

func (ws WebhookServer) Start() {
	router := gin.New()
	router.Use(gin.Recovery())

	// GCP health check
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	global := router.Group("/graffle-webhook-server")
	{
		global.POST("/event", ws.webhookHandler)
	}

	router.Run()
}

func (ws *WebhookServer) webhookHandler(c *gin.Context) {
	rawBody, _ := ioutil.ReadAll(c.Request.Body)
	var body map[string]interface{}
	json.Unmarshal(rawBody, &body)
	authHeader := c.GetHeader("Authorization")

	log.Info().
		Interface("headers", c.Request.Header).
		Str("body_raw", string(rawBody)).
		Interface("body_json", &body).
		Msg("Received event from Graffle, checking Authorization header")

	err := ws.ensureAuth(authHeader, rawBody, c.Request.Method)

	if err != nil {
		log.Warn().Msg(err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	evtData, err := json.Marshal(body["blockEventData"])
	if err != nil {
		log.Warn().Msg(err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	ws.pubsubClient.Publish(c.Request.Context(), &pubsub.Message{
		Data: evtData,
		Attributes: map[string]string{
			"eventType": body["flowEventId"].(string),
		},
	})

	c.Status(http.StatusOK)
}

func (ws WebhookServer) ensureAuth(authHeader string, rawBody []byte, requestMethod string) error {
	var body map[string]interface{}
	json.Unmarshal(rawBody, &body)

	if len(authHeader) == 0 || !strings.HasPrefix(authHeader, "hmacauth") {
		return errors.New("authorization header missing or in the wrong format")
	}

	log.Info().Interface("authheader", authHeader).Msg("Authheader!")
	raw := authHeader[9:]
	parts := strings.Split(raw, ":")

	if len(parts) != 4 {
		return errors.New("authorization header in the wrong format")
	}

	requestTimestamp, _ := strconv.ParseInt(parts[3], 10, 64)

	graffleAuthorization := GraffleAuthorization{
		CompanyID:              parts[0],
		Base64RequestSignature: parts[1],
		Nonce:                  parts[2],
		RequestTimestamp:       time.Unix(requestTimestamp, 0),
	}

	if graffleAuthorization.CompanyID != ws.graffleConfig.CompanyID {
		return errors.New("company id from Authorization header does not match configured company id")
	}

	if isReplayRequest(graffleAuthorization) {
		return errors.New("possible replay attack")
	}

	hash := md5.New()
	hash.Write(rawBody)
	b64 := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	signatureParts := []string{
		graffleAuthorization.CompanyID,
		requestMethod,
		strings.ToLower(url.QueryEscape(body["webHook"].(string))),
		strconv.FormatInt(graffleAuthorization.RequestTimestamp.Unix(), 10),
		graffleAuthorization.Nonce,
		b64,
	}
	signatureData := strings.Join(signatureParts, "")

	hash = hmac.New(sha256.New, ws.graffleConfig.Secret)
	hash.Write([]byte(signatureData))
	b64 = base64.StdEncoding.EncodeToString(hash.Sum(nil))

	log.Info().Interface("graffleAuth", graffleAuthorization).
		Interface("expectedSignature", graffleAuthorization.Base64RequestSignature).
		Interface("graffleConfig", ws.graffleConfig).
		Interface("presentSignature", b64).Msg("Signature mismatch?")

	if graffleAuthorization.Base64RequestSignature != b64 {
		return errors.New("signature from Authorization header does not match generated signature")
	}

	log.Info().Msg("Authorization header validated successfully")
	return nil
}

func isReplayRequest(graffleAuthorization GraffleAuthorization) bool {
	nonceMu.Lock()
	defer nonceMu.Unlock()

	if _, ok := nonces[graffleAuthorization.Nonce]; ok {
		log.Warn().
			Str("nonce", graffleAuthorization.Nonce).
			Msg("Nonce from Authorization header has already been used, possible replay attack")

		return true
	} else {
		cleanupNonces(nonces)
		nonces[graffleAuthorization.Nonce] = time.Now()
	}

	diff := time.Now().Sub(graffleAuthorization.RequestTimestamp)

	if diff.Seconds() > 5 {
		log.Warn().
			Float64("diff", diff.Seconds()).
			Msg("Request timestamp from Authorization header differs from current timestamp by more than 5 seconds")

		return true
	}
	return false
}

func cleanupNonces(nonces map[string]time.Time) {
	now := time.Now()

	for nonce, timestamp := range nonces {
		diff := now.Sub(timestamp)

		if diff.Hours() > 24 {
			delete(nonces, nonce)
		}
	}
}
