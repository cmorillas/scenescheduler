// backend/webserver/internal/sourcepreview/session.go
//
// Session tracking helper functions.
//
// Contents:
// - addSession - Add session to tracking maps
// - removeSession - Remove session from tracking maps

package sourcepreview

// addSession adds a new session to tracking maps.
// Must be called with appropriate locking from caller.
func (m *Manager) addSession(session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activePreviews[session.PreviewID] = session
	m.connIDToPreview[session.ConnectionID] = session.PreviewID
}

// removeSession removes a session from tracking maps by PreviewID.
// Used during error handling in processPreview goroutine.
func (m *Manager) removeSession(previewID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.activePreviews[previewID]; ok {
		delete(m.activePreviews, previewID)
		if session.ConnectionID != "" {
			delete(m.connIDToPreview, session.ConnectionID)
		}
	}
}
