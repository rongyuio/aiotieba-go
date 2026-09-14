// Package classdef 存放各 API 响应共用的数据模型。
//
// 对应 Python 包 aiotieba.api._classdef。
package classdef

// Containers 内容列表的泛型基类，约定取内容的通用接口，对应 aiotieba.api._classdef.container.Containers。
type Containers[T any] struct {
	Objs []T // 内容列表
}

// Len 返回内容数量，对应 __len__。
func (c Containers[T]) Len() int { return len(c.Objs) }

// Empty 报告内容列表是否为空，对应 __bool__。
func (c Containers[T]) Empty() bool { return len(c.Objs) == 0 }
