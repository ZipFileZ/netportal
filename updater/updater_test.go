/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package updater

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"golang.org/x/crypto/blake2b"

	"github.com/amnezia-vpn/amneziawg-windows-client/version"
)

// signLikeNpRelease mirrors cmd/np-release so the two stay compatible.
func signLikeNpRelease(t *testing.T, message string) (sigFile, publicKey string) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyNum := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	sig := ed25519.Sign(priv, []byte(message))
	sigFile = "untrusted comment: test\n" + base64.StdEncoding.EncodeToString(append(append([]byte("Ed"), keyNum...), sig...)) + "\n" + message
	publicKey = base64.StdEncoding.EncodeToString(append(append([]byte("Ed"), keyNum...), pub...))
	return
}

func TestSignedListAndCandidate(t *testing.T) {
	h := blake2b.Sum256([]byte("exe"))
	name := fmt.Sprintf("netportal-%s-99.0.0.exe", version.Arch())
	message := hex.EncodeToString(h[:]) + "  " + name + "\n" +
		hex.EncodeToString(h[:]) + "  netportal-other-99.0.0.exe\n"
	sigFile, pub := signLikeNpRelease(t, message)

	files, err := readFileListWithKey([]byte(sigFile), pub)
	if err != nil {
		t.Fatal(err)
	}
	found, err := findCandidate(files)
	if err != nil || found == nil {
		t.Fatalf("expected a candidate, got %v, %v", found, err)
	}
	if found.Version() != "99.0.0" || found.name != name || found.hash != h {
		t.Fatalf("wrong candidate: %+v", found)
	}

	tampered := []byte(sigFile[:len(sigFile)-2] + "x\n")
	if _, err := readFileListWithKey(tampered, pub); err == nil {
		t.Fatal("tampered list was accepted")
	}
	_, otherPub := signLikeNpRelease(t, message)
	if _, err := readFileListWithKey([]byte(sigFile), otherPub); err == nil {
		t.Fatal("list signed by another key was accepted")
	}
}

func TestOlderReleaseIgnored(t *testing.T) {
	h := blake2b.Sum256([]byte("exe"))
	message := hex.EncodeToString(h[:]) + "  " + fmt.Sprintf("netportal-%s-0.0.1.exe", version.Arch()) + "\n"
	sigFile, pub := signLikeNpRelease(t, message)
	files, err := readFileListWithKey([]byte(sigFile), pub)
	if err != nil {
		t.Fatal(err)
	}
	if found, err := findCandidate(files); err != nil || found != nil {
		t.Fatalf("older release should be ignored, got %v, %v", found, err)
	}
}
