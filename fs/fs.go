package fs

import (
	"fmt"
	"os"
	"time"

	irods_fs "github.com/cyverse/go-irodsclient/irods/fs"
	"github.com/cyverse/go-irodsclient/irods/session"
	"github.com/cyverse/go-irodsclient/irods/types"
	"github.com/cyverse/go-irodsclient/irods/util"
)

// FileSystem provides a file-system like interface
type FileSystem struct {
	Account *types.IRODSAccount
	Config  *FileSystemConfig
	Session *session.IRODSSession
	Cache   *FileSystemCache
}

// NewFileSystem creates a new FileSystem
func NewFileSystem(account *types.IRODSAccount, config *FileSystemConfig) *FileSystem {
	sessConfig := session.NewIRODSSessionConfig(config.ApplicationName, config.OperationTimeout, config.ConnectionIdleTimeout, config.ConnectionMax)
	sess := session.NewIRODSSession(account, sessConfig)
	cache := NewFileSystemCache(config.CacheTimeout, config.CacheCleanupTime)

	return &FileSystem{
		Account: account,
		Config:  config,
		Session: sess,
		Cache:   cache,
	}
}

// NewFileSystemWithDefault ...
func NewFileSystemWithDefault(account *types.IRODSAccount, applicationName string) *FileSystem {
	config := NewFileSystemConfigWithDefault(applicationName)
	sessConfig := session.NewIRODSSessionConfig(config.ApplicationName, config.OperationTimeout, config.ConnectionIdleTimeout, config.ConnectionMax)
	sess := session.NewIRODSSession(account, sessConfig)
	cache := NewFileSystemCache(config.CacheTimeout, config.CacheCleanupTime)

	return &FileSystem{
		Account: account,
		Config:  config,
		Session: sess,
		Cache:   cache,
	}
}

// Release ...
func (fs *FileSystem) Release() {
	fs.Session.Release()
}

// Stat returns file status
func (fs *FileSystem) Stat(path string) (*FSEntry, error) {
	dirStat, err := fs.StatDir(path)
	if err != nil {
		if !types.IsFileNotFoundError(err) {
			return nil, err
		}
	} else {
		return dirStat, nil
	}

	fileStat, err := fs.StatFile(path)
	if err != nil {
		if !types.IsFileNotFoundError(err) {
			return nil, err
		}
	} else {
		return fileStat, nil
	}

	// not a collection, not a data object
	return nil, types.NewFileNotFoundError("Could not find a data object or a directory")
}

// StatDir returns status of a directory
func (fs *FileSystem) StatDir(path string) (*FSEntry, error) {
	irodsPath := util.GetCorrectIRODSPath(path)

	return fs.getCollection(irodsPath)
}

// StatFile returns status of a file
func (fs *FileSystem) StatFile(path string) (*FSEntry, error) {
	irodsPath := util.GetCorrectIRODSPath(path)

	return fs.getDataObject(irodsPath)
}

// Exists checks file/directory existance
func (fs *FileSystem) Exists(path string) bool {
	entry, err := fs.Stat(path)
	if err != nil {
		return false
	}
	if entry.ID > 0 {
		return true
	}
	return false
}

// ExistsDir checks directory existance
func (fs *FileSystem) ExistsDir(path string) bool {
	entry, err := fs.StatDir(path)
	if err != nil {
		return false
	}
	if entry.ID > 0 {
		return true
	}
	return false
}

// ExistsFile checks file existance
func (fs *FileSystem) ExistsFile(path string) bool {
	entry, err := fs.StatFile(path)
	if err != nil {
		return false
	}
	if entry.ID > 0 {
		return true
	}
	return false
}

// List lists all file system entries under the given path
func (fs *FileSystem) List(path string) ([]*FSEntry, error) {
	irodsPath := util.GetCorrectIRODSPath(path)

	collection, err := fs.getCollection(irodsPath)
	if err != nil {
		return nil, err
	}

	return fs.listEntries(collection.Internal.(*types.IRODSCollection))
}

// RemoveDir deletes a directory
func (fs *FileSystem) RemoveDir(path string, recurse bool, force bool) error {
	irodsPath := util.GetCorrectIRODSPath(path)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.DeleteCollection(conn, irodsPath, recurse, force)
	if err != nil {
		return err
	}

	fs.removeCachePath(irodsPath)
	return nil
}

// RemoveFile deletes a file
func (fs *FileSystem) RemoveFile(path string, force bool) error {
	irodsPath := util.GetCorrectIRODSPath(path)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.DeleteDataObject(conn, irodsPath, force)
	if err != nil {
		return err
	}

	fs.removeCachePath(irodsPath)
	return nil
}

// RenameDir renames a dir
func (fs *FileSystem) RenameDir(srcPath string, destPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(srcPath)
	irodsDestPath := util.GetCorrectIRODSPath(destPath)

	destDirPath := irodsDestPath
	if fs.ExistsDir(irodsDestPath) {
		// make full file name for dest
		srcFileName := util.GetIRODSPathFileName(irodsSrcPath)
		destDirPath = util.MakeIRODSPath(irodsDestPath, srcFileName)
	}

	return fs.RenameDirToDir(irodsSrcPath, destDirPath)
}

// RenameDirToDir renames a dir
func (fs *FileSystem) RenameDirToDir(srcPath string, destPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(srcPath)
	irodsDestPath := util.GetCorrectIRODSPath(destPath)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.MoveCollection(conn, irodsSrcPath, irodsDestPath)
	if err != nil {
		return err
	}

	if util.GetIRODSPathDirname(irodsSrcPath) == util.GetIRODSPathDirname(irodsDestPath) {
		// from the same dir
		fs.invalidateCachePath(util.GetIRODSPathDirname(irodsSrcPath))

	} else {
		fs.removeCachePath(irodsSrcPath)
		fs.invalidateCachePath(util.GetIRODSPathDirname(irodsDestPath))
	}

	return nil
}

// RenameFile renames a file
func (fs *FileSystem) RenameFile(srcPath string, destPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(srcPath)
	irodsDestPath := util.GetCorrectIRODSPath(destPath)

	destFilePath := irodsDestPath
	if fs.ExistsDir(irodsDestPath) {
		// make full file name for dest
		srcFileName := util.GetIRODSPathFileName(irodsSrcPath)
		destFilePath = util.MakeIRODSPath(irodsDestPath, srcFileName)
	}

	return fs.RenameFileToFile(irodsSrcPath, destFilePath)
}

// RenameFileToFile renames a file
func (fs *FileSystem) RenameFileToFile(srcPath string, destPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(srcPath)
	irodsDestPath := util.GetCorrectIRODSPath(destPath)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.MoveDataObject(conn, irodsSrcPath, irodsDestPath)
	if err != nil {
		return err
	}

	if util.GetIRODSPathDirname(irodsSrcPath) == util.GetIRODSPathDirname(irodsDestPath) {
		// from the same dir
		fs.invalidateCachePath(util.GetIRODSPathDirname(irodsSrcPath))
	} else {
		fs.removeCachePath(irodsSrcPath)
		fs.invalidateCachePath(util.GetIRODSPathDirname(irodsDestPath))
	}

	return nil
}

// MakeDir creates a directory
func (fs *FileSystem) MakeDir(path string, recurse bool) error {
	irodsPath := util.GetCorrectIRODSPath(path)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.CreateCollection(conn, irodsPath, recurse)
	if err != nil {
		return err
	}

	fs.invalidateCachePath(util.GetIRODSPathDirname(irodsPath))

	return nil
}

// CopyFile copies a file
func (fs *FileSystem) CopyFile(srcPath string, destPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(srcPath)
	irodsDestPath := util.GetCorrectIRODSPath(destPath)

	destFilePath := irodsDestPath
	if fs.ExistsDir(irodsDestPath) {
		// make full file name for dest
		srcFileName := util.GetIRODSPathFileName(irodsSrcPath)
		destFilePath = util.MakeIRODSPath(irodsDestPath, srcFileName)
	}

	return fs.CopyFileToFile(irodsSrcPath, destFilePath)
}

// CopyFileToFile copies a file
func (fs *FileSystem) CopyFileToFile(srcPath string, destPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(srcPath)
	irodsDestPath := util.GetCorrectIRODSPath(destPath)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.CopyDataObject(conn, irodsSrcPath, irodsDestPath)
	if err != nil {
		return err
	}

	fs.invalidateCachePath(util.GetIRODSPathDirname(irodsDestPath))

	return nil
}

// TruncateFile truncates a file
func (fs *FileSystem) TruncateFile(path string, size int64) error {
	irodsPath := util.GetCorrectIRODSPath(path)

	if size < 0 {
		size = 0
	}

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.TruncateDataObject(conn, irodsPath, size)
	if err != nil {
		return err
	}

	fs.invalidateCachePath(util.GetIRODSPathDirname(irodsPath))

	return nil
}

// ReplicateFile replicates a file
func (fs *FileSystem) ReplicateFile(path string, resource string, update bool) error {
	irodsPath := util.GetCorrectIRODSPath(path)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	return irods_fs.ReplicateDataObject(conn, irodsPath, resource, update)
}

// DownloadFile downloads a file to local
func (fs *FileSystem) DownloadFile(irodsPath string, localPath string) error {
	irodsSrcPath := util.GetCorrectIRODSPath(irodsPath)
	localDestPath := util.GetCorrectIRODSPath(localPath)

	localFilePath := localDestPath
	stat, err := os.Stat(localDestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// file not exists, it's a file
			localFilePath = localDestPath
		} else {
			return err
		}
	} else {
		if stat.IsDir() {
			irodsFileName := util.GetIRODSPathFileName(irodsSrcPath)
			localFilePath = util.MakeIRODSPath(localDestPath, irodsFileName)
		} else {
			return fmt.Errorf("File %s already exists", localDestPath)
		}
	}

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	return irods_fs.DownloadDataObject(conn, irodsSrcPath, localFilePath)
}

// UploadFile uploads a local file to irods
func (fs *FileSystem) UploadFile(localPath string, irodsPath string, resource string, replicate bool) error {
	localSrcPath := util.GetCorrectIRODSPath(localPath)
	irodsDestPath := util.GetCorrectIRODSPath(irodsPath)

	irodsFilePath := irodsDestPath

	stat, err := os.Stat(localSrcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// file not exists
			return types.NewFileNotFoundError("Could not find the local file")
		}
		return err
	}

	if stat.IsDir() {
		return types.NewFileNotFoundError("The local file is a directory")
	}

	entry, err := fs.Stat(irodsDestPath)
	if err != nil {
		if !types.IsFileNotFoundError(err) {
			return err
		}
	} else {
		switch entry.Type {
		case FSFileEntry:
			// do nothing
		case FSDirectoryEntry:
			localFileName := util.GetIRODSPathFileName(localSrcPath)
			irodsFilePath = util.MakeIRODSPath(irodsDestPath, localFileName)
		default:
			return fmt.Errorf("Unknown entry type %s", entry.Type)
		}
	}

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return err
	}
	defer fs.Session.ReturnConnection(conn)

	err = irods_fs.UploadDataObject(conn, localSrcPath, irodsFilePath, resource, replicate)
	if err != nil {
		return err
	}

	fs.invalidateCachePath(util.GetIRODSPathDirname(irodsFilePath))

	return nil
}

// OpenFile opens an existing file for read/write
func (fs *FileSystem) OpenFile(path string, resource string, mode string) (*FileHandle, error) {
	irodsPath := util.GetCorrectIRODSPath(path)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return nil, err
	}

	handle, offset, err := irods_fs.OpenDataObject(conn, irodsPath, resource, mode)
	if err != nil {
		fs.Session.ReturnConnection(conn)
		return nil, err
	}

	var entry *FSEntry = nil
	if types.IsFileOpenFlagOpeningExisting(types.FileOpenMode(mode)) {
		// file may exists
		entryExisting, err := fs.StatFile(irodsPath)
		if err == nil {
			entry = entryExisting
		}
	}

	if entry == nil {
		// create a new
		entry = &FSEntry{
			ID:         0,
			Type:       FSFileEntry,
			Name:       util.GetIRODSPathFileName(irodsPath),
			Path:       irodsPath,
			Owner:      fs.Account.ClientUser,
			Size:       0,
			CreateTime: time.Now(),
			ModifyTime: time.Now(),
			CheckSum:   "",
			Internal:   nil,
		}
	}

	// do not return connection here
	return &FileHandle{
		FileSystem:  fs,
		Connection:  conn,
		IRODSHandle: handle,
		Entry:       entry,
		Offset:      offset,
		OpenMode:    types.FileOpenMode(mode),
	}, nil
}

// CreateFile opens a new file for write
func (fs *FileSystem) CreateFile(path string, resource string) (*FileHandle, error) {
	irodsPath := util.GetCorrectIRODSPath(path)

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return nil, err
	}

	handle, err := irods_fs.CreateDataObject(conn, irodsPath, resource, true)
	if err != nil {
		fs.Session.ReturnConnection(conn)
		return nil, err
	}

	// do not return connection here
	entry := &FSEntry{
		ID:         0,
		Type:       FSFileEntry,
		Name:       util.GetIRODSPathFileName(irodsPath),
		Path:       irodsPath,
		Owner:      fs.Account.ClientUser,
		Size:       0,
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
		CheckSum:   "",
		Internal:   nil,
	}

	return &FileHandle{
		FileSystem:  fs,
		Connection:  conn,
		IRODSHandle: handle,
		Entry:       entry,
		Offset:      0,
		OpenMode:    types.FileOpenModeWriteOnly,
	}, nil
}

// ClearCache ...
func (fs *FileSystem) ClearCache() {
	fs.Cache.ClearEntryCache()
	fs.Cache.ClearDirCache()
}

// InvalidateCache invalidates cache with the given path
func (fs *FileSystem) InvalidateCache(path string) {
	irodsPath := util.GetCorrectIRODSPath(path)

	fs.invalidateCachePath(irodsPath)
}

func (fs *FileSystem) getCollection(path string) (*FSEntry, error) {
	// check cache first
	cachedEntry := fs.Cache.GetEntryCache(path)
	if cachedEntry != nil {
		return cachedEntry, nil
	}

	// otherwise, retrieve it and add it to cache
	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.Session.ReturnConnection(conn)

	collection, err := irods_fs.GetCollection(conn, path)
	if err != nil {
		return nil, err
	}

	if collection.ID > 0 {
		fsEntry := &FSEntry{
			ID:         collection.ID,
			Type:       FSDirectoryEntry,
			Name:       collection.Name,
			Path:       collection.Path,
			Owner:      collection.Owner,
			Size:       0,
			CreateTime: collection.CreateTime,
			ModifyTime: collection.ModifyTime,
			CheckSum:   "",
			Internal:   collection,
		}

		// cache it
		fs.Cache.AddEntryCache(fsEntry)

		return fsEntry, nil
	}

	return nil, types.NewFileNotFoundErrorf("Could not find a directory")
}

func (fs *FileSystem) listEntries(collection *types.IRODSCollection) ([]*FSEntry, error) {
	// check cache first
	cachedEntries := []*FSEntry{}
	useCached := false

	cachedDirEntryPaths := fs.Cache.GetDirCache(collection.Path)
	if cachedDirEntryPaths != nil {
		useCached = true
		for _, cachedDirEntryPath := range cachedDirEntryPaths {
			cachedEntry := fs.Cache.GetEntryCache(cachedDirEntryPath)
			if cachedEntry != nil {
				cachedEntries = append(cachedEntries, cachedEntry)
			} else {
				useCached = false
			}
		}
	}

	if useCached {
		return cachedEntries, nil
	}

	// otherwise, retrieve it and add it to cache
	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.Session.ReturnConnection(conn)

	collections, err := irods_fs.ListSubCollections(conn, collection.Path)
	if err != nil {
		return nil, err
	}

	fsEntries := []*FSEntry{}

	for _, coll := range collections {
		fsEntry := &FSEntry{
			ID:         coll.ID,
			Type:       FSDirectoryEntry,
			Name:       coll.Name,
			Path:       coll.Path,
			Owner:      coll.Owner,
			Size:       0,
			CreateTime: coll.CreateTime,
			ModifyTime: coll.ModifyTime,
			CheckSum:   "",
			Internal:   coll,
		}

		fsEntries = append(fsEntries, fsEntry)

		// cache it
		fs.Cache.AddEntryCache(fsEntry)
	}

	dataobjects, err := irods_fs.ListDataObjectsMasterReplica(conn, collection)
	if err != nil {
		return nil, err
	}

	for _, dataobject := range dataobjects {
		if len(dataobject.Replicas) == 0 {
			continue
		}

		replica := dataobject.Replicas[0]

		fsEntry := &FSEntry{
			ID:         dataobject.ID,
			Type:       FSFileEntry,
			Name:       dataobject.Name,
			Path:       dataobject.Path,
			Owner:      replica.Owner,
			Size:       dataobject.Size,
			CreateTime: replica.CreateTime,
			ModifyTime: replica.ModifyTime,
			CheckSum:   replica.CheckSum,
			Internal:   dataobject,
		}

		fsEntries = append(fsEntries, fsEntry)

		// cache it
		fs.Cache.AddEntryCache(fsEntry)
	}

	// cache dir entries
	dirEntryPaths := []string{}
	for _, fsEntry := range fsEntries {
		dirEntryPaths = append(dirEntryPaths, fsEntry.Path)
	}
	fs.Cache.AddDirCache(collection.Path, dirEntryPaths)

	return fsEntries, nil
}

func (fs *FileSystem) getDataObject(path string) (*FSEntry, error) {
	// check cache first
	cachedEntry := fs.Cache.GetEntryCache(path)
	if cachedEntry != nil {
		return cachedEntry, nil
	}

	// otherwise, retrieve it and add it to cache
	collection, err := fs.getCollection(util.GetIRODSPathDirname(path))
	if err != nil {
		return nil, err
	}

	conn, err := fs.Session.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.Session.ReturnConnection(conn)

	dataobject, err := irods_fs.GetDataObjectMasterReplica(conn, collection.Internal.(*types.IRODSCollection), util.GetIRODSPathFileName(path))
	if err != nil {
		return nil, err
	}

	if dataobject.ID > 0 {
		fsEntry := &FSEntry{
			ID:         dataobject.ID,
			Type:       FSFileEntry,
			Name:       dataobject.Name,
			Path:       dataobject.Path,
			Owner:      dataobject.Replicas[0].Owner,
			Size:       dataobject.Size,
			CreateTime: dataobject.Replicas[0].CreateTime,
			ModifyTime: dataobject.Replicas[0].ModifyTime,
			CheckSum:   dataobject.Replicas[0].CheckSum,
			Internal:   dataobject,
		}

		// cache it
		fs.Cache.AddEntryCache(fsEntry)

		return fsEntry, nil
	}

	return nil, types.NewFileNotFoundErrorf("Could not find a data object")
}

// InvalidateCachePath invalidates cache with the given path
func (fs *FileSystem) invalidateCachePath(path string) {
	fs.Cache.RemoveEntryCache(path)
	fs.Cache.RemoveDirCache(path)
}

func (fs *FileSystem) removeCachePath(path string) {
	// if path is directory, recursively
	entry := fs.Cache.GetEntryCache(path)
	if entry != nil {
		fs.Cache.RemoveEntryCache(path)

		if entry.Type == FSDirectoryEntry {
			dirEntries := fs.Cache.GetDirCache(path)
			if dirEntries != nil {
				for _, dirEntry := range dirEntries {
					// do it recursively
					fs.removeCachePath(dirEntry)
				}
				fs.Cache.RemoveDirCache(path)
			}
		}

		fs.Cache.RemoveDirCache(util.GetIRODSPathDirname(path))
	}
}