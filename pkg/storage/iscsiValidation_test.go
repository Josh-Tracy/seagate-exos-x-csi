//
// Copyright (c) 2026 Seagate Technology LLC and/or its Affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//

package storage

import (
	"path/filepath"
	"testing"
)

func TestCanonicalMultipathWWID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "array WWN", input: "600c0ff000f8b0a13eb1be6a01000000", want: "3600c0ff000f8b0a13eb1be6a01000000"},
		{name: "already canonical", input: "3600c0ff000f8b0a13eb1be6a01000000", want: "3600c0ff000f8b0a13eb1be6a01000000"},
		{name: "trim whitespace", input: " 600c0ff000f8b0a13eb1be6a01000000\n", want: "3600c0ff000f8b0a13eb1be6a01000000"},
		{name: "empty", input: "", wantErr: true},
		{name: "path injection", input: "../../dm-0", wantErr: true},
		{name: "whitespace", input: "600c0ff0 bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := canonicalMultipathWWID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("canonicalMultipathWWID(%q) unexpectedly succeeded with %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("canonicalMultipathWWID(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("canonicalMultipathWWID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidatePersistedSinglePathMissingDevice(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-device")
	alreadyGone, err := validatePersistedSinglePath(missing, "3600c0ff000f8b0a13eb1be6a01000000")
	if err != nil {
		t.Fatal(err)
	}
	if !alreadyGone {
		t.Fatal("missing single-path device was not reported as already gone")
	}
}

func TestValidatePersistedSinglePathRejectsEmptyDevice(t *testing.T) {
	alreadyGone, err := validatePersistedSinglePath("", "3600c0ff000f8b0a13eb1be6a01000000")
	if err == nil {
		t.Fatal("empty single-path device unexpectedly passed validation")
	}
	if alreadyGone {
		t.Fatal("empty device path was incorrectly reported as an idempotent missing device")
	}
}

func TestIsRequestedISCSIByPath(t *testing.T) {
	const iqn = "iqn.1988-11.com.dell:01.array.example"
	tests := []struct {
		name string
		path string
		lun  int
		want bool
	}{
		{name: "matching IPv4 portal", path: "ip-10.10.71.2:3260-iscsi-" + iqn + "-lun-13", lun: 13, want: true},
		{name: "matching IPv6-style portal", path: "ip-[fd00::2]:3260-iscsi-" + iqn + "-lun-13", lun: 13, want: true},
		{name: "wrong LUN", path: "ip-10.10.71.2:3260-iscsi-" + iqn + "-lun-12", lun: 13},
		{name: "wrong target", path: "ip-10.10.71.2:3260-iscsi-iqn.other-lun-13", lun: 13},
		{name: "not by path", path: "pci-0000:00:00.0-iscsi-" + iqn + "-lun-13", lun: 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRequestedISCSIByPath(tt.path, iqn, tt.lun); got != tt.want {
				t.Fatalf("isRequestedISCSIByPath(%q, %q, %d) = %v, want %v",
					tt.path, iqn, tt.lun, got, tt.want)
			}
		})
	}
}
