package models

type Service struct {
	GormModel
	Type    string  `json:"type"`
	Name    string  `json:"name" gorm:"uniqueIndex;type:text collate nocase"`
	URL     string  `json:"url" gorm:"uniqueIndex;type:text collate nocase"`
	APIKey  string  `json:"api_key" gorm:"uniqueIndex;type:text collate nocase"`
	Proxies []Proxy `json:"proxies" gorm:"foreignKey:ServiceID"`
}

type App struct {
	GormModel
	Template string  `json:"template"`
	Name     string  `json:"name" gorm:"uniqueIndex;type:text collate nocase"`
	IsActive *bool   `json:"is_active"`
	Proxies  []Proxy `json:"proxies" gorm:"foreignKey:AppID"`
}

type Proxy struct {
	GormModel
	APIKey    string  `json:"api_key" gorm:"uniqueIndex;type:text collate nocase"`
	AppID     uint    `json:"app_id" gorm:"index:idx_proxy_id,unique;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	App       App     `json:"app"`
	ServiceID uint    `json:"service_id" gorm:"index:idx_proxy_id,unique;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Service   Service `json:"service"`
}

type Notification struct {
	GormModel
	URL string `json:"url" gorm:"uniqueIndex;type:text collate nocase"`
}

type Setting struct {
	GormModel
	Key       string `json:"key" gorm:"uniqueIndex;type:text collate nocase"`
	Value     string `json:"value"`
	IsDefault bool   `json:"is_default" gorm:"default:true"`
}
