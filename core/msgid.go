package core

// MsgIDPair 长度为2的 msg_id 队列，记录新旧 msg_id，对应 aiotieba.core.websocket.MsgIDPair。
type MsgIDPair struct {
	LastID int
	CurrID int
}

// Update 更新 msg_id，把当前 msg_id 移入 LastID。
//
// 参数:
//
//	currID 当前消息的msg_id
func (p *MsgIDPair) Update(currID int) {
	p.LastID = p.CurrID
	p.CurrID = currID
}

// MsgIDManager msg_id 管理器，维护各消息组的 msg_id，对应 aiotieba.core.websocket.MsgIDManager。
type MsgIDManager struct {
	PrivGID int
	GID2MID map[int]*MsgIDPair
}

// NewMsgIDManager 创建一个已包含默认私信分组的 msg_id 管理器。
func NewMsgIDManager() *MsgIDManager {
	return &MsgIDManager{GID2MID: map[int]*MsgIDPair{0: {}}}
}

// UpdateMsgID 更新 groupID 对应的 msg_id。
//
// 参数:
//
//	groupID 消息组id
//	msgID 当前消息的msg_id
//
// 与 Python 原版一致，未知分组只更新一个临时对象，不会被保存。
func (m *MsgIDManager) UpdateMsgID(groupID, msgID int) {
	pair, ok := m.GID2MID[groupID]
	if ok {
		pair.Update(msgID)
		return
	}
	pair = &MsgIDPair{LastID: msgID, CurrID: msgID}
	_ = pair
}

// GetMsgID 获取 groupID 对应的 msg_id，即上一条消息的 msg_id；未知分组返回 0。
//
// 参数:
//
//	groupID 消息组id
func (m *MsgIDManager) GetMsgID(groupID int) int {
	if pair, ok := m.GID2MID[groupID]; ok {
		return pair.LastID
	}
	return 0
}

// GetRecordID 获取 record_id。
//
// 返回私信分组的 record_id。
func (m *MsgIDManager) GetRecordID() int {
	return m.GetMsgID(m.PrivGID)*100 + 1
}
