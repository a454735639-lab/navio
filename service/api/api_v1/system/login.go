package system

import (
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiData/panelApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/api/api_v1/middleware"
	"sun-panel/biz"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/lib/password"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginApi struct {
}

// 登录输入验证
type LoginLoginVerify struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,max=50"`
	VCode    string `json:"vcode" validate:"max=6"`
	Email    string `json:"email"`
}

// @Summary 登录账号
// @Accept application/json
// @Produce application/json
// @Param LoginLoginVerify body LoginLoginVerify true "登陆验证信息"
// @Tags user
// @Router /login [post]
func (l LoginApi) Login(c *gin.Context) {
	param := LoginLoginVerify{}
	if err := c.ShouldBindJSON(&param); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if errMsg, err := base.ValidateInputStruct(param); err != nil {
		apiReturn.ErrorParamFomat(c, errMsg)
		return
	}

	settings := panelApiStructs.GlobalSettingSite{}
	biz.GlobalSetting.GetSiteStting(&settings)
	// 验证验证码
	if settings.GlobalSettingSiteSetting.IsEnableLoginCaptcha {
		if err := biz.Captcha.CaptchaVerifyHandle(c, param.VCode, true); err != nil {
			global.Logger.Errorln("Login error:", err.Error())
			apiReturn.ErrorByCode(c, apiReturn.ErrCodeCaptchaError)
			return
		}
	}

	mUser := models.User{}
	var (
		err  error
		info models.User
	)
	bToken := ""
	param.Username = strings.TrimSpace(param.Username)

	// 账号维度限流：IP 可被 XFF 伪造绕过，但爆破目标是固定账号，
	// 按账号计数即使换 IP 也无法绕过（零宽限：先检查再验证）。
	accountKey := middleware.AccountKey(param.Username)
	if middleware.IsBlocked(accountKey) {
		apiReturn.ErrorCode(c, 1006, "too many failed attempts, try again later", nil)
		return
	}

	if info, err = mUser.GetUserInfoByUsername(param.Username); err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.RecordLoginFailure(c.ClientIP())
			middleware.RecordLoginFailure(accountKey)
			apiReturn.ErrorByCode(c, 1003)
			return
		}
		apiReturn.Error(c, err.Error())
		return
	}

	ok, needUpgrade := password.Verify(info.Password, param.Password)
	if !ok {
		middleware.RecordLoginFailure(c.ClientIP())
		middleware.RecordLoginFailure(accountKey)
		apiReturn.ErrorByCode(c, 1003)
		return
	}
	if needUpgrade {
		if newHash, err := password.Hash(param.Password); err == nil {
			_ = mUser.UpdateUserInfoByUserId(info.ID, map[string]interface{}{"password": newHash})
			info.Password = newHash
		}
	}

	// 停用或未激活
	if info.Status != 1 {
		apiReturn.ErrorByCode(c, 1004)
		return
	}
	middleware.ResetLoginFailure(c.ClientIP())
	middleware.ResetLoginFailure(accountKey)

	bToken = info.Token
	if info.Token == "" {
		// 生成token
		buildTokenOver := false
		for !buildTokenOver {
			bToken = cmn.BuildRandCode(32, cmn.RAND_CODE_MODE2)
			if _, err := mUser.GetUserInfoByToken(bToken); err != nil {
				// 保存token
				mUser.UpdateUserInfoByUserId(info.ID, map[string]interface{}{
					"token": bToken,
				})
				buildTokenOver = true
			}
		}
		info.Token = bToken
	}
	info.Password = ""
	info.ReferralCode = ""

	// global.UserToken.SetDefault(bToken, info)
	cToken := uuid.NewString() + "-" + cmn.Md5(cmn.Md5("userId"+strconv.Itoa(int(info.ID))))
	global.CUserToken.SetDefault(cToken, bToken)
	global.Logger.Debug("token:", cmn.MaskToken(cToken))
	// token 完整值不再打印

	// 设置当前用户信息
	c.Set("userInfo", info)
	info.Token = cToken // 重要 采用cToken,隐藏真实token
	apiReturn.SuccessData(c, info)
}

// 安全退出
func (l *LoginApi) Logout(c *gin.Context) {
	// userInfo, _ := base.GetCurrentUserInfo(c)
	cToken := c.GetHeader("token")
	global.CUserToken.Delete(cToken)
	apiReturn.Success(c)
}
