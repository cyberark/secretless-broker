package authenticator

import (
  "bytes"
  "io/ioutil"
  "fmt"
  "log"
  "net/http"
  "strconv"
  "time"

  "github.com/kgilpin/secretless/config"
  "github.com/kgilpin/secretless/variable"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/aws/signer/v4"
)

type AWSAuthenticator struct {
  Config config.ListenerConfig
}

func (self AWSAuthenticator) Authenticate(r *http.Request) error {
  // TODO: We won't want to sign this when the original request is not signed either. See "AnonymousCredentials"
  // TODO: support credentials from an IAM Role

  var accessKeyId, secretAccessKey, accessToken string

  if valuesPtr, err := variable.Resolve(self.Config.Backend); err != nil {
    return err
  } else {
    values := *valuesPtr

    accessKeyId = values["access_key_id"]
    if accessKeyId == "" {
      return fmt.Errorf("AWS connection parameter 'access_key_id' is not available")
    }
    secretAccessKey = values["secret_access_key"]
    if secretAccessKey == "" {
      return fmt.Errorf("AWS connection parameter 'secret_access_key' is not available")
    }
    accessToken = values["access_token"]
  }

  creds := credentials.NewStaticCredentials(accessKeyId, secretAccessKey, accessToken)

  bodyBytes, err := ioutil.ReadAll(r.Body)
  if err != nil {
    return err
  }

  r.Header.Del("Content-Length")

  signer := v4.NewSigner(creds)
  signer.Debug = aws.LogDebugWithSigning
  signer.Logger = aws.NewDefaultLogger()
  if _, err := signer.Sign(r, bytes.NewReader(bodyBytes), "ec2", "us-east-1", time.Now()); err != nil {
    return err
  }

  r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

  r.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))

  log.Printf("Signed headers: %v", r.Header)

  return nil
}
