package browser

import (
	"reflect"
	"testing"
)

func TestCommandFor(t *testing.T) {
	const url = "http://127.0.0.1:10987"

	cases := []struct {
		goos     string
		wantName string
		wantArgs []string
		wantErr  bool
	}{
		{
			goos:     "windows",
			wantName: "rundll32",
			wantArgs: []string{"url.dll,FileProtocolHandler", url},
		},
		{
			goos:     "linux",
			wantName: "xdg-open",
			wantArgs: []string{url},
		},
		{
			goos:     "darwin",
			wantName: "open",
			wantArgs: []string{url},
		},
		{
			goos:    "plan9",
			wantErr: true,
		},
	}

	for _, c := range cases {
		name, args, err := commandFor(c.goos, url)
		if c.wantErr {
			if err == nil {
				t.Errorf("%s: expected an error, got %q %v", c.goos, name, args)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", c.goos, err)
			continue
		}
		if name != c.wantName {
			t.Errorf("%s: command = %q, want %q", c.goos, name, c.wantName)
		}
		if !reflect.DeepEqual(args, c.wantArgs) {
			t.Errorf("%s: args = %v, want %v", c.goos, args, c.wantArgs)
		}
	}
}
