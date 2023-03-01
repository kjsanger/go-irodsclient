package fs

import (
	irods_fs "github.com/cyverse/go-irodsclient/irods/fs"
	"github.com/cyverse/go-irodsclient/irods/types"
	"github.com/cyverse/go-irodsclient/irods/util"
)

// SearchByMeta searches all file system entries with given metadata
func (fs *FileSystem) SearchByMeta(metaname string, metavalue string) ([]*Entry, error) {
	return fs.searchEntriesByMeta(metaname, metavalue)
}

// ListMetadata lists metadata for the given path
func (fs *FileSystem) ListMetadata(path string) ([]*types.IRODSMeta, error) {
	// check cache first
	cachedEntry := fs.cache.GetMetadataCache(path)
	if cachedEntry != nil {
		return cachedEntry, nil
	}

	irodsCorrectPath := util.GetCorrectIRODSPath(path)

	// otherwise, retrieve it and add it to cache
	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.metaSession.ReturnConnection(conn)

	var metadataobjects []*types.IRODSMeta

	if fs.ExistsDir(irodsCorrectPath) {
		metadataobjects, err = irods_fs.ListCollectionMeta(conn, irodsCorrectPath)
		if err != nil {
			return nil, err
		}
	} else {
		collectionEntry, err := fs.getCollection(util.GetIRODSPathDirname(path))
		if err != nil {
			return nil, err
		}

		collection := fs.getCollectionFromEntry(collectionEntry)

		metadataobjects, err = irods_fs.ListDataObjectMeta(conn, collection, util.GetIRODSPathFileName(irodsCorrectPath))
		if err != nil {
			return nil, err
		}
	}

	// cache it
	fs.cache.AddMetadataCache(irodsCorrectPath, metadataobjects)

	return metadataobjects, nil
}

// AddMetadata adds a metadata for the path
func (fs *FileSystem) AddMetadata(irodsPath string, attName string, attValue string, attUnits string) error {
	irodsCorrectPath := util.GetCorrectIRODSPath(irodsPath)

	metadata := &types.IRODSMeta{
		Name:  attName,
		Value: attValue,
		Units: attUnits,
	}

	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.metaSession.ReturnConnection(conn)

	if fs.ExistsDir(irodsCorrectPath) {
		err = irods_fs.AddCollectionMeta(conn, irodsCorrectPath, metadata)
		if err != nil {
			return err
		}
	} else {
		err = irods_fs.AddDataObjectMeta(conn, irodsCorrectPath, metadata)
		if err != nil {
			return err
		}
	}

	fs.cache.RemoveMetadataCache(irodsCorrectPath)
	return nil
}

// DeleteMetadata deletes a metadata for the path
func (fs *FileSystem) DeleteMetadata(irodsPath string, attName string, attValue string, attUnits string) error {
	irodsCorrectPath := util.GetCorrectIRODSPath(irodsPath)

	metadata := &types.IRODSMeta{
		Name:  attName,
		Value: attValue,
		Units: attUnits,
	}

	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.metaSession.ReturnConnection(conn)

	if fs.ExistsDir(irodsCorrectPath) {
		err = irods_fs.DeleteCollectionMeta(conn, irodsCorrectPath, metadata)
		if err != nil {
			return err
		}
	} else {
		err = irods_fs.DeleteDataObjectMeta(conn, irodsCorrectPath, metadata)
		if err != nil {
			return err
		}
	}

	fs.cache.RemoveMetadataCache(irodsCorrectPath)
	return nil
}

// AddUserMetadata adds a user metadata
func (fs *FileSystem) AddUserMetadata(user string, avuid int64, attName, attValue, attUnits string) error {
	metadata := &types.IRODSMeta{
		AVUID: avuid,
		Name:  attName,
		Value: attValue,
		Units: attUnits,
	}

	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.metaSession.ReturnConnection(conn)

	err = irods_fs.AddUserMeta(conn, user, metadata)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUserMetadata deletes a user metadata
func (fs *FileSystem) DeleteUserMetadata(user string, avuid int64, attName, attValue, attUnits string) error {
	metadata := &types.IRODSMeta{
		AVUID: avuid,
		Name:  attName,
		Value: attValue,
		Units: attUnits,
	}

	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.metaSession.ReturnConnection(conn)

	err = irods_fs.DeleteUserMeta(conn, user, metadata)
	if err != nil {
		return err
	}

	return nil
}

// ListUserMetadata lists all user metadata
func (fs *FileSystem) ListUserMetadata(user string) ([]*types.IRODSMeta, error) {
	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.metaSession.ReturnConnection(conn)

	metadataobjects, err := irods_fs.ListUserMeta(conn, user)
	if err != nil {
		return nil, err
	}

	return metadataobjects, nil
}

// searchEntriesByMeta searches entries by meta
func (fs *FileSystem) searchEntriesByMeta(metaName string, metaValue string) ([]*Entry, error) {
	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.metaSession.ReturnConnection(conn)

	collections, err := irods_fs.SearchCollectionsByMeta(conn, metaName, metaValue)
	if err != nil {
		return nil, err
	}

	entries := []*Entry{}

	for _, coll := range collections {
		entry := fs.getEntryFromCollection(coll)
		entries = append(entries, entry)

		// cache it
		fs.cache.RemoveNegativeEntryCache(entry.Path)
		fs.cache.AddEntryCache(entry)
	}

	dataobjects, err := irods_fs.SearchDataObjectsMasterReplicaByMeta(conn, metaName, metaValue)
	if err != nil {
		return nil, err
	}

	for _, dataobject := range dataobjects {
		if len(dataobject.Replicas) == 0 {
			continue
		}

		entry := fs.getEntryFromDataObject(dataobject)
		entries = append(entries, entry)

		// cache it
		fs.cache.RemoveNegativeEntryCache(entry.Path)
		fs.cache.AddEntryCache(entry)
	}

	return entries, nil
}
