package aiotieba

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/rongyuio/aiotieba-go/exception"
	"github.com/rongyuio/aiotieba-go/logging"
)

// TestAPICallLogBody 锁定与 Python 版 aiotieba 逐字一致的日志正文。
//
// 只比较正文而不比较前缀，因为前缀（时间戳与级别名）由 zerolog 的
// ConsoleWriter 决定，不要求与 Python 的 formatter 一致。
func TestAPICallLogBody(t *testing.T) {
	c, err := New("", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	var buf bytes.Buffer
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        &buf,
		NoColor:    true,
		TimeFormat: time.DateTime,
	}).With().Timestamp().Logger()
	logging.SetLogger(&logger)
	t.Cleanup(func() { logging.SetLogger(nil) })

	// 失败：正文是异常本身，与 Python 的 str(TiebaServerError(340011, '')) 一致。
	c.logCallError("sign_forums", &exception.TiebaServerError{Code: 340011})
	assertLogBody(t, buf.String(), "[sign_forums] (340011, ''). args=() kwargs={}")

	// 成功：位置参数带尾逗号，与示例日志一致。
	buf.Reset()
	c.logCallSuccess("sign_forum", "盗墓笔记")
	assertLogBody(t, buf.String(), "[sign_forum] Succeeded. args=('盗墓笔记',) kwargs={}")

	// 关键字参数：未被 logging.PyKw 标记的值一律按位置参数渲染。
	buf.Reset()
	c.logCallError("get_posts", &exception.TiebaServerError{Code: 4, Msg: "x"},
		int64(123), logging.PyKw{Name: "pn", Value: 2})
	assertLogBody(t, buf.String(), "[get_posts] (4, 'x'). args=(123,) kwargs={'pn': 2}")
}

func assertLogBody(t *testing.T, got, want string) {
	t.Helper()
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), want) {
		t.Errorf("log body mismatch\n got: %q\nwant suffix: %q", got, want)
	}
	if !strings.Contains(got, time.Now().Format("2006-01-02")) {
		t.Errorf("log line should carry a local timestamp, got %q", got)
	}
}
