package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func Test_NewOptions(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions()
	asrt.Equal(DefaultPrefix, opt.Prefix)
	asrt.Equal(DefaultDirectory, opt.Directory)
	asrt.Equal(DefaultLevel.String(), opt.Level)
	asrt.Equal(DefaultTimeLayout, opt.TimeLayout)
	asrt.Equal(DefaultFormat, opt.Format)
	asrt.Equal(DefaultDisableCaller, opt.DisableCaller)
	asrt.Equal(DefaultDisableStacktrace, opt.DisableStacktrace)
	asrt.Equal(DefaultDisableSplitError, opt.DisableSplitError)
	asrt.Equal(DefaultMaxSize, opt.MaxSize)
	asrt.Equal(DefaultMaxBackups, opt.MaxBackups)
	asrt.Equal(DefaultCompress, opt.Compress)
}

func Test_Options_WithPrefix(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithPrefix("test_")
	asrt.Equal("test_", opt.Prefix)

	opt = NewOptions().WithPrefix("")
	asrt.Equal("", opt.Prefix)
}

func Test_Options_WithDirectory(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDirectory("test_dir")
	asrt.Equal("test_dir", opt.Directory)

	opt = NewOptions().WithDirectory("")
	asrt.Equal(DefaultDirectory, opt.Directory)
}

func Test_Options_WithLevel(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithLevel(zapcore.DebugLevel.String())
	asrt.Equal(zapcore.DebugLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.InfoLevel.String())
	asrt.Equal(zapcore.InfoLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.WarnLevel.String())
	asrt.Equal(zapcore.WarnLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.ErrorLevel.String())
	asrt.Equal(zapcore.ErrorLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.DPanicLevel.String())
	asrt.Equal(zapcore.DPanicLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.PanicLevel.String())
	asrt.Equal(zapcore.PanicLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.FatalLevel.String())
	asrt.Equal(zapcore.FatalLevel.String(), opt.Level)

	opt = NewOptions().WithLevel("")
	asrt.Equal(DefaultLevel.String(), opt.Level)
}

func Test_Options_WithTimeLayout(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithTimeLayout("2006-01-02 15:04:05.000")
	asrt.Equal("2006-01-02 15:04:05.000", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01-02 15:04:05")
	asrt.Equal("2006-01-02 15:04:05", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01-02 15:04")
	asrt.Equal("2006-01-02 15:04", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01-02")
	asrt.Equal("2006-01-02", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01")
	asrt.Equal("2006-01", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006")
	asrt.Equal("2006", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02 15:04:05.000")
	asrt.Equal("2006/01/02 15:04:05.000", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02 15:04:05")
	asrt.Equal("2006/01/02 15:04:05", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02 15:04")
	asrt.Equal("2006/01/02 15:04", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02")
	asrt.Equal("2006/01/02", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01")
	asrt.Equal("2006/01", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006")
	asrt.Equal("2006", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("")
	asrt.Equal(DefaultTimeLayout, opt.TimeLayout)
}

func Test_Options_WithFormat(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithFormat("console")
	asrt.Equal("console", opt.Format)

	opt = NewOptions().WithFormat("json")
	asrt.Equal("json", opt.Format)

	opt = NewOptions().WithFormat("")
	asrt.Equal(DefaultFormat, opt.Format)
}

func Test_Options_WithDisableCaller(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDisableCaller(true)
	asrt.True(opt.DisableCaller)

	opt = NewOptions().WithDisableCaller(false)
	asrt.False(opt.DisableCaller)
}

func Test_Options_WithDisableStacktrace(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDisableStacktrace(true)
	asrt.True(opt.DisableStacktrace)

	opt = NewOptions().WithDisableStacktrace(false)
	asrt.False(opt.DisableStacktrace)
}

func Test_Options_WithDisableSplitError(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDisableSplitError(true)
	asrt.True(opt.DisableSplitError)

	opt = NewOptions().WithDisableSplitError(false)
	asrt.False(opt.DisableSplitError)
}

func Test_Options_WithMaxSize(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithMaxSize(100)
	asrt.Equal(100, opt.MaxSize)

	opt = NewOptions().WithMaxSize(1000)
	asrt.Equal(1000, opt.MaxSize)

	opt = NewOptions().WithMaxSize(10000)
	asrt.Equal(10000, opt.MaxSize)

	opt = NewOptions().WithMaxSize(-1)
	asrt.Equal(DefaultMaxSize, opt.MaxSize)

	opt = NewOptions().WithMaxSize(0)
	asrt.Equal(DefaultMaxSize, opt.MaxSize)
}

func Test_Options_WithMaxBackups(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithMaxBackups(100)
	asrt.Equal(100, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(1000)
	asrt.Equal(1000, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(10000)
	asrt.Equal(10000, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(-1)
	asrt.Equal(DefaultMaxBackups, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(0)
	asrt.Equal(DefaultMaxBackups, opt.MaxBackups)
}

func Test_Options_WithCompress(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithCompress(true)
	asrt.True(opt.Compress)

	opt = NewOptions().WithCompress(false)
	asrt.False(opt.Compress)
}
