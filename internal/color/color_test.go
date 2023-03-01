package color

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	type args struct {
		s string
		t TextType
	}
	colorizer := New("emacs")
	noncolorizer := New("")

	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{
			name:  "Normal",
			args:  args{s: "select 100", t: TextTypeNormal},
			wantW: "select 100",
		},
		{
			name:  "TSQL",
			args:  args{s: "select top (1) name from sys.tables", t: TextTypeTSql},
			wantW: "\x1b[1m\x1b[38;5;129mselect\x1b[0m\x1b[38;5;250m \x1b[0m\x1b[1m\x1b[38;5;129mtop\x1b[0m\x1b[38;5;250m \x1b[0m(\x1b[38;5;241m1\x1b[0m)\x1b[38;5;250m \x1b[0mname\x1b[38;5;250m \x1b[0m\x1b[1m\x1b[38;5;129mfrom\x1b[0m\x1b[38;5;250m \x1b[0msys.tables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			colorizer.Write(w, tt.args.s, tt.args.t)
			gotW := w.String()
			assert.Equalf(t, tt.wantW, gotW, "colorizer.Write(%+v)", tt.args)
			w.Reset()
			noncolorizer.Write(w, tt.args.s, tt.args.t)
			assert.Equalf(t, tt.args.s, w.String(), "noncolorizer.Write should write unmodified string")

		})
	}
}
