// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emetsger/negtracker/id"
	"github.com/emetsger/negtracker/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type verifier struct {
	Attempts    int
	Client      *http.Client
	ResVerifier func(t *testing.T, res *http.Response)
}

func (v *verifier) verifyFunc(f func(t *testing.T, res *http.Response)) *verifier {
	v.ResVerifier = f
	return v
}

var MyVerifier *verifier

var sampleNeg = model.Neg{
	Id:          "negId",
	Film:        "FP4",
	EI:          100,
	Developer:   "Pyrocat HD",
	FrameNumber: "8",
	Tags:        []string{"druid hill", "daffodil", "spring"},
	Description: "Druid Hill",
	Format:      "120",
}

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
		fmt.Sprintf("%s/Ping", config.ListenUrl()),
		nil)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		buf := bytes.Buffer{}
		_, _ = io.Copy(&buf, res.Body)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "Pong!", buf.String())
	}).attempt(req, t)
}

// test creating a Neg
func Test_ServerNegPost(t *testing.T) {
	neg := sampleNeg
	neg.Id = id.Mint()
	body, err := json.Marshal(neg)
	assert.Nil(t, err)
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/neg", config.ListenUrl()),
		bytes.NewBuffer(body))

	// TODO accept and content-type header support/verification
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 201, res.StatusCode)
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		assert.NotEqual(t, "", res.Header.Get("Content-Length"))
		atoi, _ := strconv.Atoi(res.Header.Get("Content-Length"))
		assert.True(t, atoi > 0)
	}).attempt(req, t)
}

// test creating a neg with an absent ID field, should be populated
func Test_ServerNegPostNoId(t *testing.T) {
	body := bytes.NewBufferString(`{"Film": "Moo"}`)
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/neg", config.ListenUrl()),
		body)

	// TODO accept and content-type header support/verification
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))

	var id string

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 201, res.StatusCode)
		id = asString(res.Body)
	}).attempt(req, t)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/neg/%s", config.ListenUrl(), id), nil)
	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 200, res.StatusCode)
		created := &model.Neg{}
		json.Unmarshal(asByte(res.Body), created)
		assert.NotEqual(t, "", created.Id)
	}).attempt(req, t)
}

// test creating a neg with an empty ID field, should be populated
func Test_ServerNegPostEmptyId(t *testing.T) {
	body := bytes.NewBufferString(`{"Id": "", "Film": "Moo"}`)
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/neg", config.ListenUrl()),
		body)

	// TODO accept and content-type header support/verification
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))

	var id string

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 201, res.StatusCode)
		id = asString(res.Body)
	}).attempt(req, t)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/neg/%s", config.ListenUrl(), id), nil)
	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 200, res.StatusCode)
		created := &model.Neg{}
		json.Unmarshal(asByte(res.Body), created)
		assert.NotEqual(t, "", created.Id)
	}).attempt(req, t)
}

// test creating a neg with a duplicate ID field, should be rejected
func Test_ServerNegPostDuplicateId(t *testing.T) {
	id := id.Mint()
	body := fmt.Sprintf(`{"Id": "%s", "Film": "Moo"}`, id)
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/neg", config.ListenUrl()),
		bytes.NewBufferString(body))

	// TODO accept and content-type header support/verification
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 201, res.StatusCode)
		assert.Equal(t, id, asString(res.Body))
	}).attempt(req, t)

	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/neg", config.ListenUrl()),
		bytes.NewBufferString(body))

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 409, res.StatusCode)
		body := asString(res.Body)
		assert.True(t, len(body) > 0)
		log.Printf("Conflict request body: %s", body)
	}).attempt(req, t)
}

// TODO fix ids - test retrieving a Neg
func Test_ServerNegGet(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/neg", config.ListenUrl()),
		bytes.NewBufferString(
			`{
					"Id": "",
                    "Film": "FP4"
				}`))

	var err error
	var id string
	var created *model.Neg
	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		require.Equal(t, 201, res.StatusCode)
		// FIXME post should be setting a Location header and we should be reading that
		id = asString(res.Body)
		require.NotEqual(t, "", id)
	}).attempt(req, t)

	log.Printf("Created neg with id %s", id)

	req, err = http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/neg/%s", config.ListenUrl(), id), nil)

	require.NotNil(t, req)
	require.Nil(t, err)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		require.Equal(t, 200, res.StatusCode)

		// Etag tests
		etag := res.Header.Get("ETag")
		require.NotNil(t, etag)
		require.True(t, len(etag) > 0)
		// etag is strong for now
		require.False(t, strings.HasPrefix(etag, "W/"))
		require.True(t, strings.HasPrefix(etag, "\""))
		require.True(t, strings.HasSuffix(etag, "\""))

		// Body tests
		require.NotNil(t, res.Body)
		created = &model.Neg{}
		json.Unmarshal(asByte(res.Body), created)

		// Id was populated
		require.True(t, len(created.Id) > 0)

		// We have the neg we expect - the one we posted above
		require.Equal(t, "FP4", created.Film)
		require.Equal(t, "", created.Format)

		// Create and Updated time was populated, and not equal to the zero value
		require.True(t, time.Now().After(created.Created))
		require.True(t, time.Now().After(created.Updated))
		require.False(t, created.Created.IsZero())
		require.False(t, created.Updated.IsZero())
		log.Printf("*** %v ***", created)
	}).attempt(req, t)
}

func Test_ServerNegNotImpl(t *testing.T) {
	req, _ := http.NewRequest(http.MethodTrace,
		fmt.Sprintf("%s/neg", config.ListenUrl()),
		nil)

	MyVerifier.verifyFunc(func(t *testing.T, res *http.Response) {
		assert.Equal(t, 500, res.StatusCode)
		b := &bytes.Buffer{}
		io.Copy(b, res.Body)
		assert.True(t, strings.Contains(b.String(), "not implemented"))

	}).attempt(req, t)
}

func (v *verifier) attempt(req *http.Request, t *testing.T) {
	require.NotNil(t, req, "Request must not be nil.")
	require.NotNil(t, req.URL, "Request URL must not be nil.")

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
		require.Nil(t, err, "Error executing query %s %s after %v attempts: %s", req.Method, req.URL.String(), times, err.Error())
	}

	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()

	v.ResVerifier(t, res)
}

func asString(reader io.Reader) string {
	if b, err := ioutil.ReadAll(reader); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

func asByte(reader io.Reader) []byte {
	if b, err := ioutil.ReadAll(reader); err != nil {
		panic(err)
	} else {
		return b
	}
}
