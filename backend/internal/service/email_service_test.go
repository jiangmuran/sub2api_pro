package service

import "testing"

func TestNormalizeSMTPHost(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{name: "plain host", in: "smtp.example.com", out: "smtp.example.com"},
		{name: "host with port", in: "smtp.example.com:587", out: "smtp.example.com"},
		{name: "trim spaces", in: "  smtp.example.com  ", out: "smtp.example.com"},
		{name: "empty", in: "", out: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeSMTPHost(tc.in); got != tc.out {
				t.Fatalf("normalizeSMTPHost(%q)=%q, want %q", tc.in, got, tc.out)
			}
		})
	}
}

func TestLoginAuth(t *testing.T) {
	a := LoginAuth("user@example.com", "secret")
	la, ok := a.(*loginAuth)
	if !ok {
		t.Fatalf("expected *loginAuth")
	}

	proto, ir, err := la.Start(nil)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}
	if proto != "LOGIN" {
		t.Fatalf("unexpected auth proto: %s", proto)
	}
	if string(ir) != "user@example.com" {
		t.Fatalf("unexpected initial response: %s", string(ir))
	}

	if next, err := la.Next([]byte("Username:"), true); err != nil || string(next) != "user@example.com" {
		t.Fatalf("username challenge failed, next=%q err=%v", string(next), err)
	}
	if next, err := la.Next([]byte("Password:"), true); err != nil || string(next) != "secret" {
		t.Fatalf("password challenge failed, next=%q err=%v", string(next), err)
	}
	if next, err := la.Next(nil, false); err != nil || next != nil {
		t.Fatalf("terminal challenge failed, next=%v err=%v", next, err)
	}
}
