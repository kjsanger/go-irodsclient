package testcases

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/cyverse/go-irodsclient/irods/connection"
	"github.com/cyverse/go-irodsclient/irods/fs"
	"github.com/cyverse/go-irodsclient/irods/types"
	"github.com/stretchr/testify/assert"
)

func TestFSAPI(t *testing.T) {
	setup()

	t.Run("test PrepareSamples", testPrepareSamples)

	t.Run("test GetIRODSCollection", testGetIRODSCollection)
	t.Run("test ListIRODSCollections", testListIRODSCollections)
	t.Run("test ListIRODSCollectionMeta", testListIRODSCollectionMeta)
	t.Run("test ListIRODSCollectionAccess", testListIRODSCollectionAccess)
	t.Run("test ListIRODSDataObjects", testListIRODSDataObjects)
	t.Run("test ListIRODSDataObjectsMasterReplica", testListIRODSDataObjectsMasterReplica)
	t.Run("test GetIRODSDataObject", testGetIRODSDataObject)
	t.Run("test GetIRODSDataObjectMasterReplica", testGetIRODSDataObjectMasterReplica)
	t.Run("test ListIRODSDataObjectMeta", testListIRODSDataObjectMeta)
	t.Run("test ListIRODSDataObjectAccess", testListIRODSDataObjectAccess)
	t.Run("test CreateDeleteIRODSCollection", testCreateDeleteIRODSCollection)
	t.Run("test CreateMoveDeleteIRODSCollection", testCreateMoveDeleteIRODSCollection)
	t.Run("test CreateDeleteIRODSDataObject", testCreateDeleteIRODSDataObject)
	t.Run("test ReadWriteIRODSDataObject", testReadWriteIRODSDataObject)
	t.Run("test ListIRODSGroupUsers", testListIRODSGroupUsers)
	t.Run("test SearchDataObjectsByMeta", testSearchDataObjectsByMeta)
	t.Run("test SearchDataObjectsByMetaWildcard", testSearchDataObjectsByMetaWildcard)

	shutdown()
}

func testGetIRODSCollection(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)

	assert.Equal(t, homedir, collection.Path)
	assert.NotEmpty(t, collection.ID)
}

func testListIRODSCollections(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collections, err := fs.ListSubCollections(conn, homedir)
	assert.NoError(t, err)

	collectionPaths := []string{}

	for _, collection := range collections {
		collectionPaths = append(collectionPaths, collection.Path)
		assert.NotEmpty(t, collection.ID)
	}

	assert.Equal(t, len(collections), len(GetTestDirs()))
	assert.ElementsMatch(t, collectionPaths, GetTestDirs())
}

func testListIRODSCollectionMeta(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	metas, err := fs.ListCollectionMeta(conn, homedir)
	assert.NoError(t, err)

	for _, meta := range metas {
		assert.NotEmpty(t, meta.AVUID)
	}
}

func testListIRODSCollectionAccess(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	accesses, err := fs.ListCollectionAccess(conn, homedir)
	assert.NoError(t, err)

	for _, access := range accesses {
		assert.NotEmpty(t, access.Path)
		assert.Equal(t, homedir, access.Path)
	}
}

func testListIRODSDataObjects(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)
	assert.NotEmpty(t, collection.ID)

	dataobjects, err := fs.ListDataObjects(conn, collection)
	assert.NoError(t, err)

	dataobjectPaths := []string{}

	for _, dataobject := range dataobjects {
		dataobjectPaths = append(dataobjectPaths, dataobject.Path)
		assert.NotEmpty(t, dataobject.ID)

		for _, replica := range dataobject.Replicas {
			assert.NotEmpty(t, replica.Path)
		}
	}

	assert.Equal(t, len(dataobjects), len(GetTestFiles()))
	assert.ElementsMatch(t, dataobjectPaths, GetTestFiles())
}

func testListIRODSDataObjectsMasterReplica(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)
	assert.NotEmpty(t, collection.ID)

	dataobjects, err := fs.ListDataObjectsMasterReplica(conn, collection)
	assert.NoError(t, err)

	dataobjectPaths := []string{}

	for _, dataobject := range dataobjects {
		dataobjectPaths = append(dataobjectPaths, dataobject.Path)
		assert.NotEmpty(t, dataobject.ID)

		assert.Equal(t, 1, len(dataobject.Replicas))
		for _, replica := range dataobject.Replicas {
			assert.NotEmpty(t, replica.Path)
			assert.NotEmpty(t, replica.ResourceName)
		}
	}

	assert.Equal(t, len(dataobjects), len(GetTestFiles()))
	assert.ElementsMatch(t, dataobjectPaths, GetTestFiles())
}

func testGetIRODSDataObject(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)

	assert.Equal(t, homedir, collection.Path)
	assert.NotEmpty(t, collection.ID)

	for _, filepath := range GetTestFiles() {
		filename := path.Base(filepath)
		dirpath := path.Dir(filepath)
		assert.Equal(t, dirpath, homedir)

		dataobject, err := fs.GetDataObject(conn, collection, filename)
		assert.NoError(t, err)

		assert.NotEmpty(t, dataobject.ID)

		for _, replica := range dataobject.Replicas {
			assert.NotEmpty(t, replica.Path)
		}
	}
}

func testGetIRODSDataObjectMasterReplica(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)

	assert.Equal(t, homedir, collection.Path)
	assert.NotEmpty(t, collection.ID)

	for _, filepath := range GetTestFiles() {
		filename := path.Base(filepath)
		dirpath := path.Dir(filepath)
		assert.Equal(t, dirpath, homedir)

		dataobject, err := fs.GetDataObjectMasterReplica(conn, collection, filename)
		assert.NoError(t, err)

		assert.NotEmpty(t, dataobject.ID)

		assert.Equal(t, 1, len(dataobject.Replicas))
		for _, replica := range dataobject.Replicas {
			assert.NotEmpty(t, replica.Path)
		}
	}
}

func testListIRODSDataObjectMeta(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)
	assert.NotEmpty(t, collection.ID)

	for _, filepath := range GetTestFiles() {
		filename := path.Base(filepath)
		dirpath := path.Dir(filepath)
		assert.Equal(t, dirpath, homedir)

		metas, err := fs.ListDataObjectMeta(conn, collection, filename)
		assert.NoError(t, err)

		for _, meta := range metas {
			assert.NotEmpty(t, meta.AVUID)
		}
	}
}

func testListIRODSDataObjectAccess(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)
	assert.NotEmpty(t, collection.ID)

	for _, filepath := range GetTestFiles() {
		filename := path.Base(filepath)
		dirpath := path.Dir(filepath)
		assert.Equal(t, dirpath, homedir)

		accesses, err := fs.ListDataObjectAccess(conn, collection, filename)
		assert.NoError(t, err)

		for _, access := range accesses {
			assert.NotEmpty(t, access.Path)
			assert.Equal(t, filepath, access.Path)
		}
	}
}

func testCreateDeleteIRODSCollection(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	// create
	newCollectionPath := homedir + "/test123"

	err = fs.CreateCollection(conn, newCollectionPath, true)
	assert.NoError(t, err)

	collection, err := fs.GetCollection(conn, newCollectionPath)
	assert.NoError(t, err)

	assert.Equal(t, newCollectionPath, collection.Path)
	assert.NotEmpty(t, collection.ID)

	// delete
	err = fs.DeleteCollection(conn, newCollectionPath, true, false)
	assert.NoError(t, err)

	_, err = fs.GetCollection(conn, newCollectionPath)
	deleted := false
	if err != nil {
		if types.IsFileNotFoundError(err) {
			// Okay!
			deleted = true
		}
	}

	assert.True(t, deleted)
}

func testCreateMoveDeleteIRODSCollection(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	// create
	newCollectionPath := homedir + "/test123"

	err = fs.CreateCollection(conn, newCollectionPath, true)
	assert.NoError(t, err)

	collection, err := fs.GetCollection(conn, newCollectionPath)
	assert.NoError(t, err)

	assert.Equal(t, newCollectionPath, collection.Path)
	assert.NotEmpty(t, collection.ID)

	// move
	new2CollectionPath := newCollectionPath + "_new"
	err = fs.MoveCollection(conn, newCollectionPath, new2CollectionPath)
	assert.NoError(t, err)

	new2Collection, err := fs.GetCollection(conn, new2CollectionPath)
	assert.NoError(t, err)

	assert.Equal(t, new2CollectionPath, new2Collection.Path)
	assert.NotEmpty(t, new2Collection.ID)

	// delete
	err = fs.DeleteCollection(conn, new2CollectionPath, true, false)
	assert.NoError(t, err)

	_, err = fs.GetCollection(conn, new2CollectionPath)
	deleted := false
	if err != nil {
		if types.IsFileNotFoundError(err) {
			// Okay!
			deleted = true
		}
	}

	assert.True(t, deleted)
}

func testCreateDeleteIRODSDataObject(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	// create
	newDataObjectFilename := "testobj123"
	newDataObjectPath := homedir + "/" + newDataObjectFilename
	handle, err := fs.CreateDataObject(conn, newDataObjectPath, "", true)
	assert.NoError(t, err)

	err = fs.CloseDataObject(conn, handle)
	assert.NoError(t, err)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)

	obj, err := fs.GetDataObject(conn, collection, newDataObjectFilename)
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.ID)

	// delete
	err = fs.DeleteDataObject(conn, newDataObjectPath, true)
	assert.NoError(t, err)

	_, err = fs.GetDataObject(conn, collection, newDataObjectFilename)
	deleted := false
	if err != nil {
		if types.IsFileNotFoundError(err) {
			// Okay!
			deleted = true
		}
	}

	assert.True(t, deleted)
}

func testReadWriteIRODSDataObject(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	homedir := fmt.Sprintf("/%s/home/%s", account.ClientZone, account.ClientUser)

	// create
	newDataObjectFilename := "testobjwrite123"
	newDataObjectPath := homedir + "/" + newDataObjectFilename

	handle, err := fs.CreateDataObject(conn, newDataObjectPath, "", true)
	assert.NoError(t, err)

	data := "Hello World"
	err = fs.WriteDataObject(conn, handle, []byte(data))
	assert.NoError(t, err)

	err = fs.CloseDataObject(conn, handle)
	assert.NoError(t, err)

	collection, err := fs.GetCollection(conn, homedir)
	assert.NoError(t, err)

	obj, err := fs.GetDataObject(conn, collection, newDataObjectFilename)
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.ID)

	// read
	handle, _, err = fs.OpenDataObject(conn, newDataObjectPath, "", "r")
	assert.NoError(t, err)

	datarecv, err := fs.ReadDataObject(conn, handle, len(data))
	assert.NoError(t, err)

	assert.Equal(t, data, string(datarecv))

	err = fs.CloseDataObject(conn, handle)
	assert.NoError(t, err)

	// delete
	err = fs.DeleteDataObject(conn, newDataObjectPath, true)
	assert.NoError(t, err)
}

func testListIRODSGroupUsers(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	users, err := fs.ListGroupUsers(conn, "rodsadmin")
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(users), 1)

	adminUsers := []string{}
	for _, user := range users {
		assert.NotEmpty(t, user.ID)

		adminUsers = append(adminUsers, user.Name)
	}

	assert.Contains(t, adminUsers, "rods")
}

func testSearchDataObjectsByMeta(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	for _, testFilePath := range GetTestFiles() {
		sha1sum := sha1.New()
		_, err = sha1sum.Write([]byte(testFilePath))
		assert.NoError(t, err)

		hashBytes := sha1sum.Sum(nil)
		hashString := hex.EncodeToString(hashBytes)

		dataobjects, err := fs.SearchDataObjectsByMeta(conn, "hash", hashString)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(dataobjects))
		assert.Equal(t, testFilePath, dataobjects[0].Path)
	}
}

func testSearchDataObjectsByMetaWildcard(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	assert.NoError(t, err)
	defer conn.Disconnect()

	// this takes a long time to perform
	for _, testFilePath := range GetTestFiles() {
		sha1sum := sha1.New()
		_, err = sha1sum.Write([]byte(testFilePath))
		assert.NoError(t, err)

		hashBytes := sha1sum.Sum(nil)
		hashString := hex.EncodeToString(hashBytes)

		dataobjects, err := fs.SearchDataObjectsByMetaWildcard(conn, "hash", hashString+"%")
		assert.NoError(t, err)

		assert.Equal(t, 1, len(dataobjects))
		assert.Equal(t, testFilePath, dataobjects[0].Path)
	}

	dataobjects, err := fs.SearchDataObjectsByMetaWildcard(conn, "tag", "test%")
	assert.NoError(t, err)

	assert.Equal(t, len(GetTestFiles()), len(dataobjects))
}
