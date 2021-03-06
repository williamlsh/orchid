package email

import (
	"flag"
	"os"
	"testing"

	"go.uber.org/zap"
)

var conf ConfigOptions

func TestMain(m *testing.M) {
	flag.StringVar(&conf.From, "f", "", "")
	flag.StringVar(&conf.Host, "h", "", "")
	flag.IntVar(&conf.Port, "p", 0, "")
	flag.StringVar(&conf.Username, "u", "", "")
	flag.StringVar(&conf.Passwd, "s", "", "")
	flag.Parse()
	os.Exit(m.Run())
}

func TestSend(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	mail := New(sugar, conf, conf.From, "test")
	if err := mail.Send("123"); err != nil {
		t.Fatal(err)
	}
}
