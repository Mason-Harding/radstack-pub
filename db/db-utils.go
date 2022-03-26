package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	envDbId          = "RADSTACK_DB_ID" // don't normally set this, rather use dbName, stage, and orgId
	envDbName        = "RADSTACK_DB_NAME"
	envDbCert        = "RADSTACK_DB_CERT"
	envDbKey         = "RADSTACK_DB_KEY"
	envDbNodesString = "RADSTACK_DB_NODES"
	envStage         = "RADSTACK_STAGE"
	envOrgId         = "RADSTACK_ORG_ID"
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
	if dbId := os.Getenv(envDbId); dbId != "" {
		return dbId
	} else {
		return fmt.Sprintf("%s-%s-%s", mustSanitizedEnvVar(envOrgId), mustSanitizedEnvVar(envDbName), mustSanitizedEnvVar(envStage))
	}
}

func newDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := strings.Split(envDbNodesString, ",")

	cer, err := tls.X509KeyPair([]byte(envDbCert), []byte(envDbKey))
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

func printRQL(q *ravendb.DocumentQuery) {
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

var allowableCharsRegexp = regexp.MustCompile("[^a-z]*")

func sanitizedEnvVar(e string) (string, error) {
	s := allowableCharsRegexp.ReplaceAllString(strings.ToLower(e), "")
	if len(s) < 3 {
		return "", fmt.Errorf("For env var %s, after being sanitized to %s is too short, must be atleast 3 chars in length\n", e, s)
	}
	return s, nil
}

func mustSanitizedEnvVar(name string) string {
	s, err := sanitizedEnvVar(name)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
