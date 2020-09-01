package neghandler

import (
	"bytes"
	"fmt"
	"github.com/emetsger/negtracker/urlutil/strip"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type verifier struct {
	TestState   *testing.T
	Attempts    int
	Client      *http.Client
	ResVerifier func(t *testing.T, res *http.Response)
}

func (v *verifier) verifyFunc(f func(t *testing.T, res *http.Response)) *verifier {
	v.ResVerifier = f
	return v
}

func (v *verifier) testState(t *testing.T) {
	v.TestState = t
}

var MyVerifier *verifier

func TestMain(m *testing.M) {
	defer func() {
		stop(s)
	}()

	go main()

	// wait for the server to get into the running state
	times := 5
	startTime := time.Now()
	for state != RUNNING && times > 0 {
		time.Sleep(500 * time.Millisecond)
		times--
	}

	if state != RUNNING {
		panic(fmt.Sprintf("Server has not started after %v s", time.Since(startTime).Seconds()))
	}

	MyVerifier = &verifier{
		Attempts: 5,
		Client:   &http.Client{},
	}

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func Test_ServerMain(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/Ping", strip.TrailingSlashes(config.ListenUrl())),
		nil)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		buf := bytes.Buffer{}
		_, _ = io.Copy(&buf, res.Body)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "Pong!", buf.String())
	}).testState(t)

	attempt(req, MyVerifier)
}

// test creating a Neg
func Test_ServerNegPost(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/neg", strip.TrailingSlashes(config.ListenUrl())),
		nil)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 201, res.StatusCode)
	}).testState(t)

	attempt(req, MyVerifier)
}

// test retrieving a Neg
func Test_ServerNegGet(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/neg", strip.TrailingSlashes(config.ListenUrl())),
		nil)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 200, res.StatusCode)
	}).testState(t)

	attempt(req, MyVerifier)
}

func Test_ServerNegNotImpl(t *testing.T) {
	req, _ := http.NewRequest(http.MethodTrace,
		fmt.Sprintf("%s/neg", strip.TrailingSlashes(config.ListenUrl())),
		nil)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 500, res.StatusCode)
		b := &bytes.Buffer{}
		io.Copy(b, res.Body)
		assert.True(t, strings.Contains(b.String(), "not implemented"))

	}).testState(t)

	attempt(req, MyVerifier)
}

func attempt(req *http.Request, v *verifier) {
	var res *http.Response
	var err error
	times := v.Attempts

	if times < 1 {
		panic("times must be a positive integer")
	}
	log.Printf("Executing %s %s", req.Method, req.URL)
	res, err = v.Client.Do(req)
	times--

	for err != nil && times > 0 {
		time.Sleep(1 * time.Second)
		log.Printf("Attempt (%v, %v)", times, err)
		res, err = v.Client.Do(req)
		times--
	}

	if err != nil {
		assert.Nil(v.TestState, err, "Error executing query %s %s after %v attempts: %s", req.Method, req.URL.String(), times, err.Error())
	}

	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()

	v.ResVerifier(v.TestState, res)
}
