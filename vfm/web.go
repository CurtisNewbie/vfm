package vfm

import (
	"context"
	"strings"

	"github.com/curtisnewbie/goauth/client/goauth-client-go/gclient"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

func listGrantedFolderAccess(c *gin.Context, ec common.ExecContext, req ListGrantedFolderAccessReq) (any, error) {
	// TODO
	return nil, nil
}

func removeGrantedFolderAccess(c *gin.Context, ec common.ExecContext, req RemoveGrantedFolderAccessReq) (any, error) {
	// TODO
	return nil, nil
}

func shareVfolder(c *gin.Context, ec common.ExecContext, req ShareVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func removeFileFromVfolder(c *gin.Context, ec common.ExecContext, req RemoveFileFromVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func addFileToVfolder(c *gin.Context, ec common.ExecContext, req AddFileToVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func createVfolder(c *gin.Context, ec common.ExecContext, req CreateVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func listVfolders(c *gin.Context, ec common.ExecContext, req ListVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func listVfolderBrief(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func untagFile(c *gin.Context, ec common.ExecContext, req UntagFileReq) (any, error) {
	// TODO
	return nil, nil
}

func tagFile(c *gin.Context, ec common.ExecContext, req TagFileReq) (any, error) {
	// TODO
	return nil, nil
}

func listFileTagsEp(c *gin.Context, ec common.ExecContext, req ListFileTagReq) (any, error) {
	return ListFileTags(ec, req)
}

func listTagsEp(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func updateFile(c *gin.Context, ec common.ExecContext, req UpdateFileReq) (any, error) {
	// TODO
	return nil, nil
}

func uploadPreflightCheckEp(c *gin.Context, ec common.ExecContext) (any, error) {
	filename := strings.TrimSpace(c.Query("fileName"))
	parentFileKey := strings.TrimSpace(c.Query("parentFileKey"))
	ec.Log.Debugf("uploadPreflightCheck, filename: %v, parentFileKey: %v", filename, parentFileKey)
	return FileExists(ec, filename, parentFileKey)
}

func fetchParentFileInfo(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func moveFileIntoDir(c *gin.Context, ec common.ExecContext, req MoveIntoDirReq) (any, error) {
	// TODO
	return nil, nil
}

func makeDir(c *gin.Context, ec common.ExecContext, req MakeDirReq) (any, error) {
	// TODO
	return nil, nil
}

func grantAccess(c *gin.Context, ec common.ExecContext, req GrantAccessReq) (any, error) {
	// TODO
	return nil, nil
}

func listGrantedAccess(c *gin.Context, ec common.ExecContext, req ListGrantedAcessReq) (any, error) {
	// TODO
	return nil, nil
}

func removeGrantedAccess(c *gin.Context, ec common.ExecContext, req RemoveGrantedAccessReq) (any, error) {
	// TODO
	return nil, nil
}

func listDirs(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func listFiles(c *gin.Context, ec common.ExecContext, req ListFileReq) (any, error) {
	ec.Log.Debugf("ListFiles, %+v", req)
	return ListFiles(ec, req)
}

func deleteFile(c *gin.Context, ec common.ExecContext, req DeleteFileReq) (any, error) {
	// TODO
	return nil, nil
}

func PrepareServer() {
	if gclient.IsEnabled() {
		server.OnServerBootstrapped(func() {
			c := common.EmptyExecContext()
			if e := gclient.AddResource(context.Background(), gclient.AddResourceReq{Name: MANAGE_FILE_NAME, Code: MANAGE_FILE_CODE}); e != nil {
				c.Log.Errorf("Failed to create goauth resource, %v", e)
			}

			if e := gclient.AddResource(context.Background(), gclient.AddResourceReq{Name: ADMIN_FS_NAME, Code: ADMIN_FS_CODE}); e != nil {
				c.Log.Errorf("Failed to create goauth resource, %v", e)
			}
		})

		gclient.ReportPathsOnBootstrapped()
	}

	server.Get("/open/api/file/upload/duplication/preflight", uploadPreflightCheckEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User - preflight check for duplicate file uploads", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/parent", fetchParentFileInfo,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User fetch parent file info", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/move-to-dir", moveFileIntoDir,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/make-dir", makeDir,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User make directory", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/grant-access", grantAccess,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User grant file access", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/list-granted-access", listGrantedAccess,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list granted file access", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/remove-granted-access", removeGrantedAccess,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User remove granted file access", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/dir/list", listDirs,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list directories", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/list", listFiles,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list files", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/delete", deleteFile,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User delete file", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/info/update", updateFile,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User update file", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/tag/list/all", listTagsEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list all file tags", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/tag/list-for-file", listFileTagsEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list tags of file", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/tag", tagFile,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User tag file", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/untag", untagFile,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User untag file", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/vfolder/brief/owned", listVfolderBrief,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list virtual folder briefs", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/list", listVfolders,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list virtual folders", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/create", createVfolder,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User create virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/file/add", addFileToVfolder,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User add file to virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/file/remove", removeFileFromVfolder,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User remove file from virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/share", shareVfolder,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "Share access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/access/remove", removeGrantedFolderAccess,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "Remove granted access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/granted/list", listGrantedFolderAccess,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "List granted access to virtual folder", Code: MANAGE_FILE_CODE}))
}
