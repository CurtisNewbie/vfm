package vfm

import (
	"context"

	"github.com/curtisnewbie/goauth/client/goauth-client-go/gclient"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

func listGrantedFolderAccessEp(c *gin.Context, ec common.ExecContext, req ListGrantedFolderAccessReq) (any, error) {
	// TODO
	return nil, nil
}

func removeGrantedFolderAccessEp(c *gin.Context, ec common.ExecContext, req RemoveGrantedFolderAccessReq) (any, error) {
	// TODO
	return nil, nil
}

func shareVFolderEp(c *gin.Context, ec common.ExecContext, req ShareVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func removeFileFromVfolderEp(c *gin.Context, ec common.ExecContext, req RemoveFileFromVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func addFileToVfolderEp(c *gin.Context, ec common.ExecContext, req AddFileToVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func createVFolderEp(c *gin.Context, ec common.ExecContext, req CreateVfolderReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return CreateVFolder(ec, req)
}

func listVfoldersEp(c *gin.Context, ec common.ExecContext, req ListVfolderReq) (any, error) {
	// TODO
	return nil, nil
}

func listVfolderBriefEp(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func untagFileEp(c *gin.Context, ec common.ExecContext, req UntagFileReq) (any, error) {
	// TODO
	return nil, nil
}

func tagFileEp(c *gin.Context, ec common.ExecContext, req TagFileReq) (any, error) {
	// TODO
	return nil, nil
}

func listFileTagsEp(c *gin.Context, ec common.ExecContext, req ListFileTagReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListFileTags(ec, req)
}

func listTagsEp(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func updateFileEp(c *gin.Context, ec common.ExecContext, req UpdateFileReq) (any, error) {
	// TODO
	return nil, nil
}

func uploadPreflightCheckEp(c *gin.Context, ec common.ExecContext) (any, error) {
	filename := c.Query("fileName")
	parentFileKey := c.Query("parentFileKey")
	ec.Log.Debugf("filename: %v, parentFileKey: %v", filename, parentFileKey)
	return FileExists(ec, filename, parentFileKey)
}

func fetchParentFileInfoEp(c *gin.Context, ec common.ExecContext) (any, error) {
	fk := c.Query("fileKey")
	if fk == "" {
		return nil, common.NewWebErr("fileKey is required")
	}
	ec.Log.Debugf("fileKey: %v", fk)
	pf, e := FindParentFile(ec, fk)
	if e != nil {
		return nil, e
	}
	if pf.Zero {
		return nil, nil
	}
	return pf, nil
}

func moveFileIntoDirEp(c *gin.Context, ec common.ExecContext, req MoveIntoDirReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, MoveFileToDir(ec, req)
}

func makeDirEp(c *gin.Context, ec common.ExecContext, req MakeDirReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return MakeDir(ec, req)
}

func grantAccessEp(c *gin.Context, ec common.ExecContext, req GrantAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	// TODO
	return nil, nil
}

func listGrantedAccessEp(c *gin.Context, ec common.ExecContext, req ListGrantedAcessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	// TODO
	return nil, nil
}

func removeGrantedAccessEp(c *gin.Context, ec common.ExecContext, req RemoveGrantedAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	// TODO
	return nil, nil
}

func listDirsEp(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func listFilesEp(c *gin.Context, ec common.ExecContext, req ListFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListFiles(ec, req)
}

func deleteFileEp(c *gin.Context, ec common.ExecContext, req DeleteFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
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

	server.Get("/open/api/file/parent", fetchParentFileInfoEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User fetch parent file info", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/move-to-dir", moveFileIntoDirEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/make-dir", makeDirEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User make directory", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/grant-access", grantAccessEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User grant file access", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/list-granted-access", listGrantedAccessEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list granted file access", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/remove-granted-access", removeGrantedAccessEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User remove granted file access", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/dir/list", listDirsEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list directories", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/list", listFilesEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list files", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/delete", deleteFileEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User delete file", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/info/update", updateFileEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User update file", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/tag/list/all", listTagsEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list all file tags", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/tag/list-for-file", listFileTagsEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list tags of file", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/tag", tagFileEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User tag file", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/untag", untagFileEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User untag file", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/vfolder/brief/owned", listVfolderBriefEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list virtual folder briefs", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/list", listVfoldersEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User list virtual folders", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/create", createVFolderEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User create virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/file/add", addFileToVfolderEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User add file to virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/file/remove", removeFileFromVfolderEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User remove file from virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/share", shareVFolderEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "Share access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/access/remove", removeGrantedFolderAccessEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "Remove granted access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/vfolder/granted/list", listGrantedFolderAccessEp,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "List granted access to virtual folder", Code: MANAGE_FILE_CODE}))
}
