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
	dbName          = mustEnvVar("RADSTACK_DB_NAME")
	dbCert          = mustEnvVar("RADSTACK_DB_CERT")
	dbKey           = mustEnvVar("RADSTACK_DB_KEY")
	dbNodesString   = mustEnvVar("RADSTACK_DB_NODES")
	stage           = mustEnvVar("RADSTACK_STAGE")
	orgId           = mustEnvVar("RADSTACK_ORG_ID")
	DocumentSession = newDocumentSession(orgId, dbName, stage)
)

func newDocumentSession(orgId, dbName, stage string) *ravendb.DocumentSession {
	dbId := fmt.Sprintf("%s-%s-%s", mustSanitizedName(orgId), mustSanitizedName(dbName), mustSanitizedName(stage))
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

func newDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := strings.Split(dbNodesString, ",")

	cer, err := tls.X509KeyPair([]byte(dbCert), []byte(dbKey))
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

func mustEnvVar(envVarName string) string {
	envVarVal, hasEnvVarVal := os.LookupEnv(envVarName)
	if !hasEnvVarVal {
		log.Fatalf("Could not look up environmental variable %s\n", envVarName)
	}
	return envVarVal
}

var allowableCharsRegexp = regexp.MustCompile("[^a-z]*")

func sanitizedName(name string) (string, error) {
	s := allowableCharsRegexp.ReplaceAllString(strings.ToLower(name), "")
	if len(s) < 3 {
		return "", fmt.Errorf("name %s after being sanitized to %s is too short, must be atleast 3 chars in length\n", name, s)
	}
	return s, nil
}

func mustSanitizedName(name string) string {
	s, err := sanitizedName(name)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
