package models

type (
	PropertyType  = string
	PropertyStyle = string // UI Representation Style
)

const (
	PropertyTypeInt    PropertyType = "int"
	PropertyTypeUint   PropertyType = "uint"
	PropertyTypeFloat  PropertyType = "float"
	PropertyTypeBool   PropertyType = "bool"
	PropertyTypeString PropertyType = "string"
)

var (
	DeviceDataPropertyTypes = map[PropertyType]struct{}{
		PropertyTypeInt:    {},
		PropertyTypeUint:   {},
		PropertyTypeFloat:  {},
		PropertyTypeBool:   {},
		PropertyTypeString: {},
	}
)

type Property struct {
	Name         string        `json:"name"`         // Name 为属性的展示名称
	Desc         string        `json:"desc"`         // Desc 为属性的描述, 通常以名称旁的?形式进行展示
	Type         PropertyType  `json:"type"`         // Type 为该属性的数据类型
	Style        PropertyStyle `json:"style"`        // Style 为该属性在前端的展示样式
	Default      string        `json:"default"`      // Default 该属性默认的属性值
	Range        string        `json:"range"`        // Range 为属性值的可选范围
	Precondition string        `json:"precondition"` // Precondition 为当前属性展示的前置条件, 用来实现简单的动态依赖功能
	Required     bool          `json:"required"`     // Required 表示该属性是否为必填项
	Multiple     bool          `json:"multiple"`     // Multiple 表示是否支持多选(下拉框), 列表(输入), Map(K,V)
	MaxLen       int64         `json:"max_len"`      // MaxLen 表示当Multiple为true时, 可选择的最大数量
}
