package panel

import (
	"os"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

type PublicGalleryApi struct{}

var publicGalleryAgreeExts = []string{
	".png", ".jpg", ".gif", ".jpeg", ".webp", ".svg", ".ico", ".avif", ".heic",
}

// GetImagesList 公共图库列表
// POST /panel/publicGallery/getImagesList {"page":1,"pageSize":20}
func (a *PublicGalleryApi) GetImagesList(c *gin.Context) {
	req := struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 20
	}

	db := models.Db.Model(&models.File{}).Where("is_public_gallery = ?", true)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	var files []models.File
	if err := db.Order("created_at desc").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Find(&files).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.SuccessListData(c, files, count)
}

// UploadImg 上传图片到公共图库
// POST /api/panel/publicGallery/uploadImg  multipart: imgfile, type(可选)
func (a *PublicGalleryApi) UploadImg(c *gin.Context) {
	fileInfo, err := base.UploadFileAndSave(c, publicGalleryAgreeExts, 0)
	if err != nil {
		switch err {
		case base.ErrUploadExtensionNameNotAllowed:
			apiReturn.ErrorByCode(c, apiReturn.ErrCodeUnsupportedFileFormat)
		default:
			global.Logger.Error(err.Error())
			apiReturn.ErrorByCode(c, apiReturn.ErrCodeUploadFailed)
		}
		return
	}

	// 标记为图库图片并公开
	fileInfo.Type = models.FileTypeGalleryImg
	fileInfo.IsPublicGallery = true
	if err := models.Db.Model(&models.File{}).Where("id = ?", fileInfo.ID).
		Updates(map[string]interface{}{"type": fileInfo.Type, "is_public_gallery": true}).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.SuccessData(c, fileInfo)
}

// Deletes 从公共图库删除
// POST /panel/publicGallery/deletes {"ids":[1,2]}
func (a *PublicGalleryApi) Deletes(c *gin.Context) {
	req := struct {
		IDs []uint `json:"ids"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if len(req.IDs) == 0 {
		apiReturn.ErrorParamFomat(c, "ids is empty")
		return
	}

	// 先查出待删除记录，删除数据库行后同步删除物理文件，
	// 否则文件会永久残留在上传目录且仍可通过 /uploads 静态路由公开访问。
	var files []models.File
	if err := models.Db.Where("id in ? and is_public_gallery = ?", req.IDs, true).
		Find(&files).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	if err := models.Db.Where("id in ? and is_public_gallery = ?", req.IDs, true).
		Delete(&models.File{}).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	// 物理文件删除失败只记录日志，不影响数据库记录已删除的结果
	sourcePath := global.Config.GetValueString("base", "source_path")
	for _, f := range files {
		src := strings.TrimLeft(f.Src, "/")
		if src == "" {
			continue
		}
		fullPath := strings.TrimRight(sourcePath, "/") + "/" + src
		if exists, _ := cmn.PathExists(fullPath); exists {
			if err := os.Remove(fullPath); err != nil {
				global.Logger.Errorln("delete public gallery file failed:", fullPath, err)
			}
		}
	}
	apiReturn.Success(c)
}

// UpdateFileType 修改公共图库文件的类型（管理员）
// POST /panel/publicGallery/updateFileType {"id":1,"type":1}
func (a *PublicGalleryApi) UpdateFileType(c *gin.Context) {
	req := struct {
		ID   uint `json:"id"`
		Type int  `json:"type"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	mFile := models.File{}
	if err := models.Db.Where("id = ?", req.ID).First(&mFile).Error; err != nil {
		apiReturn.ErrorDataNotFound(c)
		return
	}
	if err := models.Db.Model(&mFile).Update("type", req.Type).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.SuccessData(c, mFile)
}
