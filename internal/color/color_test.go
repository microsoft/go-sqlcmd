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
	colorizer := New(true)
	noncolorizer := New(false)

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
		{
			name:  "Header",
			args:  args{s: "header", t: TextTypeHeader},
			wantW: "\x1b[1m\x1b[38;5;4mheader\x1b[0m",
		},
		{
			name:  "Cell",
			args:  args{s: "cell", t: TextTypeCell},
			wantW: "\x1b[38;5;2mcell\x1b[0m",
		},
		{
			name:  "Separator",
			args:  args{s: "sep", t: TextTypeSeparator},
			wantW: "\x1b[38;5;131msep\x1b[0m",
		},
		{
			name:  "Error",
			args:  args{s: "error", t: TextTypeError},
			wantW: "\x1b[38;5;196merror\x1b[0m",
		},
		{
			name:  "Warning",
			args:  args{s: "warn", t: TextTypeWarning},
			wantW: "\x1b[3mwarn\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := colorizer.Write(w, tt.args.s, "emacs", tt.args.t)
			assert.NoErrorf(t, err, "Write returned an error %+v", tt.args)
			gotW := w.String()
			assert.Equalf(t, tt.wantW, gotW, "colorizer.Write(%+v)", tt.args)
			w.Reset()
			noncolorizer.Write(w, tt.args.s, "emacs", tt.args.t)
			assert.Equalf(t, tt.args.s, w.String(), "noncolorizer.Write should write unmodified string")

		})
	}
}
