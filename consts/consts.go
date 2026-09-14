// Package consts holds version strings and base hosts shared by the Tieba APIs.
//
// It mirrors the Python module aiotieba.const.
package consts

const (
	// Version is the library version, mirroring aiotieba.__version__.
	Version = "1.0.0"

	// LatestVersion is used by most APIs.
	LatestVersion = "22.6.5.1"
	// LegacyVersion is used by the APIs that depend on the legacy response
	// format (for example get_threads).
	LegacyVersion = "12.64.1.1"
	// ChatVersion is used by the chat / BLCP protocol.
	ChatVersion = "12.68.1.0"

	// AppBaseHost is the host of the app HTTP API.
	AppBaseHost = "tiebac.baidu.com"
	// WebBaseHost is the host of the web HTTP API.
	WebBaseHost = "tieba.baidu.com"
)
