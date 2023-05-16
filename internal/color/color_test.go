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
			wantW: "\x1b[1m\x1b[38;2;170;34;255mselect\x1b[0m\x1b[38;2;187;187;187m \x1b[0m\x1b[1m\x1b[38;2;170;34;255mtop\x1b[0m\x1b[38;2;187;187;187m \x1b[0m(\x1b[38;2;102;102;102m1\x1b[0m)\x1b[38;2;187;187;187m \x1b[0mname\x1b[38;2;187;187;187m \x1b[0m\x1b[1m\x1b[38;2;170;34;255mfrom\x1b[0m\x1b[38;2;187;187;187m \x1b[0msys.tables",
		},
		{
			name:  "Header",
			args:  args{s: "header", t: TextTypeHeader},
			wantW: "\x1b[1m\x1b[38;2;0;0;128mheader\x1b[0m",
		},
		{
			name:  "Cell",
			args:  args{s: "cell", t: TextTypeCell},
			wantW: "\x1b[38;2;0;128;0mcell\x1b[0m",
		},
		{
			name:  "Separator",
			args:  args{s: "sep", t: TextTypeSeparator},
			wantW: "\x1b[38;2;187;68;68msep\x1b[0m",
		},
		{
			name:  "Error",
			args:  args{s: "error", t: TextTypeError},
			wantW: "\x1b[38;2;255;0;0merror\x1b[0m",
		},
		{
			name:  "Warning",
			args:  args{s: "warn", t: TextTypeWarning},
			wantW: "\x1b[3mwarn\x1b[0m",
		},
		{
			name:  "XML",
			args:  args{s: "<node>value</node>", t: TextTypeXml},
			wantW: "\x1b[1m\x1b[38;2;0;128;0m<node>\x1b[0mvalue\x1b[1m\x1b[38;2;0;128;0m</node>\x1b[0m",
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
			err = noncolorizer.Write(w, tt.args.s, "emacs", tt.args.t)
			assert.NoErrorf(t, err, "nonColorizer.Write returned an error %+v", tt.args)
			assert.Equalf(t, tt.args.s, w.String(), "noncolorizer.Write should write unmodified string")

		})
	}
}

func TestStyles(t *testing.T) {
	c := New(false)
	s := c.Styles()
	assert.Contains(t, s, "emacs", "emacs style not found")
}
