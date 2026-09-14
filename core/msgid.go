package core

// MsgIDPair keeps the last two msg ids of a message group. It mirrors
// aiotieba.core.websocket.MsgIDPair.
type MsgIDPair struct {
	LastID int
	CurrID int
}

// Update shifts the current msg id into LastID.
func (p *MsgIDPair) Update(currID int) {
	p.LastID = p.CurrID
	p.CurrID = currID
}

// MsgIDManager tracks the msg ids of every message group. It mirrors
// aiotieba.core.websocket.MsgIDManager.
type MsgIDManager struct {
	PrivGID int
	GID2MID map[int]*MsgIDPair
}

// NewMsgIDManager creates a manager seeded with the default private group.
func NewMsgIDManager() *MsgIDManager {
	return &MsgIDManager{GID2MID: map[int]*MsgIDPair{0: {}}}
}

// UpdateMsgID records msgID for the given message group.
//
// Like the Python original, an unknown group id only updates a throw-away pair
// and is not stored.
func (m *MsgIDManager) UpdateMsgID(groupID, msgID int) {
	pair, ok := m.GID2MID[groupID]
	if ok {
		pair.Update(msgID)
		return
	}
	pair = &MsgIDPair{LastID: msgID, CurrID: msgID}
	_ = pair
}

// GetMsgID returns the previous msg id of the group, or 0 when unknown.
func (m *MsgIDManager) GetMsgID(groupID int) int {
	if pair, ok := m.GID2MID[groupID]; ok {
		return pair.LastID
	}
	return 0
}

// GetRecordID returns the record id of the private group.
func (m *MsgIDManager) GetRecordID() int {
	return m.GetMsgID(m.PrivGID)*100 + 1
}
