package biz

import (
	"encoding/json"
	"errors"
	"sun-panel/global"
	"sun-panel/models"
)

type ItemType struct {
}

var (
	ErrItemOnlyNameExisted = errors.New("only name existed")
)

func (p *ItemType) Edit(userId uint, itemIcon models.ItemIcon) (models.ItemIcon, error) {

	// 没有分组自行选择一个排在第一位的分组
	if itemIcon.ItemIconGroupId == 0 {
		defaultGroup := models.ItemIconGroup{}
		global.Db.Order("sort ,created_at").Where("user_id=?", userId).First(&defaultGroup)
		itemIcon.ItemIconGroupId = int(defaultGroup.ID)
	}

	itemIcon.UserId = userId

	// json转字符串
	if j, err := json.Marshal(itemIcon.Icon); err == nil {
		itemIcon.IconJson = string(j)
	}

	if itemIcon.ID != 0 {

		if v, err := p.FindByOnlyName(userId, itemIcon.OnlyName); err == nil && v.ID != itemIcon.ID && itemIcon.OnlyName != "" {
			return itemIcon, ErrItemOnlyNameExisted
		}

		// 修改
		updateField := []string{
			"IconJson",
			"Icon",
			"Title",
			"Url",
			"LanUrl",
			"Description",
			"OpenMethod",
			"GroupId",
			"UserId",
			"ItemIconGroupId",
			"OnlyName",
			"CardType",
			"BackgroundColor",
			"ExpandParam",
		}
		// 修改不更新sort
		// if itemIcon.Sort != 0 {
		// 	updateField = append(updateField, "Sort")
		// }
		// 修复：数据库写入错误必须上抛，否则 openAPI 会误报成功、静默丢数据
		if err := global.Db.Model(&models.ItemIcon{}).
			Select(updateField).
			Where("id=?", itemIcon.ID).Updates(&itemIcon).Error; err != nil {
			return itemIcon, err
		}
	} else {
		if _, err := p.FindByOnlyName(userId, itemIcon.OnlyName); err == nil && itemIcon.OnlyName != "" {
			return itemIcon, ErrItemOnlyNameExisted
		}
		itemIcon.Sort = 9999
		// 创建
		if err := global.Db.Create(&itemIcon).Error; err != nil {
			return itemIcon, err
		}
	}

	return itemIcon, nil

}

func (p *ItemType) FindByOnlyName(userId uint, onlyName string) (res models.ItemIcon, gormErr error) {
	gormErr = global.Db.First(&res, "only_name=? AND user_id=?", onlyName, userId).Error
	if gormErr != nil {
		return
	}
	// Icon 字段带 gorm:"-"，不会从数据库自动填充，真实图标存放在 IconJson。
	// 这里必须解析，否则调用方（尤其 openAPI 的 update/getInfo）拿到零值 Icon，
	// 写回时会用空图标覆盖 IconJson，造成图标永久丢失。内部 API 的列表接口同样做了此解析。
	_ = json.Unmarshal([]byte(res.IconJson), &res.Icon)
	return
}
