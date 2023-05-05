package vfm

import (
	"context"

	"github.com/curtisnewbie/goauth/client/goauth-client-go/gclient"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

type ExportAsZipReq struct {
	FileIds []string `json:"fileIds"`
}

type MoveIntoDirReq struct {
	Uuid           string `json:"uuid" validation:"notEmpty"`
	ParentFileUuid string `json:"parentFileUuid"`
}

type MakeDirReq struct {
	ParentFile string `json:"parentFile"`                 // Key of parent file
	Name       string `json:"name" validation:"notEmpty"` // name of the directory
	UserGroup  string `json:"userGroup"`                  // User Group
}

func uploadPreflightCheck(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func fetchParentFileInfo(c *gin.Context, ec common.ExecContext) (any, error) {
	// TODO
	return nil, nil
}

func exportAsZip(c *gin.Context, ec common.ExecContext, req ExportAsZipReq) (any, error) {
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

func PrepareServer() {
	if gclient.IsEnabled() {
		server.OnServerBootstrapped(func() {
			c := common.EmptyExecContext()
			if e := gclient.AddResource(context.Background(), gclient.AddResourceReq{Name: MANAGE_FILE_NAME, Code: MANAGE_FILE_CODE}); e != nil {
				c.Log.Errorf("Failed to create goauth resource", e)
			}

			if e := gclient.AddResource(context.Background(), gclient.AddResourceReq{Name: ADMIN_FS_NAME, Code: ADMIN_FS_CODE}); e != nil {
				c.Log.Errorf("Failed to create goauth resource", e)
			}
		})

		gclient.ReportPathsOnBootstrapped()
	}

	server.Get("/open/api/file/upload/duplication/preflight", uploadPreflightCheck,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User - preflight check for duplicate file uploads", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/parent", fetchParentFileInfo,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User fetch parent file info", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/export-as-zip", exportAsZip,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User export files", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/move-to-dir", moveFileIntoDir,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}))

	server.PostJ("/open/api/file/make-dir", makeDir,
		gclient.PathDocExtra(gclient.PathDoc{Desc: "User make directory", Code: MANAGE_FILE_CODE}))



}
