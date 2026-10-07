package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	SiteVerifyURL   = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	DefaultTimeout  = 3 * time.Second
	maxTokenLength  = 2048
	maxAnswerLength = 64 * 1024
)

var (
	errTokenShape      = errors.New("captcha token is empty or too long")
	errServiceStatus   = errors.New("captcha service answered with an error status")
	errNotSuccessful   = errors.New("captcha token was rejected")
	errHostnameChanged = errors.New("captcha was solved on another hostname")
	errActionChanged   = errors.New("captcha was solved for another action")
)

type Options struct {
	Secret           string
	ExpectedHostname string
	Endpoint         string
	Timeout          time.Duration
}

type Challenge struct {
	Token    string
	Action   string
	RemoteIP string
}

type Turnstile struct {
	opts   Options
	client *http.Client
}

func NewTurnstile(opts Options) *Turnstile {
	return &Turnstile{opts: opts, client: &http.Client{Timeout: opts.Timeout}}
}

func (t *Turnstile) Verify(ctx context.Context, challenge Challenge) error {
	if challenge.Token == "" || len(challenge.Token) > maxTokenLength {
		return failed(errTokenShape)
	}
	answer, err := t.siteVerify(ctx, challenge)
	if err != nil {
		return failed(err)
	}
	if err := answer.check(t.opts.ExpectedHostname, challenge.Action); err != nil {
		return failed(err)
	}
	return nil
}

func (t *Turnstile) siteVerify(ctx context.Context, challenge Challenge) (answer siteVerifyAnswer, err error) {
	form := url.Values{"secret": {t.opts.Secret}, "response": {challenge.Token}, "remoteip": {challenge.RemoteIP}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, t.opts.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return siteVerifyAnswer{}, fmt.Errorf("build siteverify request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := t.client.Do(request)
	if err != nil {
		return siteVerifyAnswer{}, fmt.Errorf("call siteverify: %w", err)
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode != http.StatusOK {
		return siteVerifyAnswer{}, fmt.Errorf("%w: %d", errServiceStatus, response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxAnswerLength)).Decode(&answer); err != nil {
		return siteVerifyAnswer{}, fmt.Errorf("decode siteverify answer: %w", err)
	}
	return answer, nil
}

type siteVerifyAnswer struct {
	Success    bool     `json:"success"`
	Hostname   string   `json:"hostname"`
	Action     string   `json:"action"`
	ErrorCodes []string `json:"error-codes"`
	Metadata   struct {
		ResultWithTestingKey bool `json:"result_with_testing_key"`
	} `json:"metadata"`
}

func (a siteVerifyAnswer) check(expectedHostname, expectedAction string) error {
	switch {
	case !a.Success:
		return fmt.Errorf("%w: %s", errNotSuccessful, strings.Join(a.ErrorCodes, ","))
	case a.Metadata.ResultWithTestingKey:
		return nil
	case a.Hostname != expectedHostname:
		return errHostnameChanged
	case a.Action != expectedAction:
		return errActionChanged
	}
	return nil
}

func failed(cause error) error {
	return apperror.New(apperror.CodeCaptchaFailed).WithCause(cause)
}
