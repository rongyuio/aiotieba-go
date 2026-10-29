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

// PrivChatCursors 与某个用户的私信会话游标。
type PrivChatCursors struct {
	LocalMsgID  int // 本端最后一次发送使用的 local_msg_id，由服务端回显
	PulledMsgID int // 已拉取到的最新 msg_id
	SentMsgID   int // 最近一次上报的已读位置 msg_id
	ReadMsgID   int // 对方已读到的最新 msg_id
}

// MsgIDManager msg_id 管理器，维护各消息组的 msg_id，对应 aiotieba.core.websocket.MsgIDManager。
type MsgIDManager struct {
	PrivGID     int
	GID2MID     map[int]*MsgIDPair
	UID2Cursors map[int]*PrivChatCursors
}

// NewMsgIDManager 创建一个已包含默认私信分组的 msg_id 管理器。
func NewMsgIDManager() *MsgIDManager {
	return &MsgIDManager{
		GID2MID:     map[int]*MsgIDPair{0: {}},
		UID2Cursors: map[int]*PrivChatCursors{},
	}
}

// UpdateMsgID 更新 groupID 对应的 msg_id。
//
// 参数:
//
//	groupID 消息组id
//	msgID 当前消息的msg_id
//
// 未知分组会先登记再更新。旧实现只改一个临时对象并丢弃，导致
// 非私信分组的 msg_id 永远停留在 0。
func (m *MsgIDManager) UpdateMsgID(groupID, msgID int) {
	pair, ok := m.GID2MID[groupID]
	if ok {
		pair.Update(msgID)
		return
	}
	m.GID2MID[groupID] = &MsgIDPair{LastID: msgID, CurrID: msgID}
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

// GetCursors 获取与某个用户的私信会话游标，不存在时创建。
//
// 参数:
//
//	userID 用户的user_id
func (m *MsgIDManager) GetCursors(userID int) *PrivChatCursors {
	cursors, ok := m.UID2Cursors[userID]
	if !ok {
		cursors = &PrivChatCursors{}
		m.UID2Cursors[userID] = cursors
	}
	return cursors
}

// GetLocalMsgID 返回下一条私信的 local_msg_id。
//
// local_msg_id 必须单调递增：以「已拉取到的最新 msg_id × 100」为下界，
// 保证新值永远排在对方已发消息之后。
//
// 参数:
//
//	userID 接收方的user_id
func (m *MsgIDManager) GetLocalMsgID(userID int) int {
	cursors := m.GetCursors(userID)
	base := max(cursors.LocalMsgID, cursors.PulledMsgID*100)
	return base + 1
}

// UpdateLocalMsgID 更新最近一次发送使用的 local_msg_id。
func (m *MsgIDManager) UpdateLocalMsgID(userID, localMsgID int) {
	if cursors := m.GetCursors(userID); localMsgID > cursors.LocalMsgID {
		cursors.LocalMsgID = localMsgID
	}
}

// UpdatePulledMsgID 更新已拉取到的最新 msg_id。
func (m *MsgIDManager) UpdatePulledMsgID(userID, msgID int) {
	if cursors := m.GetCursors(userID); msgID > cursors.PulledMsgID {
		cursors.PulledMsgID = msgID
	}
}

// UpdateSentMsgID 更新最近一次上报的已读位置 msg_id。
func (m *MsgIDManager) UpdateSentMsgID(userID, msgID int) {
	if cursors := m.GetCursors(userID); msgID > cursors.SentMsgID {
		cursors.SentMsgID = msgID
	}
}

// UpdateReadMsgID 更新对方已读到的最新 msg_id。
func (m *MsgIDManager) UpdateReadMsgID(userID, msgID int) {
	if cursors := m.GetCursors(userID); msgID > cursors.ReadMsgID {
		cursors.ReadMsgID = msgID
	}
}
