package core

import (
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	controllers "github.com/iter8-tools/etc3/controllers"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

var log *logrus.Logger

var logLevel logrus.Level = logrus.InfoLevel

// SetLogLevel sets level for logging.
func SetLogLevel(l logrus.Level) {
	logLevel = l
	if log != nil {
		log.SetLevel(logLevel)
	}
}

// GetLogger returns a logger, if needed after creating it.
func GetLogger() *logrus.Logger {
	if log == nil {
		log = logrus.New()
		log.SetLevel(logLevel)
		log.SetFormatter(&logrus.TextFormatter{
			DisableQuote: true,
		})
	}
	return log
}

// ContextKey is the type of key that will be used to index into context.
type ContextKey string

// CompletePath determines complete path of a file
var CompletePath func(prefix string, suffix string) string = controllers.CompletePath

// Int32Pointer takes an int32 as input, creates a new variable with the input value, and returns a pointer to the variable
func Int32Pointer(i int32) *int32 {
	return &i
}

// Float32Pointer takes an float32 as input, creates a new variable with the input value, and returns a pointer to the variable
func Float32Pointer(f float32) *float32 {
	return &f
}

// Float64Pointer takes an float64 as input, creates a new variable with the input value, and returns a pointer to the variable
func Float64Pointer(f float64) *float64 {
	return &f
}

// StringPointer takes a string as input, creates a new variable with the input value, and returns a pointer to the variable
func StringPointer(s string) *string {
	return &s
}

// BoolPointer takes a bool as input, creates a new variable with the input value, and returns a pointer to the variable
func BoolPointer(b bool) *bool {
	return &b
}

// HTTPMethod is either GET or POST
type HTTPMethod string

const (
	// GET method
	GET HTTPMethod = "GET"
	// POST method
	POST = "POST"
)

// HTTPMethodPointer takes an HTTPMethod as input, creates a new variable with the input value, and returns a pointer to the variable
func HTTPMethodPointer(h HTTPMethod) *HTTPMethod {
	return &h
}

// WaitTimeoutOrError waits for one of the following three events
// 1) all goroutines in the waitgroup to finish normally -- no error is returned
// 2) a timeout occurred before all go routines could finish normally -- an error is returned
// 3) an error in the errCh channel sent by one of the goroutines -- an error is returned
// See https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
func WaitTimeoutOrError(wg *sync.WaitGroup, timeout time.Duration, errCh chan error) error {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c: // completed normally
		return nil
	case <-time.After(timeout): // timeout
		return errors.New("timed out waiting for go routines to complete") // timed out
	case err := <-errCh: // error in channel
		return err
	}
}

// GetJSONBytes downloads JSON from URL and returns a byte slice
func GetJSONBytes(url string) ([]byte, error) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil || r.StatusCode >= 400 {
		return nil, errors.New("error while fetching payload")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, err
}

// GetTokenFromSecret gets token from k8s secret object
// can be used in notification, gitops and other tasks that use secret tokens
func GetTokenFromSecret(secret *corev1.Secret) (string, error) {
	token := string(secret.Data["token"])
	if token == "" {
		return "", errors.New("empty token in secret")
	}
	return token, nil
}
