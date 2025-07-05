package routes

import (
	"github.com/gin-gonic/gin"
)

// RouteModule 路由模块接口，类似于Flask的Blueprint
type RouteModule interface {
	RegisterRoutes(r *gin.Engine)
	GetPrefix() string
	GetDescription() string
}

// RouteManager 路由管理器
type RouteManager struct {
	modules []RouteModule
}

// NewRouteManager 创建新的路由管理器
func NewRouteManager() *RouteManager {
	return &RouteManager{
		modules: make([]RouteModule, 0),
	}
}

// RegisterModule 注册路由模块，类似于Flask的register_blueprint
func (rm *RouteManager) RegisterModule(module RouteModule) {
	rm.modules = append(rm.modules, module)
}

// InitializeRoutes 初始化所有注册的路由模块
func (rm *RouteManager) InitializeRoutes(r *gin.Engine) {
	for _, module := range rm.modules {
		module.RegisterRoutes(r)
	}
}

// GetRegisteredModules 获取所有已注册的模块信息
func (rm *RouteManager) GetRegisteredModules() []map[string]string {
	modules := make([]map[string]string, 0, len(rm.modules))
	for _, module := range rm.modules {
		modules = append(modules, map[string]string{
			"prefix":      module.GetPrefix(),
			"description": module.GetDescription(),
		})
	}
	return modules
}
