package panel

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/initialize/probe"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type NetworkApi struct{}

// GetProbeConfig 读取探针配置（管理员）
func (a *NetworkApi) GetProbeConfig(c *gin.Context) {
	cfg := probe.LoadConfig()
	status := map[string]any{}
	if p, err := probe.NewProvider(cfg); err == nil {
		if st, err := p.Status(); err == nil {
			status = map[string]any{
				"provider":  st.Provider,
				"notAfter":  st.NotAfter,
				"issuer":    st.Issuer,
				"lastError": st.LastError,
			}
		} else {
			status = map[string]any{"lastError": err.Error()}
		}
	}
	apiReturn.SuccessData(c, gin.H{
		"config": cfg,
		"cert":   status,
	})
}

// SaveProbeConfig 保存探针配置（管理员），保存后证书按 mtime 热加载
func (a *NetworkApi) SaveProbeConfig(c *gin.Context) {
	req := probe.Config{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if err := probe.SaveConfig(req); err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	apiReturn.Success(c)
}

// TestProbe 校验当前配置：证书能否加载、到期时间、探针端口
func (a *NetworkApi) TestProbe(c *gin.Context) {
	cfg := probe.LoadConfig()
	provider, err := probe.NewProvider(cfg)
	if err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	certFile, keyFile, err := provider.Ensure()
	if err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	st, _ := provider.Status()
	info, _ := probe.ParseCertInfo(certFile)
	apiReturn.SuccessData(c, gin.H{
		"provider": provider.Name(),
		"certFile": certFile,
		"keyFile":  keyFile,
		"port":     cfg.Port,
		"cert":     st,
		"info":     info,
	})
}
