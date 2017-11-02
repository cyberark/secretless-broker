package authenticator

import (
  "bytes"
  "io/ioutil"
  "log"
  "net/http"
  "strconv"
  "time"

  "github.com/kgilpin/secretless/config"

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

  var accessKeyId, secretAccessKey string
  accessToken := ""
  for _, v := range self.Config.Backend {
    name := v.Name
    value := v.Value
    if name == "access_key_id" && value != "" {
      accessKeyId = value
    }
    if name == "secret_access_key" && value != "" {
      secretAccessKey = value
    }
  }

  // TODO: remove
  /*
  log.Printf("AccessKeyId: %s", accessKeyId)
  log.Printf("SecretAccessKey: %s", secretAccessKey)
  log.Printf("AccessToken: %s", accessToken)
  */

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
