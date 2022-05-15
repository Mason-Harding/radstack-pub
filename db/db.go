package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/radstack/radstack-go-client/config"
	"github.com/ravendb/ravendb-go-client"
	"log"
	"regexp"
	"sort"
	"strings"
)

var (
	keyDbId          = "RADSTACK_DB_ID"
	keyDbName        = "RADSTACK_DB_NAME"
	keyDbCert        = "RADSTACK_DB_CERT"
	keyDbKey         = "RADSTACK_DB_KEY"
	keyDbNodesString = "RADSTACK_DB_NODES"
	keyStage         = "RADSTACK_STAGE"
	keyOrgId         = "RADSTACK_ORG_ID"
	DocumentSession  = newDocumentSession()
)

func newDocumentSession() *ravendb.DocumentSession {
	dbId := Id()
	store, err := newDocumentStore(dbId)
	if err != nil {
		log.Fatalf("newDocumentStore for db %s failed with %s\n", dbId, err)
	}

	session, err := store.OpenSession(dbId)
	if err != nil {
		log.Fatalf("OpenSession for db %s failed with %s\n", dbId, err)
	}
	return session
}

func Id() string {
	if dbId := configVal(keyDbId); dbId != nil {
		return *dbId
	} else {
		return fmt.Sprintf("%s-%s-%s", mustSanitizedConfigValue(keyOrgId), mustSanitizedConfigValue(keyDbName), mustSanitizedConfigValue(keyStage))
	}
}

func newDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := strings.Split(mustConfigVal(keyDbNodesString), ",")

	cer, err := tls.X509KeyPair([]byte(mustConfigVal(keyDbCert)), []byte(mustConfigVal(keyDbKey)))
	if err != nil {
		return nil, err
	}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	store.Certificate = &cer
	x509cert, err2 := x509.ParseCertificate(cer.Certificate[0])
	if err2 != nil {
		return nil, err2
	}
	store.TrustStore = x509cert
	if err3 := store.Initialize(); err3 != nil {
		return nil, err3
	}
	return store, nil
}

func PrintRQL(q *ravendb.DocumentQuery) {
	iq, err := q.GetIndexQuery()
	if err != nil {
		log.Fatalf("q.GetIndexQuery() returned '%s'\n", err)
	}
	fmt.Printf("RQL: %s\n", iq.GetQuery())
	params := iq.GetQueryParameters()
	if len(params) == 0 {
		return
	}
	fmt.Printf("Parameters:\n")
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  $%s: %#v\n", key, params[key])
	}
	fmt.Print("\n")
}

var c *config.Config

func configVal(name string) *string {
	if c == nil {
		c = config.NewConfig()
	}
	v := c.GetValue(name)
	return v
}

func mustConfigVal(name string) string {
	if c == nil {
		c = config.NewConfig()
	}
	v := c.MustGetValue(name)
	return v
}

var allowableCharsRegexp = regexp.MustCompile("[^a-z]*")

func sanitizedConfigValue(name string) (string, error) {
	v := configVal(name)
	if v == nil {
		return "", fmt.Errorf("Could not look up config value %s\n", name)
	}
	vLower := allowableCharsRegexp.ReplaceAllString(strings.ToLower(*v), "")
	if len(vLower) < 3 {
		return "", fmt.Errorf("For env var %s, after being sanitized to %s is too short, must be atleast 3 chars in length\n", name, vLower)
	}
	return vLower, nil
}

func mustSanitizedConfigValue(name string) string {
	s, err := sanitizedConfigValue(name)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
